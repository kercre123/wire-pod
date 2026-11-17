// Package wirepod_wyoming is a speech-to-text provider that sends the robot's
// audio to an external server over the Wyoming protocol (the one Home Assistant
// uses): wyoming-faster-whisper, wyoming-vosk, GigaAM via onnx-asr and so on.
//
// The server address, model and language are set in the web interface
// (Server Settings -> STT) and stored in apiConfig.json. The WYOMING_STT_*
// environment variables are used as a fallback, so the provider can also be
// configured from the systemd unit.
package wirepod_wyoming

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
)

var Name string = "wyoming"

const (
	DefaultPort    = "10300"
	DefaultTimeout = 15 * time.Second

	sampleRate  = 16000
	sampleWidth = 2
	channels    = 1
	// one second of 16 kHz 16-bit mono audio per audio-chunk event
	chunkBytes = 32000
	// a sane upper bound for a single event's data or payload from the server
	maxEventBytes = 16 << 20
)

// Settings is the effective provider configuration.
type Settings struct {
	Address  string
	Model    string
	Language string
	Timeout  time.Duration
}

// CurrentSettings reads apiConfig.json, falling back to the environment.
func CurrentSettings() Settings {
	c := vars.Config().STT
	s := Settings{
		Address:  firstNonEmpty(c.WyomingAddress, os.Getenv("WYOMING_STT_ADDRESS")),
		Model:    firstNonEmpty(c.WyomingModel, os.Getenv("WYOMING_STT_MODEL")),
		Language: firstNonEmpty(c.WyomingLanguage, os.Getenv("WYOMING_STT_LANGUAGE")),
		Timeout:  DefaultTimeout,
	}
	if s.Language == "" {
		// intents are picked by STT.Language ("ru-RU"), the server wants "ru"
		s.Language = strings.Split(c.Language, "-")[0]
	}
	if c.WyomingTimeout > 0 {
		s.Timeout = time.Duration(c.WyomingTimeout) * time.Second
	} else if v, err := strconv.Atoi(os.Getenv("WYOMING_STT_TIMEOUT")); err == nil && v > 0 {
		s.Timeout = time.Duration(v) * time.Second
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

// NormalizeAddress accepts "host:port", "tcp://host:port" or a bare host
// (port 10300 is assumed) and returns "host:port".
func NormalizeAddress(addr string) (string, error) {
	addr = strings.TrimPrefix(strings.TrimSpace(addr), "tcp://")
	if addr == "" {
		return "", errors.New("Wyoming server address is not set")
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(strings.Trim(addr, "[]"), DefaultPort)
		if _, _, err := net.SplitHostPort(addr); err != nil {
			return "", fmt.Errorf("invalid Wyoming server address %q", addr)
		}
	}
	return addr, nil
}

func Init() error {
	s := CurrentSettings()
	addr, err := NormalizeAddress(s.Address)
	if err != nil {
		logger.Println("Wyoming STT: " + err.Error() + ". Set it in Server Settings -> STT or with WYOMING_STT_ADDRESS.")
		// wire-pod must still start, the address can be set from the web interface
		return nil
	}
	programs, err := Describe(addr, 5*time.Second)
	if err != nil {
		logger.Println("Wyoming STT: server " + addr + " is not reachable right now: " + err.Error())
		return nil
	}
	model := s.Model
	if model == "" {
		model = "server default"
	}
	logger.Println("Wyoming STT: " + addr + " (" + Summary(programs) + "), model: " + model + ", language: " + s.Language)
	return nil
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Wyoming) Processing...")
	for {
		if _, err := req.GetNextStreamChunk(); err != nil {
			return "", err
		}
		if done, _ := req.DetectEndOfSpeech(); done {
			break
		}
	}
	start := time.Now()
	text, err := Transcribe(req.DecodedMicData, CurrentSettings())
	if err != nil {
		logger.Println("(Bot " + req.Device + ", Wyoming) " + err.Error())
		return "", err
	}
	text = strings.ToLower(strings.TrimSpace(text))
	logger.Println("(Bot " + req.Device + ", Wyoming) Recognized in " + time.Since(start).Round(time.Millisecond).String())
	logger.Println("Bot " + req.Device + " Transcribed text: " + text)
	return text, nil
}

// Transcribe sends 16 kHz 16-bit mono PCM to the server and returns the text.
func Transcribe(pcm []byte, s Settings) (string, error) {
	addr, err := NormalizeAddress(s.Address)
	if err != nil {
		return "", err
	}
	conn, err := net.DialTimeout("tcp", addr, s.Timeout)
	if err != nil {
		return "", fmt.Errorf("connect to Wyoming server %s: %w", addr, err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(s.Timeout))

	w := bufio.NewWriter(conn)
	transcribe := map[string]any{}
	if s.Language != "" {
		transcribe["language"] = s.Language
	}
	if s.Model != "" {
		transcribe["name"] = s.Model
	}
	format := map[string]any{"rate": sampleRate, "width": sampleWidth, "channels": channels}
	if err := writeEvent(w, "transcribe", transcribe, nil); err != nil {
		return "", err
	}
	if err := writeEvent(w, "audio-start", format, nil); err != nil {
		return "", err
	}
	for i := 0; i < len(pcm); i += chunkBytes {
		if err := writeEvent(w, "audio-chunk", format, pcm[i:min(i+chunkBytes, len(pcm))]); err != nil {
			return "", err
		}
	}
	if err := writeEvent(w, "audio-stop", nil, nil); err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("send audio to Wyoming server %s: %w", addr, err)
	}

	r := bufio.NewReader(conn)
	for {
		ev, err := readEvent(r)
		if err != nil {
			return "", fmt.Errorf("read from Wyoming server %s: %w", addr, err)
		}
		switch ev.Type {
		case "transcript":
			text, _ := ev.Data["text"].(string)
			return text, nil
		case "error":
			msg, _ := ev.Data["text"].(string)
			return "", fmt.Errorf("Wyoming server %s returned an error: %s", addr, msg)
		}
		// anything else (transcript-start/-chunk of streaming servers) is skipped
	}
}

// ModelInfo and ProgramInfo describe the speech-to-text part of a server's
// "info" event; the web interface shows them after "Test connection".
type ModelInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Languages   []string `json:"languages"`
}

type ProgramInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Models      []ModelInfo `json:"models"`
}

// Describe asks the server what speech-to-text programs and models it has.
func Describe(address string, timeout time.Duration) ([]ProgramInfo, error) {
	addr, err := NormalizeAddress(address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))

	if err := writeEvent(conn, "describe", nil, nil); err != nil {
		return nil, err
	}
	r := bufio.NewReader(conn)
	for {
		ev, err := readEvent(r)
		if err != nil {
			return nil, fmt.Errorf("read from %s: %w", addr, err)
		}
		if ev.Type != "info" {
			continue
		}
		raw, _ := json.Marshal(ev.Data["asr"])
		var programs []ProgramInfo
		json.Unmarshal(raw, &programs)
		if len(programs) == 0 {
			return nil, fmt.Errorf("%s answers, but has no speech-to-text service (is it a TTS port?)", addr)
		}
		return programs, nil
	}
}

// Summary is a one-line description of the server's programs and models.
func Summary(programs []ProgramInfo) string {
	var parts []string
	for _, p := range programs {
		var models []string
		for _, m := range p.Models {
			models = append(models, m.Name)
		}
		parts = append(parts, p.Name+": "+strings.Join(models, ", "))
	}
	return strings.Join(parts, "; ")
}

type event struct {
	Type    string
	Data    map[string]any
	Payload []byte
}

// writeEvent writes a header line, then the JSON data and the payload as
// separate blocks whose sizes are given in the header. Servers also accept
// data inline in the header, but this is the canonical form.
func writeEvent(w io.Writer, eventType string, data any, payload []byte) error {
	header := map[string]any{"type": eventType}
	var dataBytes []byte
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		dataBytes = b
		header["data_length"] = len(dataBytes)
	}
	if len(payload) > 0 {
		header["payload_length"] = len(payload)
	}
	h, err := json.Marshal(header)
	if err != nil {
		return err
	}
	buf := make([]byte, 0, len(h)+1+len(dataBytes)+len(payload))
	buf = append(buf, h...)
	buf = append(buf, '\n')
	buf = append(buf, dataBytes...)
	buf = append(buf, payload...)
	_, err = w.Write(buf)
	return err
}

func readEvent(r *bufio.Reader) (event, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return event{}, err
	}
	var header struct {
		Type          string         `json:"type"`
		Data          map[string]any `json:"data"`
		DataLength    int            `json:"data_length"`
		PayloadLength int            `json:"payload_length"`
	}
	if err := json.Unmarshal(line, &header); err != nil {
		return event{}, fmt.Errorf("bad event header: %w", err)
	}
	if header.DataLength > maxEventBytes || header.PayloadLength > maxEventBytes {
		return event{}, errors.New("event is too large")
	}
	ev := event{Type: header.Type, Data: header.Data}
	if ev.Data == nil {
		ev.Data = map[string]any{}
	}
	if header.DataLength > 0 {
		b := make([]byte, header.DataLength)
		if _, err := io.ReadFull(r, b); err != nil {
			return event{}, err
		}
		var extra map[string]any
		if err := json.Unmarshal(b, &extra); err != nil {
			return event{}, fmt.Errorf("bad event data: %w", err)
		}
		for k, v := range extra {
			ev.Data[k] = v
		}
	}
	if header.PayloadLength > 0 {
		ev.Payload = make([]byte, header.PayloadLength)
		if _, err := io.ReadFull(r, ev.Payload); err != nil {
			return event{}, err
		}
	}
	return ev, nil
}
