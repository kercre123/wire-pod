package jdocsserver

import (
	"context"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	tokenserver "github.com/kercre123/wire-pod/chipper/pkg/servers/token"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func peerIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}
	return vars.HostOnly(p.Addr.String())
}

func authorizeThing(ctx context.Context, rawThing string) (string, string, error) {
	esn, err := vars.ESNFromThing(rawThing)
	if err != nil {
		return "", "", status.Error(codes.InvalidArgument, err.Error())
	}
	thing := tokenserver.Thingifier(rawThing)
	if certThing, ok := tokenserver.PeerCertThing(ctx); ok && certThing != thing {
		return "", "", status.Error(codes.PermissionDenied, "not authorized")
	}
	return thing, esn, nil
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	thing, esn, err := authorizeThing(ctx, req.Thing)
	if err != nil {
		return nil, err
	}
	if req.Doc == nil {
		return nil, status.Error(codes.InvalidArgument, "missing doc")
	}
	logger.Println("Jdocs: Incoming WriteDoc request, Item to write: " + req.DocName + ", Robot ID: " + thing)
	vars.AddJdoc(thing, req.DocName, vars.AJdoc{
		DocVersion:     req.Doc.DocVersion,
		FmtVersion:     req.Doc.FmtVersion,
		ClientMetadata: req.Doc.ClientMetadata,
		JsonDoc:        req.Doc.JsonDoc,
	})
	if err := vars.RecordBot(esn, peerIP(ctx), ""); err != nil {
		logger.Println("Error writing bot info: ", err)
	}
	return &jdocspb.WriteDocResp{
		Status:           jdocspb.WriteDocResp_ACCEPTED,
		LatestDocVersion: req.Doc.DocVersion,
	}, nil
}

func (s *JdocServer) ReadDocs(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	thing, esn, err := authorizeThing(ctx, req.Thing)
	if err != nil {
		return nil, err
	}
	logger.Println("Jdocs: Incoming ReadDocs request, Robot ID: " + thing + ", Item(s) to return: ")
	logger.Println(req.Items)
	ip := peerIP(ctx)
	if err := vars.RecordBot(esn, ip, ""); err != nil {
		logger.Println("Error writing bot info: ", err)
	}

	var resp jdocspb.ReadDocsResp
	for _, item := range req.Items {
		if item == nil {
			continue
		}
		ajdoc, exists := vars.GetJdoc(thing, item.DocName)
		if !exists && item.DocName == "vic.AppTokens" {
			logger.Println("App tokens jdoc not found for " + esn + ", creating token and hash")
			guid, err := tokenserver.AddAppToken(thing)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if err := vars.RecordBot(esn, ip, guid); err != nil {
				logger.Println("Error writing bot info: ", err)
			}
			tokenserver.WriteToIniSecondary(esn, guid, ip)
			ajdoc, exists = vars.GetJdoc(thing, item.DocName)
		}
		if !exists {
			resp.Items = append(resp.Items, &jdocspb.ReadDocsResp_Item{
				Status: jdocspb.ReadDocsResp_CHANGED,
				Doc: &jdocspb.Jdoc{
					ClientMetadata: "wirepod-noexist",
				},
			})
			continue
		}
		resp.Items = append(resp.Items, &jdocspb.ReadDocsResp_Item{
			Status: jdocspb.ReadDocsResp_CHANGED,
			Doc: &jdocspb.Jdoc{
				DocVersion:     ajdoc.DocVersion,
				FmtVersion:     ajdoc.FmtVersion,
				ClientMetadata: ajdoc.ClientMetadata,
				JsonDoc:        ajdoc.JsonDoc,
			},
		})
	}
	return &resp, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
