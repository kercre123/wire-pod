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
	if os.Getenv("OPENAI_KEY") == "" && os.Getenv("STT_HOST") == "" {
		logger.Println("This is an early implementation of the Whisper API which has not been implemented into the web interface. You must set the OPENAI_KEY env var, OR set STT_HOST to point to a custom server.")
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
	// Check for custom host
	host := os.Getenv("STT_HOST")
	var url string
	if host != "" {
		// Use custom host. Assume it's the full base URL or handle appending.
		// Let's assume STT_HOST is the FULL URL (e.g. http://1.2.3.4:8001/v1/audio/transcriptions)
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		url = host
		// If it looks like just a base (no 'audio/transcriptions'), append it?
		if !strings.Contains(url, "/transcriptions") {
			if strings.HasSuffix(url, "/") {
				url += "v1/audio/transcriptions"
			} else {
				url += "/v1/audio/transcriptions"
			}
		}
	} else {
		url = "https://api.openai.com/v1/audio/transcriptions"
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("model", "whisper-1")
	sendFile, _ := w.CreateFormFile("file", "audio.wav")
	sendFile.Write(in)
	w.Close()

	httpReq, _ := http.NewRequest("POST", url, buf)
	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	// Use dummy key if not set, or forwarded key
	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		key = "dummy"
	}
	httpReq.Header.Set("Authorization", "Bearer "+key)

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

	// Fallback/Debugging
	if aiResponse.Text == "" {
		logger.Println("Whisper response empty. Raw response: " + string(response))
	}

	return aiResponse.Text
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Whisper) Processing...")
	speechIsDone := false
	var err error
	for {
		_, err = req.GetNextStreamChunk()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		} // Why double ?
		if err != nil {
			return "", err
		}
		// has to be split into 320 []byte chunks for VAD
		speechIsDone, _ = req.DetectEndOfSpeech()
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
