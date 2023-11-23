package jdocsserver

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/kercre123/chipper/pkg/logger"
	tokenserver "github.com/kercre123/chipper/pkg/servers/token"
	"github.com/kercre123/chipper/pkg/vars"
	"google.golang.org/grpc/peer"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	logger.Println("Jdocs: Incoming WriteDoc request, Item to write: " + req.DocName + ", Robot ID: " + req.Thing)
	var ajdoc vars.AJdoc
	ajdoc.ClientMetadata = req.Doc.ClientMetadata
	ajdoc.DocVersion = req.Doc.DocVersion
	ajdoc.FmtVersion = req.Doc.FmtVersion
	ajdoc.JsonDoc = req.Doc.JsonDoc
	latestVersion := vars.AddJdoc(req.Thing, req.DocName, ajdoc)
	vars.WriteJdocs()

	esn := strings.Split(req.Thing, ":")[1]
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.Split(p.Addr.String(), ":")[0]

	for ind, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn && bot.IPAddress != ipAddr {
			logger.Println(esn + "'s IP address has changed to " + ipAddr + ", noting")
			vars.BotInfo.Robots[ind].IPAddress = ipAddr
			writeBytes, _ := json.Marshal(vars.BotInfo)
			os.WriteFile(vars.BotInfoPath, writeBytes, 0644)
		}
	}

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
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.Split(p.Addr.String(), ":")[0]

	for ind, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn && bot.IPAddress != ipAddr {
			logger.Println(esn + "'s IP address has changed to " + ipAddr + ", noting")
			vars.BotInfo.Robots[ind].IPAddress = ipAddr
			writeBytes, _ := json.Marshal(vars.BotInfo)
			os.WriteFile(vars.BotInfoPath, writeBytes, 0644)
		}
	}

	for _, pair := range tokenserver.SessionWriteStoreNames {
		if ipAddr == strings.Split(pair[0], ":")[0] {
			vars.DeleteData(req.Thing)
			break
		}
	}
	if strings.Contains(req.Items[0].DocName, "vic.AppTokens") {
		StoreBotInfo(ctx, req.Thing)
		_, tokenExists := vars.GetJdoc(req.Thing, "vic.AppTokens")
		if !tokenExists {
			logger.Println("App tokens jdoc not found for this bot, trying bots in TokenHashStore")
			matched := false
			botGUID := ""
			for num, pair := range tokenserver.TokenHashStore {
				if strings.EqualFold(pair[0], ipAddr) {
					err := tokenserver.WriteTokenHash(strings.ToLower(strings.TrimSpace(esn)), pair[2])
					if err != nil {
						logger.Println("Error writing token hash to vic.AppTokens")
						logger.Println(err)
					}
					err = tokenserver.SetBotGUID(esn, pair[1], pair[2])
					botGUID = pair[1]
					if err != nil {
						logger.Println("Error writing token hash to " + vars.BotInfoPath)
						logger.Println(err)
					}
					logger.Println("ReadJdocs: bot " + esn + " matched with IP " + ipAddr + " in token store")
					matched = true
					tokenserver.RemoveFromPrimaryStore(num)
				}
			}
			sessionMatched := false
			for num, pair := range tokenserver.SessionWriteStoreNames {
				if strings.EqualFold(ipAddr, strings.Split(pair[0], ":")[0]) {
					sessionMatched = true
					fullPath := vars.SDKIniPath + pair[1] + "-" + esn + ".cert"
					if _, err := os.Stat(vars.SDKIniPath); err != nil {
						logger.Println("Creating " + vars.SDKIniPath + " directory")
						os.Mkdir(vars.SDKIniPath, 0755)
					}
					logger.Println("Outputting session cert to " + fullPath)
					// export to ~/.anki_vector
					os.WriteFile(fullPath, tokenserver.SessionWriteStoreCerts[num], 0755)
					// export to ./session-certs
					os.WriteFile(vars.SessionCertPath+"/"+esn, tokenserver.SessionWriteStoreCerts[num], 0755)
					WriteToIniPrimary(pair[1], esn, botGUID, ipAddr)
					tokenserver.RemoveFromSessionStore(num)
					logger.Println("Session certificate successfully output")
					break
				}
			}
			logger.LogUI("New bot being associated with wire-pod. ESN: " + esn + ", IP: " + ipAddr)
			if !matched {
				if !isAlreadyKnown {
					logger.Println("Bot was not known to wire-pod, creating token and hash (in ReadDocs)")
					guid, hash, _ := tokenserver.CreateTokenAndHashedToken()
					tokenserver.SecondaryTokenStore = append(tokenserver.SecondaryTokenStore, [4]string{esn, ipAddr, guid, hash})
					// creates apptoken jdoc file
					tokenserver.WriteTokenHash(esn, hash)
					if !sessionMatched {
						WriteToIniSecondary(esn, guid, ipAddr)
					}
					// bot is not authenticated yet, do not write to botinfo json
					tokenJdoc, _ := vars.GetJdoc(req.Thing, "vic.AppToken")
					var truejdoc jdocspb.Jdoc
					truejdoc.ClientMetadata = tokenJdoc.ClientMetadata
					truejdoc.DocVersion = tokenJdoc.DocVersion
					truejdoc.FmtVersion = tokenJdoc.FmtVersion
					truejdoc.JsonDoc = tokenJdoc.JsonDoc
					tokenserver.RemoveFromSecondStore(len(tokenserver.SecondaryTokenStore) - 1)
					return &jdocspb.ReadDocsResp{
						Items: []*jdocspb.ReadDocsResp_Item{
							{
								Status: jdocspb.ReadDocsResp_CHANGED,
								Doc:    &truejdoc,
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
		gottenDoc, jdocExists := vars.GetJdoc(req.Thing, item.DocName)
		if jdocExists {
			var truejdoc jdocspb.Jdoc
			truejdoc.DocVersion = gottenDoc.DocVersion
			truejdoc.FmtVersion = gottenDoc.FmtVersion
			truejdoc.ClientMetadata = gottenDoc.ClientMetadata
			truejdoc.JsonDoc = gottenDoc.JsonDoc
			returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_CHANGED, Doc: &truejdoc})
		} else {
			var noJdoc jdocspb.Jdoc
			noJdoc.DocVersion = 0
			noJdoc.FmtVersion = 0
			noJdoc.ClientMetadata = "wirepod-noexist"
			noJdoc.JsonDoc = ""
			returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_CHANGED, Doc: &noJdoc})
		}
	}
	return &jdocspb.ReadDocsResp{Items: returnItems}, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
