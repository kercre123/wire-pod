package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	"google.golang.org/grpc/peer"
	"gopkg.in/ini.v1"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

var (
	TimeFormat     = time.RFC3339Nano
	ExpirationTime = time.Hour * 24
	UserId         = "wirepod"
	GlobalGUID     = "tni1TRsTRTaNSapjo0Y+Sw=="
)

// array of {"target", "guid", "guidhash"}
// for primary user auth
var TokenHashStore [][3]string

// array of {"esn", "target", "guid", "guidhash"}
// for secondary user auth
var SecondaryTokenStore [][4]string

// {"target", "name"}
var SessionWriteStoreNames [][2]string
var SessionWriteStoreCerts [][]byte

type RobotInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
		// 192.168.1.150:443
		GUID      string `json:"guid"`
		Activated bool   `json:"activated"`
	} `json:"robots"`
}

func GetEsnFromTarget(target string) (string, error) {
	jsonBytes, err := os.ReadFile(vars.BotInfoPath)
	if err != nil {
		return "", err
	}
	var robotInfo RobotInfoStore
	err = json.Unmarshal(jsonBytes, &robotInfo)
	if err != nil {
		return "", err
	}
	for _, robot := range robotInfo.Robots {
		if strings.TrimSpace(target) == strings.TrimSpace(robot.IPAddress) {
			return robot.Esn, nil
		}
	}
	return "", fmt.Errorf("bot not found")
}

func SetBotGUID(esn string, guid string, guidHash string) error {
	matched := false
	for num, robot := range vars.BotInfo.Robots {
		if strings.EqualFold(esn, robot.Esn) {
			vars.BotInfo.Robots[num].GUID = guid
			vars.BotInfo.Robots[num].Activated = true
			logger.Println("GUID and hash successfully written for " + robot.Esn)
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Errorf("bot not found")
	}
	writeBytes, err := json.Marshal(vars.BotInfo)
	if err != nil {
		logger.Println(err)
		return err
	}
	os.WriteFile(vars.BotInfoPath, writeBytes, 0644)
	return nil
}

func WriteTokenHash(esn string, tokenHash string) error {
	// will return blank jdoc if it doesn't exist
	jdoc, jdocExists := vars.GetJdoc(esn, "vic.AppTokens")
	var tokenJson ClientTokenManager
	if !jdocExists {
		jdoc.DocVersion = 1
		jdoc.FmtVersion = 1
		jdoc.ClientMetadata = "wirepod-new-token"
	}
	json.Unmarshal([]byte(jdoc.JsonDoc), &tokenJson)
	var clientToken ClientToken
	clientToken.IssuedAt = time.Now().Format(TimeFormat)
	clientToken.ClientName = "wirepod"
	clientToken.Hash = tokenHash
	clientToken.AppId = "SDK"
	tokenJson.ClientTokens = append(tokenJson.ClientTokens, clientToken)
	jdocJsoc, err := json.Marshal(tokenJson)
	if err != nil {
		logger.Println("Error marshaling token hash json")
		logger.Println(err)
	}
	jdoc.JsonDoc = string(jdocJsoc)
	var ajdoc vars.AJdoc
	ajdoc.ClientMetadata = jdoc.ClientMetadata
	ajdoc.DocVersion = jdoc.DocVersion
	ajdoc.FmtVersion = jdoc.FmtVersion
	ajdoc.JsonDoc = jdoc.JsonDoc
	vars.AddJdoc("vic:"+esn, "vic.AppTokens", ajdoc)
	vars.WriteJdocs()
	return nil
}

func RemoveFromSecondStore(index int) {
	logger.Println("Removing " + SecondaryTokenStore[index][0] + " from temporary token-hash store")
	SecondaryTokenStore = append(SecondaryTokenStore[:index], SecondaryTokenStore[index+1:]...)
}

func RemoveFromPrimaryStore(index int) {
	logger.Println("Removing " + TokenHashStore[index][0] + " from temporary token-hash store")
	TokenHashStore = append(TokenHashStore[:index], TokenHashStore[index+1:]...)
}

func RemoveFromSessionStore(index int) {
	//var SessionWriteStoreNames [][2]string
	//var SessionWriteStoreCerts [][]byte
	logger.Println("Removing " + SessionWriteStoreNames[index][0] + " from cert-write store")
	SessionWriteStoreNames = append(SessionWriteStoreNames[:index], SessionWriteStoreNames[index+1:]...)
	SessionWriteStoreCerts = append(SessionWriteStoreCerts[:index], SessionWriteStoreCerts[index+1:]...)
}

func ChangeGUIDInIni(esn string) {
	// 	[008060ec]
	// cert = /home/kerigan/.anki_vector/Vector-B6H9-008060ec.cert
	// ip = 192.168.1.155
	// name = Vector-B6H9
	// guid = 1YbXk1yrS9C1I78snYy8xA==

	userIniData, err := ini.Load(vars.SDKIniPath + "sdk_config.ini")
	if err != nil {
		logger.Println(err)
		return
	}
	for _, robot := range vars.BotInfo.Robots {
		matched := false
		for _, section := range userIniData.Sections() {
			if strings.EqualFold(section.Name(), esn) {
				matched = true
				section.Key("ip").SetValue(robot.IPAddress)
				if robot.GUID == "" {
					section.Key("guid").SetValue(vars.BotInfo.GlobalGUID)
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

func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

func CreateJWT(ctx context.Context, skipGuid bool, isPrimary bool) *tokenpb.TokenBundle {
	// defaults
	requestorId := "vic:00601b50"
	clientToken := GlobalGUID
	bundle := &tokenpb.TokenBundle{}
	secondary := false
	secondaryGUID := ""
	secondaryHash := ""

	// figure out current time and the time in one day
	currentTime := time.Now().Format(TimeFormat)
	expiresAt := time.Now().Add(ExpirationTime).Format(TimeFormat)
	logger.Println("Current time: " + currentTime)
	logger.Println("Token expires: " + expiresAt)

	// get esn using ip address of request
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.TrimSpace(strings.Split(p.Addr.String(), ":")[0])
	esn, err := GetEsnFromTarget(ipAddr)

	// secondary handler
	if err == nil {
		for num, robot := range SecondaryTokenStore {
			if robot[0] == esn {
				skipGuid = true
				secondary = true
				secondaryGUID = robot[2]
				secondaryHash = robot[3]
				RemoveFromSecondStore(num)
				break
			}
		}
	}

	// create token and hash
	// if esn is not found, put tokenHash into ram
	if err == nil && !isPrimary {
		logger.Println("Found ESN for target " + ipAddr + ": " + esn)
		requestorId = "vic:" + esn
		if !skipGuid {
			guid, tokenHash, _ := CreateTokenAndHashedToken()
			WriteTokenHash(esn, tokenHash)
			SetBotGUID(esn, guid, tokenHash)
			ChangeGUIDInIni(esn)
			clientToken = guid
		}
	} else {
		logger.Println("ESN not found in store or this is an associate primary user request, act as if this is a new robot")
		if !skipGuid {
			logger.Println("Adding " + ipAddr + " to TokenHashStore")
			guid, tokenHash, _ := CreateTokenAndHashedToken()
			TokenHashStore = append(TokenHashStore, [3]string{ipAddr, guid, tokenHash})
			clientToken = guid
		}
	}
	if !skipGuid {
		bundle.ClientToken = clientToken
	}

	if secondary {
		SetBotGUID(esn, secondaryGUID, secondaryHash)
		bundle.ClientToken = secondaryGUID
		logger.Println("Secondary client: " + secondaryGUID)
	}

	requestUUID := GenerateUUID()
	logger.Println("UUID for this token request: " + requestUUID)

	// create actual JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":     expiresAt,
		"iat":         currentTime,
		"permissions": nil,
		// the requestorId will be vic:00601b50 on first auth because we don't have access
		// to the factory certs like the official servers do. future token requests should
		// have the actual bot esn because they are "associated" with wire-pod
		"requestor_id": requestorId,
		"token_id":     requestUUID,
		"token_type":   "user+robot",
		"user_id":      UserId,
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	bundle.Token = tokenString
	return bundle
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	logger.Println("Token: Incoming Associate Primary User request")
	pemBytes, _ := pem.Decode(req.SessionCertificate)
	cert, _ := x509.ParseCertificate(pemBytes.Bytes)
	SessionWriteStoreCerts = append(SessionWriteStoreCerts, req.SessionCertificate)
	p, _ := peer.FromContext(ctx)
	SessionWriteStoreNames = append(SessionWriteStoreNames, [2]string{p.Addr.String(), cert.Issuer.CommonName})
	return &tokenpb.AssociatePrimaryUserResponse{
		Data: CreateJWT(ctx, false, true),
	}, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	logger.Println("Token: Incoming Associate Secondary Client request")
	return &tokenpb.AssociateSecondaryClientResponse{
		Data: CreateJWT(ctx, false, false),
	}, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	logger.Println("Token: Incoming Refresh Token request")
	return &tokenpb.RefreshTokenResponse{
		Data: CreateJWT(ctx, false, false),
	}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
