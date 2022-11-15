package wirepod

import (
	"errors"
	"strconv"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/digital-dream-labs/opus-go/opus"
)

type SpeechRequest struct {
	// bot esn
	Device string
	// seemingly random string
	Session string
	// first bytes in voice stream
	FirstReq []byte
	// rest is behind-the-scenes stuff
	Stream         interface{}
	MicData        []byte
	DecodedMicData []byte
	PrevLen        int
	IsOpus         bool
	OpusStream     opus.OggStream
	BotNum         int
}

func opusDetect(req SpeechRequest) bool {
	var isOpus bool
	if len(req.FirstReq) > 0 {
		if req.FirstReq[0] == 0x4f {
			logger("Bot " + strconv.Itoa(req.BotNum) + " Stream type: OPUS")
			isOpus = true
		} else {
			isOpus = false
			logger("Bot " + strconv.Itoa(req.BotNum) + " Stream type: PCM")
		}
	}
	return isOpus
}

func opusDecode(req SpeechRequest) []byte {
	if req.IsOpus {
		n, err := req.OpusStream.Decode(req.MicData)
		if err != nil {
			logger(err)
		}
		return n
	} else {
		return req.MicData
	}
}

func reqToSpeechRequest(req interface{}) SpeechRequest {
	var request SpeechRequest
	request.BotNum = botNum
	if str, ok := req.(*vtt.IntentRequest); ok {
		var req1 *vtt.IntentRequest = str
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else if str, ok := req.(*vtt.KnowledgeGraphRequest); ok {
		var req1 *vtt.KnowledgeGraphRequest = str
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else if str, ok := req.(*vtt.IntentGraphRequest); ok {
		var req1 *vtt.IntentGraphRequest = str
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else {
		logger("reqToSpeechRequest: invalid type")
	}
	isOpus := opusDetect(request)
	if isOpus {
		request.OpusStream = opus.OggStream{}
		request.IsOpus = true
	}
	return request
}

func getNextStreamChunk(req SpeechRequest) (SpeechRequest, []byte, error) {
	// returns next chunk in voice stream as pcm
	if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = opusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = opusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingKnowledgeGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingKnowledgeGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = opusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	}
	logger("invalid type")
	return req, nil, errors.New("invalid type")
}
