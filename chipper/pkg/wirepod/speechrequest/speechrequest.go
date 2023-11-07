package speechrequest

import (
	"encoding/binary"
	"errors"
	"os"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/opus-go/opus"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vtt"
	"github.com/maxhawkins/go-webrtcvad"
)

// one type and many functions for dealing with intent, intent-graph, and knowledge-graph requests
// also some functions to help decode the stream bytes into ones friendly for stt engines

var debugWriteFile bool = false
var debugFile *os.File

type SpeechRequest struct {
	Device         string
	Session        string
	FirstReq       []byte
	Stream         interface{}
	IsKG           bool
	IsIG           bool
	MicData        []byte
	DecodedMicData []byte
	PrevLen        int
	PrevLenRaw     int
	InactiveFrames int
	ActiveFrames   int
	VADInst        *webrtcvad.VAD
	LastAudioChunk []byte
	IsOpus         bool
	OpusStream     *opus.OggStream
}

func BytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func (req *SpeechRequest) OpusDetect() bool {
	var isOpus bool
	if len(req.FirstReq) > 0 {
		if req.FirstReq[0] == 0x4f {
			logger.Println("Bot " + req.Device + " Stream type: OPUS")
			isOpus = true
		} else {
			isOpus = false
			logger.Println("Bot " + req.Device + " Stream type: PCM")
		}
	}
	return isOpus
}

func (req *SpeechRequest) OpusDecode(chunk []byte) []byte {
	if req.IsOpus {
		n, err := req.OpusStream.Decode(chunk)
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

// Uses VAD to detect when the user stops speaking
func (req *SpeechRequest) DetectEndOfSpeech() bool {
	// changes InactiveFrames and ActiveFrames in req
	inactiveNumMax := 23
	vad := req.VADInst
	vad.SetMode(3)
	for _, chunk := range SplitVAD(req.LastAudioChunk) {
		active, err := vad.Process(16000, chunk)
		if err != nil {
			logger.Println("VAD err:")
			logger.Println(err)
			return true
		}
		if active {
			req.ActiveFrames = req.ActiveFrames + 1
			req.InactiveFrames = 0
		} else {
			req.InactiveFrames = req.InactiveFrames + 1
		}
		if req.InactiveFrames >= inactiveNumMax && req.ActiveFrames > 18 {
			logger.Println("(Bot " + req.Device + ") End of speech detected.")
			return true
		}
	}
	return false
}

// Converts a vtt.*Request to a SpeechRequest, which allows functions like DetectEndOfSpeech to work
func ReqToSpeechRequest(req interface{}) SpeechRequest {
	if debugWriteFile {
		debugFile, _ = os.Create("/tmp/wirepodtest.ogg")
	}
	var request SpeechRequest
	request.PrevLen = 0
	var err error
	request.VADInst, err = webrtcvad.New()
	if err != nil {
		logger.Println(err)
	}
	if str, ok := req.(*vtt.IntentRequest); ok {
		var req1 *vtt.IntentRequest = str
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else if str, ok := req.(*vtt.KnowledgeGraphRequest); ok {
		var req1 *vtt.KnowledgeGraphRequest = str
		request.IsKG = true
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else if str, ok := req.(*vtt.IntentGraphRequest); ok {
		request.IsIG = true
		var req1 *vtt.IntentGraphRequest = str
		request.Device = req1.Device
		request.Session = req1.Session
		request.Stream = req1.Stream
		request.FirstReq = req1.FirstReq.InputAudio
		if debugWriteFile {
			debugFile.Write(req1.FirstReq.InputAudio)
		}
		request.MicData = append(request.MicData, req1.FirstReq.InputAudio...)
	} else {
		logger.Println("reqToSpeechRequest: invalid type")
	}
	isOpus := request.OpusDetect()
	if isOpus {
		request.OpusStream = &opus.OggStream{}
		decodedFirstReq, _ := request.OpusStream.Decode(request.FirstReq)
		request.FirstReq = decodedFirstReq
		request.DecodedMicData = append(request.DecodedMicData, decodedFirstReq...)
		request.LastAudioChunk = request.DecodedMicData[request.PrevLen:]
		request.PrevLen = len(request.DecodedMicData)
		request.IsOpus = true
	}
	return request
}

// Returns the next chunk in the stream as 16000 Hz PCM
func (req *SpeechRequest) GetNextStreamChunk() ([]byte, error) {
	// returns next chunk in voice stream as pcm
	if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		if debugWriteFile {
			debugFile.Write(chunk.InputAudio)
		}
		return dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingKnowledgeGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingKnowledgeGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		return dataReturn, nil
	}
	logger.Println("invalid type")
	return nil, errors.New("invalid type")
}

// Returns next chunk in the stream as whatever the original format is (OPUS 99% of the time)
func (req *SpeechRequest) GetNextStreamChunkOpus() ([]byte, error) {
	if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.MicData[req.PrevLenRaw:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		req.PrevLenRaw = len(req.MicData)
		return dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingIntentGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingIntentGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.MicData[req.PrevLenRaw:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		req.PrevLenRaw = len(req.MicData)
		return dataReturn, nil
	} else if str, ok := req.Stream.(pb.ChipperGrpc_StreamingKnowledgeGraphServer); ok {
		var stream pb.ChipperGrpc_StreamingKnowledgeGraphServer = str
		chunk, chunkErr := stream.Recv()
		if chunkErr != nil {
			logger.Println(chunkErr)
			return nil, chunkErr
		}
		req.MicData = append(req.MicData, chunk.InputAudio...)
		req.DecodedMicData = append(req.DecodedMicData, req.OpusDecode(chunk.InputAudio)...)
		dataReturn := req.MicData[req.PrevLenRaw:]
		req.LastAudioChunk = req.DecodedMicData[req.PrevLen:]
		req.PrevLen = len(req.DecodedMicData)
		req.PrevLenRaw = len(req.MicData)
		return dataReturn, nil
	}
	logger.Println("invalid type")
	return nil, errors.New("invalid type")
}
