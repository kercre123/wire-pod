package wirepod_whisper

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	"github.com/orcaman/writerseeker"
)

var Name string = "whisper"

type openAiResp struct {
	Text string `json:"text"`
}

func Init() error {
	if os.Getenv("OPENAI_KEY") == "" {
		logger.Println("This is an early implementation of the Whisper API which has not been implemented into the web interface. You must set the OPENAI_KEY env var.")
		//os.Exit(1)
	}
	return nil
}

func pcm2wav(in io.Reader) []byte {

	// Output file.
	out := &writerseeker.WriterSeeker{}

	// 8 kHz, 16 bit, 1 channel, WAV.
	e := wav.NewEncoder(out, 16000, 16, 1, 1)

	// Create new audio.IntBuffer.
	audioBuf, err := newAudioIntBuffer(in)
	if err != nil {
		logger.Println(err)
	}
	// Write buffer to output file. This writes a RIFF header and the PCM chunks from the audio.IntBuffer.
	if err := e.Write(audioBuf); err != nil {
		logger.Println(err)
	}
	if err := e.Close(); err != nil {
		logger.Println(err)
	}
	outBuf := new(bytes.Buffer)
	io.Copy(outBuf, out.BytesReader())
	return outBuf.Bytes()
}

func newAudioIntBuffer(r io.Reader) (*audio.IntBuffer, error) {
	buf := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  16000,
		},
	}
	for {
		var sample int16
		err := binary.Read(r, binary.LittleEndian, &sample)
		switch {
		case err == io.EOF:
			return &buf, nil
		case err != nil:
			return nil, err
		}
		buf.Data = append(buf.Data, int(sample))
	}
}

func makeOpenAIReq(in []byte) string {
	url := "https://api.openai.com/v1/audio/transcriptions"

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("model", "whisper-1")
	sendFile, _ := w.CreateFormFile("file", "audio.mp3")
	sendFile.Write(in)
	w.Close()

	httpReq, _ := http.NewRequest("POST", url, buf)
	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_KEY"))

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Println(err)
		return "There was an error."
	}

	defer resp.Body.Close()

	response, _ := io.ReadAll(resp.Body)

	var aiResponse openAiResp
	json.Unmarshal(response, &aiResponse)

	return aiResponse.Text
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Whisper) Processing...")
	speechIsDone := false
	var err error
	for {
		_, err = req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		if err != nil {
			return "", err
		}
		// has to be split into 320 []byte chunks for VAD
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}

	pcmBufTo := &writerseeker.WriterSeeker{}
	pcmBufTo.Write(req.DecodedMicData)
	pcmBuf := pcm2wav(pcmBufTo.BytesReader())

	transcribedText := strings.ToLower(makeOpenAIReq(pcmBuf))
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
