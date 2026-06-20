package wirepod_groq

// Groq cloud Speech-to-Text engine for wire-pod.
//
// This implements the same engine interface as the built-in VOSK/Whisper
// modules (an exported Name string, an Init() error, and an
// STT(sr.SpeechRequest) (string, error) function), but instead of running a
// model locally it streams the captured audio to Groq's OpenAI-compatible
// transcription endpoint, which runs whisper-large-v3 on Groq's hardware.
//
// This is ideal for low-power hosts (e.g. a Raspberry Pi 3B+) that cannot run
// a large local model: the Pi only captures and forwards audio, while Groq
// does the recognition. It also handles accented speech and non-English
// languages (e.g. Greek) far better than the small local VOSK model.
//
// Configuration is via environment variables (set these in
// chipper/source.sh so start.sh picks them up):
//
//	GROQ_API_KEY       (required) Your Groq API key.
//	GROQ_STT_MODEL     (optional) Transcription model.
//	                              Default: "whisper-large-v3" (most accurate, multilingual).
//	                              Alternatives: "whisper-large-v3-turbo" (faster, slightly
//	                              less accurate), "distil-whisper-large-v3-en" (English-only, fastest).
//	GROQ_STT_LANGUAGE  (optional) ISO-639-1 language code to force, e.g. "en" or "el".
//	                              Leave unset for automatic language detection (good for
//	                              switching between English and Greek). Forcing a language
//	                              improves accuracy/latency on short commands.
//	GROQ_STT_PROMPT    (optional) A short biasing prompt passed to Whisper. Useful to nudge
//	                              recognition toward expected vocabulary, e.g.
//	                              "Hey Vector. Commands: set timer, cancel timer, what time is it".
//	GROQ_API_URL       (optional) Override the endpoint. Defaults to Groq's transcription URL.
//	                              You can point this at OpenAI or any OpenAI-compatible
//	                              transcription endpoint instead.
//
// Because Groq's API is OpenAI-compatible, the request shape here matches the
// standard /audio/transcriptions multipart form.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	"github.com/orcaman/writerseeker"
)

// Name is the value wire-pod uses to identify this engine (STT_SERVICE=groq).
var Name string = "groq"

const (
	defaultURL   = "https://api.groq.com/openai/v1/audio/transcriptions"
	defaultModel = "whisper-large-v3"
	// httpTimeout bounds a single transcription request. Groq transcribes a
	// short utterance in well under a second, so this is generous headroom.
	httpTimeout = 30 * time.Second
	// maxRetries is how many additional attempts to make on a transient
	// failure (network error / HTTP 429 / HTTP 5xx). Kept small so the user
	// isn't left waiting: a voice assistant should fail fast rather than hang.
	maxRetries = 2
)

// groqResp models the JSON returned for response_format=json. On success the
// transcript is in Text; on failure Groq/OpenAI return an Error object.
type groqResp struct {
	Text  string `json:"text"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func apiURL() string {
	if v := strings.TrimSpace(os.Getenv("GROQ_API_URL")); v != "" {
		return v
	}
	return defaultURL
}

func model() string {
	if v := strings.TrimSpace(os.Getenv("GROQ_STT_MODEL")); v != "" {
		return v
	}
	return defaultModel
}

// Init validates configuration and logs the active settings. It intentionally
// does NOT exit on a missing key (mirroring the built-in Whisper module) so
// that wire-pod can still boot and present its web setup page.
func Init() error {
	if strings.TrimSpace(os.Getenv("GROQ_API_KEY")) == "" {
		logger.Println("[groq-stt] WARNING: GROQ_API_KEY is not set. " +
			"Set it in chipper/source.sh (export GROQ_API_KEY=\"...\") or transcription will fail.")
	}
	lang := strings.TrimSpace(os.Getenv("GROQ_STT_LANGUAGE"))
	if lang == "" {
		lang = "auto-detect"
	}
	logger.Println(fmt.Sprintf("[groq-stt] initialized (model=%s, language=%s, url=%s)",
		model(), lang, apiURL()))
	return nil
}

// newAudioIntBuffer reads little-endian signed 16-bit PCM from r into an
// IntBuffer suitable for the WAV encoder.
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

// pcm2wav wraps raw 16 kHz / 16-bit / mono PCM in a WAV container so the
// transcription endpoint receives a self-describing audio file.
func pcm2wav(in io.Reader) ([]byte, error) {
	out := &writerseeker.WriterSeeker{}
	// 16 kHz, 16-bit, 1 channel, PCM (WAV audio format 1).
	enc := wav.NewEncoder(out, 16000, 16, 1, 1)

	audioBuf, err := newAudioIntBuffer(in)
	if err != nil {
		return nil, fmt.Errorf("decoding pcm: %w", err)
	}
	if err := enc.Write(audioBuf); err != nil {
		return nil, fmt.Errorf("encoding wav: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("finalizing wav: %w", err)
	}
	outBuf := new(bytes.Buffer)
	if _, err := io.Copy(outBuf, out.BytesReader()); err != nil {
		return nil, fmt.Errorf("buffering wav: %w", err)
	}
	return outBuf.Bytes(), nil
}

// buildRequest constructs a fresh multipart transcription request. It is built
// per-attempt because an *http.Request body is not reusable across retries.
func buildRequest(wavData []byte) (*http.Request, error) {
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)

	if err := w.WriteField("model", model()); err != nil {
		return nil, err
	}
	// response_format=json gives us {"text": "..."} and lets us surface
	// structured errors.
	if err := w.WriteField("response_format", "json"); err != nil {
		return nil, err
	}
	// temperature=0 makes recognition deterministic (best for commands).
	if err := w.WriteField("temperature", "0"); err != nil {
		return nil, err
	}
	if lang := strings.TrimSpace(os.Getenv("GROQ_STT_LANGUAGE")); lang != "" {
		if err := w.WriteField("language", lang); err != nil {
			return nil, err
		}
	}
	if prompt := strings.TrimSpace(os.Getenv("GROQ_STT_PROMPT")); prompt != "" {
		if err := w.WriteField("prompt", prompt); err != nil {
			return nil, err
		}
	}

	// The filename's extension tells the server the container format.
	part, err := w.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(wavData); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(os.Getenv("GROQ_API_KEY")))
	return req, nil
}

// transcribe sends the WAV to Groq and returns the recognized text. It retries
// a small number of times on transient failures (network errors, HTTP 429, and
// HTTP 5xx) with a short backoff, and fails fast on client errors (e.g. a bad
// API key) so problems are surfaced immediately in the logs.
func transcribe(wavData []byte) (string, error) {
	client := &http.Client{Timeout: httpTimeout}
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 500 * time.Millisecond
			logger.Println(fmt.Sprintf("[groq-stt] retry %d/%d after %s (%v)",
				attempt, maxRetries, backoff, lastErr))
			time.Sleep(backoff)
		}

		req, err := buildRequest(wavData)
		if err != nil {
			// Construction errors are not transient; stop.
			return "", fmt.Errorf("building request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue // network error: retry
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("reading response: %w", readErr)
			continue
		}

		// Retry on rate limiting and server errors; fail fast otherwise.
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server returned HTTP %d: %s",
				resp.StatusCode, strings.TrimSpace(string(respBody)))
			continue
		}

		var parsed groqResp
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			// Unparseable body on a non-OK status is still a hard error.
			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("HTTP %d and unparseable response: %s",
					resp.StatusCode, strings.TrimSpace(string(respBody)))
			}
			return "", fmt.Errorf("parsing response: %w", err)
		}

		if parsed.Error != nil {
			// Structured API error (e.g. invalid key, bad model): do not retry.
			return "", fmt.Errorf("groq api error: %s", parsed.Error.Message)
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d: %s",
				resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		return parsed.Text, nil
	}

	if lastErr == nil {
		lastErr = errors.New("transcription failed for an unknown reason")
	}
	return "", fmt.Errorf("groq transcription failed after %d attempts: %w",
		maxRetries+1, lastErr)
}

// STT satisfies the wire-pod engine interface. It drains the bot's audio
// stream until end-of-speech is detected, packages the captured PCM as WAV,
// sends it to Groq, and returns the lowercased transcript for intent matching.
func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Groq) Processing...")

	for {
		if _, err := req.GetNextStreamChunk(); err != nil {
			return "", err
		}
		done, _ := req.DetectEndOfSpeech()
		if done {
			break
		}
	}

	pcmBuf := &writerseeker.WriterSeeker{}
	if _, err := pcmBuf.Write(req.DecodedMicData); err != nil {
		return "", fmt.Errorf("buffering mic data: %w", err)
	}
	wavData, err := pcm2wav(pcmBuf.BytesReader())
	if err != nil {
		return "", err
	}

	text, err := transcribe(wavData)
	if err != nil {
		logger.Println("[groq-stt] " + err.Error())
		return "", err
	}

	transcribedText := strings.ToLower(strings.TrimSpace(text))
	transcribedText = strings.TrimRight(transcribedText, ".,!?;:")
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
