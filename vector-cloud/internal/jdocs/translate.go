package jdocs

import (
	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	pb "github.com/digital-dream-labs/api/go/jdocspb"
)

type cladDoc cloud.Doc
type pbDoc pb.Jdoc

type cladWriteReq cloud.WriteRequest
type cladReadReq cloud.ReadRequest
type cladDeleteReq cloud.DeleteRequest

type protoWriteResp pb.WriteDocResp
type protoWriteStatus pb.WriteDocResp_Status
type protoReadResp pb.ReadDocsResp
type protoReadStatus pb.ReadDocsResp_Status

func (c *cladDoc) toProto() *pb.Jdoc {
	return &pb.Jdoc{
		DocVersion:     c.DocVersion,
		FmtVersion:     c.FmtVersion,
		ClientMetadata: c.Metadata,
		JsonDoc:        c.JsonDoc,
	}
}

func (p *pbDoc) toClad() *cloud.Doc {
	return &cloud.Doc{
		DocVersion: p.DocVersion,
		FmtVersion: p.FmtVersion,
		Metadata:   p.ClientMetadata,
		JsonDoc:    p.JsonDoc,
	}
}

func (c *cladWriteReq) toProto() *pb.WriteDocReq {
	return &pb.WriteDocReq{
		UserId:  c.Account,
		Thing:   c.Thing,
		DocName: c.DocName,
		Doc:     (*cladDoc)(&c.Doc).toProto(),
	}
}

func (c *cladReadReq) toProto() *pb.ReadDocsReq {
	ret := &pb.ReadDocsReq{
		UserId: c.Account,
		Thing:  c.Thing,
	}
	ret.Items = make([]*pb.ReadDocsReq_Item, len(c.Items))
	for i, c := range c.Items {
		ret.Items[i] = &pb.ReadDocsReq_Item{
			DocName:      c.DocName,
			MyDocVersion: c.MyDocVersion,
		}
	}
	return ret
}

func (c *cladDeleteReq) toProto() *pb.DeleteDocReq {
	return &pb.DeleteDocReq{
		UserId:  c.Account,
		Thing:   c.Thing,
		DocName: c.DocName,
	}
}

func (p *protoWriteResp) toClad() *cloud.WriteResponse {
	return &cloud.WriteResponse{
		LatestVersion: p.LatestDocVersion,
		Status:        writeStatusMap[p.Status],
	}
}

func (p *protoReadResp) toClad() *cloud.ReadResponse {
	ret := &cloud.ReadResponse{}
	ret.Items = make([]cloud.ResponseDoc, len(p.Items))
	for i, p := range p.Items {
		ret.Items[i] = cloud.ResponseDoc{
			Status: readStatusMap[p.Status],
			Doc:    *(*pbDoc)(p.Doc).toClad(),
		}
	}
	return ret
}

var readStatusMap = map[pb.ReadDocsResp_Status]cloud.ReadStatus{
	pb.ReadDocsResp_UNCHANGED: cloud.ReadStatus_Unchanged,
	pb.ReadDocsResp_CHANGED:   cloud.ReadStatus_Changed,
	pb.ReadDocsResp_NOT_FOUND: cloud.ReadStatus_NotFound,
}

var writeStatusMap = map[pb.WriteDocResp_Status]cloud.WriteStatus{
	pb.WriteDocResp_ACCEPTED:                 cloud.WriteStatus_Accepted,
	pb.WriteDocResp_REJECTED_BAD_DOC_VERSION: cloud.WriteStatus_RejectedDocVersion,
	pb.WriteDocResp_REJECTED_BAD_FMT_VERSION: cloud.WriteStatus_RejectedFmtVersion,
}
