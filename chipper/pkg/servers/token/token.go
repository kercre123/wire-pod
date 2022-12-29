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

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
	"github.com/kercre123/chipper/pkg/logger"
	"google.golang.org/grpc/peer"
	"gopkg.in/ini.v1"
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
	InfoPath       = JdocsPath + BotInfoFile
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
			logger.Println(robot.Esn + " sucessfully activated with wire-pod")
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Errorf("bot not found")
	}
	writeBytes, err := json.Marshal(robotInfo)
	if err != nil {
		logger.Println(err)
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
	if err != nil {
		logger.Println(err)
	}
	return err
}

func RemoveFromSecondStore(s [][4]string, index int) {
	logger.Println("Removing " + s[index][0] + " from temporary token-hash store")
	SecondaryTokenStore = append(s[:index], s[index+1:]...)
	return
}

func RemoveFromPrimaryStore(s [][3]string, index int) {
	logger.Println("Removing " + s[index][0] + " from temporary token-hash store")
	TokenHashStore = append(s[:index], s[index+1:]...)
	return
}

func RemoveFromSessionStore(index int) {
	//var SessionWriteStoreNames [][2]string
	//var SessionWriteStoreCerts [][]byte
	logger.Println("Removing " + SessionWriteStoreNames[index][0] + " from cert-write store")
	SessionWriteStoreNames = append(SessionWriteStoreNames[:index], SessionWriteStoreNames[index+1:]...)
	SessionWriteStoreCerts = append(SessionWriteStoreCerts[:index], SessionWriteStoreCerts[index+1:]...)
	return
}

func ChangeGUIDInIni(esn string) {
	// 	[008060ec]
	// cert = /home/kerigan/.anki_vector/Vector-B6H9-008060ec.cert
	// ip = 192.168.1.155
	// name = Vector-B6H9
	// guid = 1YbXk1yrS9C1I78snYy8xA==

	var robotSDKInfo RobotInfoStore
	eFileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	} else {
		logger.Println("Error reading robotSDKInfo")
		logger.Println(err)
		return
	}
	userIniData, err := ini.Load("../../.anki_vector/sdk_config.ini")
	if err != nil {
		logger.Println(err)
		return
	}
	for _, robot := range robotSDKInfo.Robots {
		matched := false
		for _, section := range userIniData.Sections() {
			if strings.EqualFold(section.Name(), esn) {
				matched = true
				section.Key("ip").SetValue(robot.IPAddress)
				if robot.GUID == "" {
					section.Key("guid").SetValue(robotSDKInfo.GlobalGUID)
				} else {
					section.Key("guid").SetValue(robot.GUID)
				}
			}
		}
		if !matched {
			logger.Println("Bot is not in sdk_config.ini. Clear your bot's userdata and try authenticating again to create it.")
		}
	}
	userIniData.SaveTo("../../.anki_vector/sdk_config.ini")
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
				RemoveFromSecondStore(SecondaryTokenStore, num)
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
		logger.Println("ESN not found in store or this is an associate primary user request, this bot is new")
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
