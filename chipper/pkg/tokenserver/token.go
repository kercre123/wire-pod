package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/peer"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

const (
	TimeFormat     = time.RFC3339Nano
	JdocsPath      = "./jdocs/"
	BotInfoFile    = "botSdkInfo.json"
	ExpirationTime = time.Hour * 24
	UserId         = "wirepod"
	GlobalGUID     = "tni1TRsTRTaNSapjo0Y+Sw=="
)

// array of {"target", "guid", "guidhash"}
// temporary and used for less than a second per bot, so it is better to keep in ram
var TokenHashStore [][3]string

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
	jsonBytes, err := os.ReadFile(JdocsPath + BotInfoFile)
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
	jsonBytes, err := os.ReadFile(JdocsPath + BotInfoFile)
	if err != nil {
		return err
	}
	var robotInfo RobotInfoStore
	err = json.Unmarshal(jsonBytes, &robotInfo)
	if err != nil {
		return err
	}
	matched := false
	for num, robot := range robotInfo.Robots {
		if strings.EqualFold(esn, robot.Esn) {
			robotInfo.Robots[num].GUID = guid
			robotInfo.Robots[num].Activated = true
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Errorf("bot not found")
	}
	writeBytes, err := json.Marshal(robotInfo)
	if err != nil {
		fmt.Println(err)
		return err
	}
	os.WriteFile(JdocsPath+BotInfoFile, writeBytes, 0644)
	return nil
}

func WriteTokenHash(esn string, tokenHash string) error {
	var tokenJson ClientTokenManager
	jdoc := jdocspb.Jdoc{}
	filename := JdocsPath + "vic:" + esn + "-vic.AppTokens.json"
	jsonBytes, err := os.ReadFile(filename)
	if err == nil {
		json.Unmarshal(jsonBytes, &jdoc)
	} else {
		jdoc.DocVersion = 1
		jdoc.FmtVersion = 1
		jdoc.ClientMetadata = "wirepod-new-tokens"
	}
	var clientToken ClientToken
	clientToken.IssuedAt = time.Now().Format(TimeFormat)
	clientToken.ClientName = "wirepod"
	clientToken.Hash = tokenHash
	clientToken.AppId = "SDK"
	tokenJson.ClientTokens = append(tokenJson.ClientTokens, clientToken)
	newJsonBytes, _ := json.Marshal(tokenJson)
	jdoc.JsonDoc = string(newJsonBytes)
	writeBytes, _ := json.Marshal(jdoc)
	err = os.WriteFile(filename, writeBytes, 0644)
	return err
}

func createJWT(ctx context.Context, skipGuid bool) *tokenpb.TokenBundle {
	// defaults
	requestorId := "vic:00601b50"
	clientToken := GlobalGUID
	bundle := &tokenpb.TokenBundle{}

	// figure out current time and the time in one day
	currentTime := time.Now().Format(TimeFormat)
	expiresAt := time.Now().Add(ExpirationTime).Format(TimeFormat)
	fmt.Println("Current time: " + currentTime)
	fmt.Println("Token expires: " + expiresAt)

	// get esn using ip address of request
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.TrimSpace(strings.Split(p.Addr.String(), ":")[0])
	esn, err := GetEsnFromTarget(ipAddr)

	// create token and hash
	// if esn is not found, put tokenHash into ram
	if err == nil {
		fmt.Println("Found ESN for target " + ipAddr + ": " + esn)
		requestorId = "vic:" + esn
		if !skipGuid {
			guid, tokenHash, _ := CreateTokenAndHashedToken()
			WriteTokenHash(esn, tokenHash)
			SetBotGUID(esn, guid, tokenHash)
			clientToken = guid
		}
	} else {
		fmt.Println("ESN not found in store, this bot is new")
		if !skipGuid {
			fmt.Println("Adding " + ipAddr + " to TokenHashStore")
			guid, tokenHash, _ := CreateTokenAndHashedToken()
			TokenHashStore = append(TokenHashStore, [3]string{ipAddr, guid, tokenHash})
			clientToken = guid
		}
	}
	if !skipGuid {
		bundle.ClientToken = clientToken
	}

	// create actual JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      expiresAt,
		"iat":          currentTime,
		"permissions":  nil,
		"requestor_id": requestorId,
		"token_id":     "11ec68ca-1d4c-4e45-b1a2-715fd5e0abf9",
		"token_type":   "user+robot",
		"user_id":      UserId,
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	bundle.Token = tokenString
	return bundle
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	fmt.Println("Token: Incoming Associate Primary User request")
	return &tokenpb.AssociatePrimaryUserResponse{
		Data: createJWT(ctx, req.SkipClientToken),
	}, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	fmt.Println("Token: Incoming Associate Secondary Client request")
	return &tokenpb.AssociateSecondaryClientResponse{
		Data: createJWT(ctx, false),
	}, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	fmt.Println("Token: Incoming Refresh Token request")
	var refresh bool = true
	if req.RefreshJwtTokens {
		refresh = false
	}
	return &tokenpb.RefreshTokenResponse{
		Data: createJWT(ctx, refresh),
	}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
