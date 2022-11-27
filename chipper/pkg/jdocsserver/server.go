package jdocsserver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	fmt.Println("Jdocs: Incoming WriteDoc request, Item to write: " + req.DocName + ", Robot ID: " + req.Thing)
	os.WriteFile("./jdocs/"+strings.TrimSpace(req.Thing)+"-"+strings.TrimSpace(req.DocName)+".json", []byte(req.Doc.JsonDoc), 0644)
	return &jdocspb.WriteDocResp{
		Status: jdocspb.WriteDocResp_ACCEPTED,
	}, nil
}
func (s *JdocServer) ReadDocs(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	fmt.Println("Jdocs: Incoming ReadDocs request, Robot ID: " + req.Thing + ", Item(s) to return: ")
	fmt.Println(req.Items)
	storeBotInfo(ctx, req.Thing)
	if strings.Contains(req.Items[0].DocName, "vic.AppToken") {
		return &jdocspb.ReadDocsResp{
			Items: []*jdocspb.ReadDocsResp_Item{
				{
					Status: jdocspb.ReadDocsResp_CHANGED,
					Doc: &jdocspb.Jdoc{
						DocVersion:     214,
						FmtVersion:     1,
						ClientMetadata: "placeholder",
						JsonDoc:        `{"client_tokens":[{"hash":"J5TAnJTPRCioMExFo5KzH2fHOAXyM5fuO8YRbQSamIsNzymnJ8KDIerFxuJV4qBN","client_name":"","app_id":"","issued_at":"2022-11-26T18:23:08Z","is_primary":true},{"hash":"9P3CesvMJeSA1rn9QdmqUi1BdW4stboEY+NNEcVU+SJcYwtjQaits/EtxXNkGipR","client_name":"escapepod","app_id":"SDK","issued_at":"2022-11-24T09:58:40Z","is_primary":false},{"hash":"uQHwJkDOBOSKxFAZ9LjXQ1pXLhbsQ4eJcWawxwYPQPz3DLQ3F8Ck53C1DV5tqPOl","client_name":"escapepod","app_id":"SDK","issued_at":"2022-11-24T09:57:18Z","is_primary":false},{"hash":"C+eBGDeAGbPePoDEQD7PtyK2SuV8958mQBSoRmy617wONJ/aAhj8p2J4QXh1KTmz","client_name":"","app_id":"SDK","issued_at":"2022-11-24T09:56:09Z","is_primary":false},{"hash":"YeVtE0t4ATT4AEPGiRnyyFVqTZXNo91TbDmMNK1TbWCEqyGXjlVm2VyK0QjiEIZE","client_name":"","app_id":"","issued_at":"2022-11-23T20:57:35Z","is_primary":false},{"hash":"xTyhqVDmxjOS6Vxv9lnaA/0gXreNrXHGCcMOgpQAU7tkX8KNYfIt40REqDwZS9Jl","client_name":"Vector-B7X3","app_id":"SDK","issued_at":"2022-11-13T16:28:14Z","is_primary":false},{"hash":"PG3xFA64DDFGjF9qmM5UdWZ0m+awffjsesB3ToqHNp712chugg1rZmXfJCptjGs5","client_name":"Vector-W4B2","app_id":"SDK","issued_at":"2022-11-13T04:12:38Z","is_primary":false},{"hash":"VnsPa6OmfDrbWnHycgbgqGTP3w/w2hnq1XP5MqSpCZpvga1KCdfJ1p2tYsuEimg3","client_name":"Vector-C3X3","app_id":"SDK","issued_at":"2022-11-13T03:07:28Z","is_primary":false},{"hash":"4il4GlGe1O5Acn3zsMaQte8+ti9hu7z0sNXHEKAp55i7RhcEKI7xWHa40+s6M0gW","client_name":"Vector-C3X3","app_id":"SDK","issued_at":"2022-11-12T19:43:52Z","is_primary":false},{"hash":"6yBkEaa/WCpcomOZtfGlOwbFaJe38Ocv8ikz1chZnrXmVbsmKHx3W4qJsWOfMN4o","client_name":"","app_id":"SDK","issued_at":"2022-10-17T18:26:37Z","is_primary":false},{"hash":"QHdWyVsdViPrC02UCdUsn1XMl5aJm/HaaOF98qLmpIoDDOHjR5fykT9JrAc3f7iD","client_name":"","app_id":"SDK","issued_at":"2022-10-16T05:30:44Z","is_primary":false},{"hash":"dsucqKTC+j0tPZIol5+jl3letW16sQVYSN+a+cEDYg79zSuAo/7Uafqj6NK1g8EA","client_name":"escapepod","app_id":"SDK","issued_at":"2022-08-18T03:29:43Z","is_primary":false},{"hash":"7KmmRM28jeUGAgVPD3Sz484IBgso/scVwfaj0krwMTZiu4Wlf9PB4wFZLbniGvRt","client_name":"","app_id":"SDK","issued_at":"2022-04-10T16:36:21Z","is_primary":false},{"hash":"u6+d6KA9sS91kZaeKJ7c6dLnu6aPC+TrXwS3TpolEjFs+oHnbApIv2hPmlY9c5PJ","client_name":"","app_id":"SDK","issued_at":"2022-04-10T04:53:20Z","is_primary":false},{"hash":"LNNU8AscTHcM1fT9CHyy9SbUbzs4l/GTpD+t7HYjvIR7X/M7rm+Oipf3jAkd7Q1d","client_name":"","app_id":"SDK","issued_at":"2022-04-10T03:35:48Z","is_primary":false},{"hash":"ts6sZyh/JMWiGmxXU9T7SS9YSuotdVHzLyci6YmoJfSE6YEOw9mSbVPcRumQ1i3t","client_name":"","app_id":"SDK","issued_at":"2022-04-10T03:35:38Z","is_primary":false},{"hash":"mBOBS6Rvw/d/VzBlpzELO+WLP+1+gTWPsEwvDHBrz5HosibHjs7pMMJhN1nJ0QMa","client_name":"","app_id":"SDK","issued_at":"2022-04-10T03:29:01Z","is_primary":false},{"hash":"Dy86qgnnD2vB4LSwH0YreTtnOEH/Y5FgdSCt1C3WMUSRtAPpSWvLEo8qsfaDnwWE","client_name":"thinkpad","app_id":"SDK","issued_at":"2022-02-24T04:05:05Z","is_primary":false},{"hash":"WiiLgeE1gt/5RbDsbzRvO3hhU4vowRtJ/OjOEFEe9j7V5bUGMSdxDli3fidqzqnc","client_name":"","app_id":"SDK","issued_at":"2022-02-23T02:18:37Z","is_primary":false},{"hash":"vSc5C6QRQX22bRmWhtB2wszNTHJ3aBf3Y3qUZjEvAiF1Bx04Cz07Fe4YxeJ+i9kV","client_name":"","app_id":"SDK","issued_at":"2022-02-06T06:17:20Z","is_primary":false},{"hash":"5pj/66KDtKfOzZIOYhTAL4icsVIBTF6UdsXny1j2KjcajyYi+4C5HERZXtUMW96p","client_name":"","app_id":"SDK","issued_at":"2022-01-31T05:23:48Z","is_primary":false},{"hash":"BrsCUNhaAwEO3rQDvJ4E7KIrAH6dYYhilaZDR4rkui3U8FqiBizX1MynYGS1djgY","client_name":"","app_id":"SDK","issued_at":"2022-01-31T05:23:42Z","is_primary":false},{"hash":"dtsnjA8KdUoINcf9/ThRjufxrD4aWU/tg0gpVmK0PjDUIZr9qlFqFYX7o7oC7uon","client_name":"","app_id":"SDK","issued_at":"2022-01-31T05:23:30Z","is_primary":false},{"hash":"MIvAPkwwkf/kvhmPASBoCJ++vcTOA/V9Hthk/zeZ/S7DyoxwNfPzZpLXwLdhJdru","client_name":"","app_id":"SDK","issued_at":"2022-01-31T05:18:19Z","is_primary":false},{"hash":"0yVLQ9eed2detp5OZH4oCEuqRLyiERZ8gwYccaRg0hMNT1Zq2rpeFiniCTp39NEG","client_name":"","app_id":"SDK","issued_at":"2022-01-31T05:18:13Z","is_primary":false}]}`,
					},
				},
			},
		}, nil
	}
	var returnItems []*jdocspb.ReadDocsResp_Item
	for _, item := range req.Items {
		jsonByte, err := os.ReadFile("./jdocs/" + strings.TrimSpace(req.Thing) + "-" + strings.TrimSpace(item.DocName) + ".json")
		if err != nil {
			returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_NOT_FOUND, Doc: &jdocspb.Jdoc{DocVersion: 1, FmtVersion: 1, ClientMetadata: "placeholder", JsonDoc: ""}})
			continue
		}
		returnItems = append(returnItems, &jdocspb.ReadDocsResp_Item{Status: jdocspb.ReadDocsResp_CHANGED, Doc: &jdocspb.Jdoc{DocVersion: 1, FmtVersion: 1, ClientMetadata: "placeholder", JsonDoc: string(jsonByte)}})
	}
	return &jdocspb.ReadDocsResp{Items: returnItems}, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
