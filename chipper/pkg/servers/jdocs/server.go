package jdocsserver

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/kercre123/chipper/pkg/logger"
	tokenserver "github.com/kercre123/chipper/pkg/servers/token"
	"google.golang.org/grpc/peer"
)

const (
	JdocsPath = tokenserver.JdocsPath
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func ConvertToProperJdoc(filename string) {
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		logger.Println(err)
		return
	}
	var jdoc *jdocspb.Jdoc
	jdoc.FmtVersion = 1
	jdoc.DocVersion = 1
	jdoc.JsonDoc = string(jsonBytes)
	writeBytes, err := json.Marshal(jdoc)
	if err != nil {
		logger.Println(err)
		return
	}
	os.WriteFile(filename, writeBytes, 0644)
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	logger.Println("Jdocs: Incoming WriteDoc request, Item to write: " + req.DocName + ", Robot ID: " + req.Thing)
	filename := JdocsPath + strings.TrimSpace(req.Thing) + "-" + strings.TrimSpace(req.DocName) + ".json"
	var latestVersion uint64 = 0

	// decode already-existing json (if it exists)
	jdocBytes, err := os.ReadFile(filename)
	if err != nil {
		jdoc := jdocspb.Jdoc{}
		json.Unmarshal(jdocBytes, &jdoc)
		latestVersion = jdoc.DocVersion
	}

	// encode to json
	jdoc := jdocspb.Jdoc{}
	jdoc.DocVersion = req.Doc.DocVersion
	jdoc.FmtVersion = req.Doc.FmtVersion
	jdoc.JsonDoc = req.Doc.JsonDoc
	writeBytes, err := json.Marshal(jdoc)
	if err != nil {
		logger.Println(err)
	}
	os.WriteFile(filename, writeBytes, 0644)
	return &jdocspb.WriteDocResp{
		Status:           jdocspb.WriteDocResp_ACCEPTED,
		LatestDocVersion: latestVersion,
	}, nil
}

func (s *JdocServer) ReadDocs(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	globalGUIDHash := `{"client_tokens":[{"hash":"J5TAnJTPRCioMExFo5KzH2fHOAXyM5fuO8YRbQSamIsNzymnJ8KDIerFxuJV4qBN","client_name":"","app_id":"","issued_at":"2022-11-26T18:23:08Z","is_primary":true}]}`
	// global guid now only used in edge cases
	logger.Println("Jdocs: Incoming ReadDocs request, Robot ID: " + req.Thing + ", Item(s) to return: ")
	logger.Println(req.Items)
	esn := strings.Split(req.Thing, ":")[1]
	isAlreadyKnown := IsBotInInfo(esn)
	StoreBotInfo(ctx, req.Thing)
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.Split(p.Addr.String(), ":")[0]
	if strings.Contains(req.Items[0].DocName, "vic.AppToken") {
		if _, err := os.Stat(JdocsPath + strings.TrimSpace(req.Thing) + "-vic.AppTokens.json"); err != nil {
			logger.Println("App tokens jdoc not found for this bot, trying bots in TokenHashStore")
			matched := false
			for num, pair := range tokenserver.TokenHashStore {
				if strings.EqualFold(pair[0], ipAddr) {
					err := tokenserver.WriteTokenHash(strings.ToLower(strings.TrimSpace(esn)), pair[2])
					if err != nil {
						logger.Println("Error writing token hash to vic.AppTokens.json")
						logger.Println(err)
					}
					err = tokenserver.SetBotGUID(esn, pair[1], pair[2])
					if err != nil {
						logger.Println("Error writing token hash to " + tokenserver.BotInfoFile)
						logger.Println(err)
					}
					logger.Println("ReadJdocs: bot " + esn + " matched with IP " + ipAddr + " in token store")
					matched = true
					tokenserver.RemoveFromPrimaryStore(tokenserver.TokenHashStore, num)
				}
			}
			sessionMatched := false
			for num, pair := range tokenserver.SessionWriteStoreNames {
				if ipAddr == strings.Split(pair[0], ":")[0] {
					sessionMatched = true
					fullPath, _ := os.Getwd()
					fullPath = strings.TrimSuffix(fullPath, "/wire-pod/chipper") + "/.anki_vector/" + pair[1] + "-" + esn + ".cert"
					logger.Println(fullPath)
					os.WriteFile(fullPath, tokenserver.SessionWriteStoreCerts[num], 0755)
					WriteToIni(pair[1])
					break
				}
			}
			if !sessionMatched {
				WriteToIni("")
			}
			if !matched {
				if !isAlreadyKnown {
					logger.Println("Bot was not known to wire-pod, creating token and hash (in ReadDocs)")
					guid, hash, _ := tokenserver.CreateTokenAndHashedToken()
					tokenserver.SecondaryTokenStore = append(tokenserver.SecondaryTokenStore, [4]string{esn, ipAddr, guid, hash})
					// creates apptoken jdoc file
					tokenserver.WriteTokenHash(esn, hash)
					// bot is not authenticated yet, do not write to botinfo json
					filename := JdocsPath + strings.TrimSpace(req.Thing) + "-vic.AppTokens.json"
					fileBytes, _ := os.ReadFile(filename)
					var jdoc jdocspb.Jdoc
					json.Unmarshal(fileBytes, &jdoc)
					return &jdocspb.ReadDocsResp{
						Items: []*jdocspb.ReadDocsResp_Item{
							{
								Status: jdocspb.ReadDocsResp_CHANGED,
								Doc:    &jdoc,
							},
						},
					}, nil
				}
				logger.Println("Bot not found in any store, providing global GUID")
				return &jdocspb.ReadDocsResp{
					Items: []*jdocspb.ReadDocsResp_Item{
						{
							Status: jdocspb.ReadDocsResp_CHANGED,
							Doc: &jdocspb.Jdoc{
								DocVersion:     1,
								FmtVersion:     1,
								ClientMetadata: "placeholder",
								JsonDoc:        globalGUIDHash,
							},
						},
					},
				}, nil
			}
		}
	}
	var returnItems []*jdocspb.ReadDocsResp_Item
	for _, item := range req.Items {
		filename := JdocsPath + strings.TrimSpace(req.Thing) + "-" + strings.TrimSpace(item.DocName) + ".json"
		jsonByte, err := os.ReadFile(filename)
		if err != nil {
			jdoc := jdocspb.Jdoc{}
			jdoc.DocVersion = 0
			jdoc.FmtVersion = 0
			jdoc.ClientMetadata = "wirepod-error"
			returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_NOT_FOUND, Doc: &jdoc})
			continue
		}
		jdoc := jdocspb.Jdoc{}
		err = json.Unmarshal(jsonByte, &jdoc)
		if err != nil {
			logger.Println(err)
			jdoc := jdocspb.Jdoc{}
			jdoc.DocVersion = 1
			jdoc.FmtVersion = 1
			jdoc.ClientMetadata = "wirepod-outdated-fmt"
			jdoc.JsonDoc = string(jsonByte)
			ConvertToProperJdoc(filename)
			logger.Println("Deprecated jdoc format found, converting to proper jdoc")
			returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_CHANGED, Doc: &jdoc})
			continue
		}
		returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_CHANGED, Doc: &jdoc})
	}
	return &jdocspb.ReadDocsResp{Items: returnItems}, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
