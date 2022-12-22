package jdocsserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/chipper/pkg/tokenserver"
	"google.golang.org/grpc/peer"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func ConvertToProperJdoc(filename string) {
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	var jdoc *jdocspb.Jdoc
	jdoc.FmtVersion = 1
	jdoc.DocVersion = 1
	jdoc.JsonDoc = string(jsonBytes)
	writeBytes, err := json.Marshal(jdoc)
	if err != nil {
		fmt.Println(err)
		return
	}
	os.WriteFile(filename, writeBytes, 0644)
	return
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	fmt.Println("Jdocs: Incoming WriteDoc request, Item to write: " + req.DocName + ", Robot ID: " + req.Thing)
	filename := "./jdocs/" + strings.TrimSpace(req.Thing) + "-" + strings.TrimSpace(req.DocName) + ".json"
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
		fmt.Println(err)
	}
	os.WriteFile(filename, writeBytes, 0644)
	return &jdocspb.WriteDocResp{
		Status:           jdocspb.WriteDocResp_ACCEPTED,
		LatestDocVersion: latestVersion,
	}, nil
}

func (s *JdocServer) ReadDocs(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	globalGUIDHash := `{"client_tokens":[{"hash":"J5TAnJTPRCioMExFo5KzH2fHOAXyM5fuO8YRbQSamIsNzymnJ8KDIerFxuJV4qBN","client_name":"","app_id":"","issued_at":"2022-11-26T18:23:08Z","is_primary":true}]}`
	fmt.Println("Jdocs: Incoming ReadDocs request, Robot ID: " + req.Thing + ", Item(s) to return: ")
	fmt.Println(req.Items)
	StoreBotInfo(ctx, req.Thing)
	esn := strings.Split(req.Thing, ":")[1]
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.Split(p.Addr.String(), ":")[0]
	if strings.Contains(req.Items[0].DocName, "vic.AppToken") {
		if _, err := os.Stat("./jdocs/" + strings.TrimSpace(req.Thing) + "-vic.AppTokens.json"); err != nil {
			fmt.Println("App tokens jdoc not found for this bot, trying bots in TokenHashStore")
			matched := false
			for _, pair := range tokenserver.TokenHashStore {
				if strings.EqualFold(pair[0], ipAddr) {
					tokenserver.WriteTokenHash(esn, pair[1])
					fmt.Println("ReadJdocs: " + esn + " matched with bot " + ipAddr + "in token store")
					matched = true
				}
			}
			if !matched {
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
		filename := "./jdocs/" + strings.TrimSpace(req.Thing) + "-" + strings.TrimSpace(item.DocName) + ".json"
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
			fmt.Println(err)
			jdoc := jdocspb.Jdoc{}
			jdoc.DocVersion = 1
			jdoc.FmtVersion = 1
			jdoc.ClientMetadata = "wirepod-outdated-fmt"
			jdoc.JsonDoc = string(jsonByte)
			ConvertToProperJdoc(filename)
			fmt.Println("Deprecated jdoc format found, converting to proper jdoc")
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
