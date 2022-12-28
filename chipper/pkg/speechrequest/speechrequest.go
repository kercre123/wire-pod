package speechrequest

import (
	"encoding/binary"
	"errors"
	"os"
	"strconv"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/logger"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/digital-dream-labs/opus-go/opus"
	"github.com/maxhawkins/go-webrtcvad"
)

// one type and many functions for dealing with intent, intent-graph, and knowledge-graph requests
// also some functions to help decode the stream bytes into ones friendly for stt engines

var BotNum = 0

// need a good place for everywhere to be able to see this without causing import cycle
var SttLanguage = "en-US"

type SpeechRequest struct {
	Device         string
	Session        string
	FirstReq       []byte
	Stream         interface{}
	MicData        []byte
	DecodedMicData []byte
	PrevLen        int
	InactiveFrames int
	ActiveFrames   int
	VADInst        *webrtcvad.VAD
	LastAudioChunk []byte
	IsOpus         bool
	OpusStream     opus.OggStream
	BotNum         int
}

func InitLanguage() {
	SttLanguage = os.Getenv("STT_LANGUAGE")
	if len(SttLanguage) == 0 {
		SttLanguage = "en-US"
	}
}

func BytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func OpusDetect(req SpeechRequest) bool {
	var isOpus bool
	if len(req.FirstReq) > 0 {
		if req.FirstReq[0] == 0x4f {
			logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Stream type: OPUS")
			isOpus = true
		} else {
			isOpus = false
			logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Stream type: PCM")
		}
	}
	return isOpus
}

func OpusDecode(req SpeechRequest) []byte {
	if req.IsOpus {
		n, err := req.OpusStream.Decode(req.MicData)
		if err != nil {
			logger.Println(err)
		}
		return n
	} else {
		return req.MicData
	}
}

func SplitVAD(buf []byte) [][]byte {
	var chunk [][]byte
	for len(buf) >= 320 {
		chunk = append(chunk, buf[:320])
		buf = buf[320:]
	}
	return chunk
}

func BytesToIntVAD(stream opus.OggStream, data []byte, die bool, isOpus bool) [][]byte {
	// detect if data is pcm or opus
	if die {
		return nil
	}
	if isOpus {
		// opus
		n, err := stream.Decode(data)
		if err != nil {
			logger.Println(err)
		}
		byteArray := SplitVAD(n)
		return byteArray
	} else {
		// pcm
		byteArray := SplitVAD(data)
		return byteArray
	}
}

func DetectEndOfSpeech(req SpeechRequest) (SpeechRequest, bool) {
	// changes InactiveFrames and ActiveFrames in req
	inactiveNumMax := 20
	vad := req.VADInst
	vad.SetMode(3)
	for _, chunk := range SplitVAD(req.LastAudioChunk) {
		active, err := vad.Process(16000, chunk)
		if err != nil {
			logger.Println("VAD err:")
			logger.Println(err)
			return req, true
		}
		if active {
			req.ActiveFrames = req.ActiveFrames + 1
			req.InactiveFrames = 0
		} else {
			req.InactiveFrames = req.InactiveFrames + 1
		}
		if req.InactiveFrames >= inactiveNumMax && req.ActiveFrames > 20 {
			logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ") End of speech detected.")
			return req, true
		}
	}
	return req, false
}

func ReqToSpeechRequest(req interface{}) SpeechRequest {
	var request SpeechRequest
	var err error
	request.VADInst, err = webrtcvad.New()
	if err != nil {
		logger.Println(err)
	}
	request.BotNum = BotNum
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
		logger.Println("reqToSpeechRequest: invalid type")
	}
	isOpus := OpusDetect(request)
	if isOpus {
		request.OpusStream = opus.OggStream{}
		request.IsOpus = true
	}
	return request
}

func GetNextStreamChunk(req SpeechRequest) (SpeechRequest, []byte, error) {
	// returns next chunk in voice stream as pcm
	if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = OpusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = OpusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingKnowledgeGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingKnowledgeGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return req, nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = OpusDecode(req)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return req, dataReturn, nil
	}
	logger.Println("invalid type")
	return req, nil, errors.New("invalid type")
}
