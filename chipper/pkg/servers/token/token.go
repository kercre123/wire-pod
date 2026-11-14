package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gopkg.in/ini.v1"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

var (
	TimeFormat     = time.RFC3339Nano
	ExpirationTime = time.Hour * 24
	UserId         = "wirepod"
)

const maxAppTokens = 5

var (
	jwtKeyOnce sync.Once
	jwtKey     *rsa.PrivateKey
	jwtKeyErr  error
)

func sessionIsValid(session string) bool {
	return true
}

func userIDForSession(session string) string {
	return UserId
}

func userOwnsBot(thing, userID string) bool {
	return true
}

func Thingifier(esn string) string {
	esn = strings.ToLower(strings.TrimSpace(esn))
	if strings.HasPrefix(esn, "vic:") {
		return esn
	}
	return "vic:" + esn
}

func peerIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}
	return vars.HostOnly(p.Addr.String())
}

func PeerCertThing(ctx context.Context) (string, bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return "", false
	}
	return Thingifier(tlsInfo.State.PeerCertificates[0].Subject.CommonName), true
}

func thingFromPeer(ctx context.Context) (string, error) {
	if thing, ok := PeerCertThing(ctx); ok {
		if _, err := vars.ESNFromThing(thing); err != nil {
			return "", err
		}
		return thing, nil
	}
	ip := peerIP(ctx)
	for _, bot := range vars.GetBotInfo().Robots {
		if ip != "" && bot.IPAddress == ip {
			return Thingifier(bot.Esn), nil
		}
	}
	return "", errors.New("no client certificate and bot isn't known by IP")
}

func metadataValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func userIDFromJWT(tokenString string) string {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return ""
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return ""
	}
	userID, _ := payload["user_id"].(string)
	return userID
}

func AddAppToken(thing string) (string, error) {
	guid, tokenHash, err := CreateTokenAndHashedToken()
	if err != nil {
		return "", err
	}
	vars.UpdateJdoc(thing, "vic.AppTokens", func(jdoc *vars.AJdoc, exists bool) {
		if !exists {
			jdoc.FmtVersion = 1
			jdoc.ClientMetadata = "wirepod-new-token"
		}
		var tokens ClientTokenManager
		json.Unmarshal([]byte(jdoc.JsonDoc), &tokens)
		tokens.ClientTokens = append(tokens.ClientTokens, ClientToken{
			Hash:       tokenHash,
			ClientName: "wirepod",
			AppId:      "SDK",
			IssuedAt:   time.Now().Format(TimeFormat),
		})
		if len(tokens.ClientTokens) > maxAppTokens {
			tokens.ClientTokens = tokens.ClientTokens[len(tokens.ClientTokens)-maxAppTokens:]
		}
		jdocJson, _ := json.Marshal(tokens)
		jdoc.JsonDoc = string(jdocJson)
		jdoc.DocVersion++
	})
	return guid, nil
}

func signedJWT(userID, thing string) (string, error) {
	jwtKeyOnce.Do(func() {
		jwtKey, jwtKeyErr = rsa.GenerateKey(rand.Reader, 2048)
	})
	if jwtKeyErr != nil {
		return "", jwtKeyErr
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      time.Now().AddDate(0, 1, 0).Format(TimeFormat),
		"iat":          time.Now().Format(TimeFormat),
		"permissions":  nil,
		"requestor_id": thing,
		"token_id":     uuid.New().String(),
		"token_type":   "user+robot",
		"user_id":      userID,
	})
	return token.SignedString(jwtKey)
}

func issueToken(ctx context.Context, userID, thing string) (*tokenpb.TokenBundle, error) {
	esn := strings.TrimPrefix(thing, "vic:")
	guid, err := AddAppToken(thing)
	if err != nil {
		return nil, err
	}
	if err := vars.RecordBot(esn, peerIP(ctx), guid); err != nil {
		logger.Println("Error writing bot info: ", err)
	}
	jwtToken, err := signedJWT(userID, thing)
	if err != nil {
		return nil, err
	}
	return &tokenpb.TokenBundle{
		Token:       jwtToken,
		ClientToken: guid,
	}, nil
}

func robotNameFromSessionCert(sessionCert []byte) (string, error) {
	block, _ := pem.Decode(sessionCert)
	if block == nil {
		return "", errors.New("session certificate isn't valid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", errors.New("invalid session certificate: " + err.Error())
	}
	if !vars.ValidRobotName(cert.Issuer.CommonName) {
		return "", errors.New("invalid robot name in session certificate")
	}
	return cert.Issuer.CommonName, nil
}

func saveSessionCert(name, esn string, sessionCert []byte) {
	if err := vars.WriteFileAtomic(filepath.Join(vars.SessionCertPath, esn), sessionCert, 0644); err != nil {
		logger.Println("Error writing session cert: ", err)
	}
	if err := ensureDir(vars.SDKIniPath); err != nil {
		logger.Println(err)
		return
	}
	fullPath := filepath.Join(vars.SDKIniPath, name+"-"+esn+".cert")
	logger.Println("Outputting session cert to " + fullPath)
	if err := vars.WriteFileAtomic(fullPath, sessionCert, 0644); err != nil {
		logger.Println("Error writing session cert: ", err)
	}
}

func ChangeGUIDInIni(esn string) {
	userIniData, err := ini.Load(vars.SDKIniPath + "sdk_config.ini")
	if err != nil {
		logger.Println(err)
		return
	}
	botInfo := vars.GetBotInfo()
	for _, robot := range botInfo.Robots {
		if !strings.EqualFold(robot.Esn, esn) {
			continue
		}
		matched := false
		for _, section := range userIniData.Sections() {
			if strings.EqualFold(section.Name(), esn) {
				matched = true
				section.Key("ip").SetValue(robot.IPAddress)
				if robot.GUID == "" {
					section.Key("guid").SetValue(botInfo.GlobalGUID)
				} else {
					section.Key("guid").SetValue(robot.GUID)
				}
			}
		}
		if !matched {
			logger.Println("Bot is not in sdk_config.ini. Clear your bot's userdata and try authenticating again to create it.")
		}
	}
	userIniData.SaveTo(vars.SDKIniPath + "sdk_config.ini")
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	logger.Println("Token: Incoming Associate Primary User request")
	thing, err := thingFromPeer(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	session := metadataValue(ctx, "anki-user-session")
	if !sessionIsValid(session) {
		return nil, status.Error(codes.Unauthenticated, "session_expired")
	}
	name, err := robotNameFromSessionCert(req.SessionCertificate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	esn := strings.TrimPrefix(thing, "vic:")
	ip := peerIP(ctx)
	logger.LogUI("New bot being associated with wire-pod. ESN: " + esn + ", IP: " + ip)

	vars.DeleteData(thing)
	bundle, err := issueToken(ctx, userIDForSession(session), thing)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	saveSessionCert(name, esn, req.SessionCertificate)
	WriteToIniPrimary(name, esn, bundle.ClientToken, ip)
	vars.AddToRInfo(esn, name, ip)
	return &tokenpb.AssociatePrimaryUserResponse{Data: bundle}, nil
}

func (s *TokenServer) refresh(ctx context.Context, checkSession bool, session string) (*tokenpb.TokenBundle, error) {
	thing, err := thingFromPeer(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if checkSession && !sessionIsValid(session) {
		return nil, status.Error(codes.Unauthenticated, "session_expired")
	}
	userID := userIDFromJWT(metadataValue(ctx, "anki-access-token"))
	if userID == "" {
		userID = userIDForSession(session)
	}
	if !userOwnsBot(thing, userID) {
		return nil, status.Error(codes.PermissionDenied, "bot not associated with account")
	}
	bundle, err := issueToken(ctx, userID, thing)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ChangeGUIDInIni(strings.TrimPrefix(thing, "vic:"))
	return bundle, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	logger.Println("Token: Incoming Associate Secondary Client request")
	bundle, err := s.refresh(ctx, true, req.UserSession)
	if err != nil {
		return nil, err
	}
	return &tokenpb.AssociateSecondaryClientResponse{Data: bundle}, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	logger.Println("Token: Incoming Refresh Token request")
	bundle, err := s.refresh(ctx, false, "")
	if err != nil {
		return nil, err
	}
	return &tokenpb.RefreshTokenResponse{Data: bundle}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
