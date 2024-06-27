package speechrequest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/opus-go/opus"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vtt"
	"github.com/maxhawkins/go-webrtcvad"
)

// one type and many functions for dealing with intent, intent-graph, and knowledge-graph requests
// also some functions to help decode the stream bytes into ones friendly for stt engines

var debugWriteFile bool = false
var debugFile *os.File

type SpeechRequest struct {
	Device          string
	Session         string
	FirstReq        []byte
	Stream          interface{}
	IsKG            bool
	IsIG            bool
	MicData         []byte
	DecodedMicData  []byte
	FilteredMicData []byte
	PrevLen         int
	PrevLenRaw      int
	InactiveFrames  int
	ActiveFrames    int
	VADInst         *webrtcvad.VAD
	LastAudioChunk  []byte
	IsOpus          bool
	OpusStream      *opus.OggStream
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
		return chunk
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
func (req *SpeechRequest) DetectEndOfSpeech() (bool, bool) {
	// changes InactiveFrames and ActiveFrames in req
	inactiveNumMax := 23
	for _, chunk := range SplitVAD(req.LastAudioChunk) {
		active, err := req.VADInst.Process(16000, chunk)
		if err != nil {
			logger.Println("VAD err:")
			logger.Println(err)
			return true, false
		}
		if active {
			req.ActiveFrames = req.ActiveFrames + 1
			req.InactiveFrames = 0
		} else {
			req.InactiveFrames = req.InactiveFrames + 1
		}
		if req.InactiveFrames >= inactiveNumMax && req.ActiveFrames > 18 {
			logger.Println("(Bot " + req.Device + ") End of speech detected.")
			return true, true
		}
	}
	if req.ActiveFrames < 5 {
		return false, false
	}
	return false, true
}

func bytesToInt16(data []byte) ([]int16, error) {
	var samples []int16
	buf := bytes.NewReader(data)
	for buf.Len() > 0 {
		var sample int16
		err := binary.Read(buf, binary.LittleEndian, &sample)
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func int16ToBytes(samples []int16) []byte {
	buf := new(bytes.Buffer)
	for _, sample := range samples {
		err := binary.Write(buf, binary.LittleEndian, sample)
		if err != nil {
			return nil
		}
	}
	return buf.Bytes()
}

func applyGain(samples []int16, gain float64) []int16 {
	for i, sample := range samples {
		amplifiedSample := float64(sample) * gain
		if amplifiedSample > math.MaxInt16 {
			samples[i] = math.MaxInt16
		} else if amplifiedSample < math.MinInt16 {
			samples[i] = math.MinInt16
		} else {
			samples[i] = int16(amplifiedSample)
		}
	}
	return samples
}

// remove noise
func highPassFilter(data []byte) []byte {
	bTime := time.Now()
	sampleRate := 16000
	cutoffFreq := 300.0
	samples, err := bytesToInt16(data)
	if err != nil {
		return nil
	}
	samples = applyGain(samples, 5)
	filteredSamples := make([]float64, len(samples))
	rc := 1.0 / (2.0 * math.Pi * cutoffFreq)
	dt := 1.0 / float64(sampleRate)
	alpha := dt / (rc + dt)

	previous := float64(samples[0])
	for i := 1; i < len(samples); i++ {
		current := float64(samples[i])
		filtered := alpha * (filteredSamples[i-1] + current - previous)
		filteredSamples[i] = filtered
		previous = current
	}
	int16FilteredSamples := make([]int16, len(filteredSamples))
	for i, sample := range filteredSamples {
		int16FilteredSamples[i] = int16(sample)
	}

	gained := applyGain(int16FilteredSamples, 1.5)
	if os.Getenv("DEBUG_PRINT_HIGHPASS") == "true" {
		logger.Println("highpass filter took: " + fmt.Sprint(time.Since(bTime)))
	}

	return int16ToBytes(gained)
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
	request.VADInst.SetMode(2)
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
		request.FirstReq = highPassFilter(decodedFirstReq)
		request.FilteredMicData = append(request.FilteredMicData, request.FirstReq...)
		request.DecodedMicData = append(request.DecodedMicData, decodedFirstReq...)
		request.LastAudioChunk = request.FilteredMicData[request.PrevLen:]
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
		req.FilteredMicData = append(req.FilteredMicData, highPassFilter(req.OpusDecode(chunk.InputAudio))...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.FilteredMicData[req.PrevLen:]
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
		req.FilteredMicData = append(req.FilteredMicData, highPassFilter(req.OpusDecode(chunk.InputAudio))...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.FilteredMicData[req.PrevLen:]
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
		req.FilteredMicData = append(req.FilteredMicData, highPassFilter(req.OpusDecode(chunk.InputAudio))...)
		dataReturn := req.DecodedMicData[req.PrevLen:]
		req.LastAudioChunk = req.FilteredMicData[req.PrevLen:]
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
