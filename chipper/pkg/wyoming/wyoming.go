// Package wyoming implements a minimal client for the Wyoming protocol
// (https://github.com/OHF-Voice/wyoming), used by Home Assistant voice
// services such as wyoming-faster-whisper (speech-to-text) and wyoming-piper
// (text-to-speech).
//
// Every event is a single JSON header line, optionally followed by a JSON data
// block (data_length bytes) and a binary payload (payload_length bytes).
package wyoming

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const protocolVersion = "1.5.0"

// DefaultTimeout bounds a whole request (connect, send audio/text, read reply).
var DefaultTimeout = 30 * time.Second

// DialTimeout bounds only the TCP connect, so an unreachable server is detected
// quickly and the caller can fall back to another engine.
var DialTimeout = 3 * time.Second

// Event is a single Wyoming protocol message.
type Event struct {
	Type    string
	Data    map[string]interface{}
	Payload []byte
}

type header struct {
	Type          string                 `json:"type"`
	Version       string                 `json:"version,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	DataLength    int                    `json:"data_length,omitempty"`
	PayloadLength int                    `json:"payload_length,omitempty"`
}

// Conn is a connection to a Wyoming server.
type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Address joins host and port into a dialable address.
func Address(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Dial connects to a Wyoming server and sets an overall deadline for the request.
func Dial(addr string, timeout time.Duration) (*Conn, error) {
	c, err := net.DialTimeout("tcp", addr, DialTimeout)
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		c.SetDeadline(time.Now().Add(timeout))
	}
	return &Conn{conn: c, reader: newReader(c)}, nil
}

func newReader(r io.Reader) *bufio.Reader {
	return bufio.NewReaderSize(r, 64*1024)
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// SetDeadline extends or shortens the deadline of the current request.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// WriteEvent sends one event.
func (c *Conn) WriteEvent(eventType string, data map[string]interface{}, payload []byte) error {
	h := header{Type: eventType, Version: protocolVersion}
	var dataBytes []byte
	if len(data) > 0 {
		var err error
		dataBytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
		h.DataLength = len(dataBytes)
	}
	h.PayloadLength = len(payload)
	headerBytes, err := json.Marshal(h)
	if err != nil {
		return err
	}
	buf := make([]byte, 0, len(headerBytes)+1+len(dataBytes)+len(payload))
	buf = append(buf, headerBytes...)
	buf = append(buf, '\n')
	buf = append(buf, dataBytes...)
	buf = append(buf, payload...)
	_, err = c.conn.Write(buf)
	return err
}

// ReadEvent reads one event. Data may arrive inline in the header or as a
// separate block - both forms are accepted.
func (c *Conn) ReadEvent() (*Event, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var h header
	if err := json.Unmarshal(line, &h); err != nil {
		return nil, fmt.Errorf("wyoming: invalid header: %w", err)
	}
	ev := &Event{Type: h.Type, Data: h.Data}
	if h.DataLength > 0 {
		dataBytes := make([]byte, h.DataLength)
		if _, err := io.ReadFull(c.reader, dataBytes); err != nil {
			return nil, err
		}
		extra := map[string]interface{}{}
		if err := json.Unmarshal(dataBytes, &extra); err != nil {
			return nil, fmt.Errorf("wyoming: invalid data block: %w", err)
		}
		if ev.Data == nil {
			ev.Data = extra
		} else {
			for k, v := range extra {
				ev.Data[k] = v
			}
		}
	}
	if h.PayloadLength > 0 {
		ev.Payload = make([]byte, h.PayloadLength)
		if _, err := io.ReadFull(c.reader, ev.Payload); err != nil {
			return nil, err
		}
	}
	return ev, nil
}

// readUntil reads events until one of the wanted types arrives. An "error"
// event from the server is turned into a Go error.
func (c *Conn) readUntil(types ...string) (*Event, error) {
	for {
		ev, err := c.ReadEvent()
		if err != nil {
			return nil, err
		}
		if ev.Type == "error" {
			return nil, fmt.Errorf("wyoming server error: %v", ev.Data["text"])
		}
		for _, t := range types {
			if ev.Type == t {
				return ev, nil
			}
		}
	}
}

// AudioFormat describes raw PCM audio.
type AudioFormat struct {
	Rate     int
	Width    int // bytes per sample
	Channels int
}

func (f AudioFormat) data() map[string]interface{} {
	return map[string]interface{}{"rate": f.Rate, "width": f.Width, "channels": f.Channels}
}

func formatFromData(d map[string]interface{}) AudioFormat {
	return AudioFormat{Rate: intValue(d["rate"]), Width: intValue(d["width"]), Channels: intValue(d["channels"])}
}

func intValue(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

// --- speech-to-text ---------------------------------------------------------

// Transcription streams audio to an ASR server (e.g. wyoming-faster-whisper).
type Transcription struct {
	conn   *Conn
	format AudioFormat
}

// StartTranscription opens a connection and announces the audio stream.
// language is an ISO 639-1 code ("pl", "en"); empty lets the server decide.
func StartTranscription(addr, language string, format AudioFormat) (*Transcription, error) {
	conn, err := Dial(addr, DefaultTimeout)
	if err != nil {
		return nil, err
	}
	transcribe := map[string]interface{}{}
	if language != "" {
		transcribe["language"] = language
	}
	if err := conn.WriteEvent("transcribe", transcribe, nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.WriteEvent("audio-start", format.data(), nil); err != nil {
		conn.Close()
		return nil, err
	}
	return &Transcription{conn: conn, format: format}, nil
}

// SendAudio sends one chunk of PCM audio.
func (t *Transcription) SendAudio(pcm []byte) error {
	if len(pcm) == 0 {
		return nil
	}
	return t.conn.WriteEvent("audio-chunk", t.format.data(), pcm)
}

// Finish ends the audio stream and waits for the transcript.
func (t *Transcription) Finish() (string, error) {
	defer t.conn.Close()
	// Whisper only starts decoding after audio-stop - give it its own time budget.
	t.conn.SetDeadline(time.Now().Add(DefaultTimeout))
	if err := t.conn.WriteEvent("audio-stop", nil, nil); err != nil {
		return "", err
	}
	ev, err := t.conn.readUntil("transcript")
	if err != nil {
		return "", err
	}
	text, _ := ev.Data["text"].(string)
	return text, nil
}

// Abort closes the stream without waiting for a transcript.
func (t *Transcription) Abort() {
	t.conn.Close()
}

// --- text-to-speech ---------------------------------------------------------

// Synthesize turns text into raw PCM audio using a TTS server (e.g.
// wyoming-piper). voice may be empty to use the server default.
func Synthesize(addr, text, voice string) ([]byte, AudioFormat, error) {
	conn, err := Dial(addr, DefaultTimeout)
	if err != nil {
		return nil, AudioFormat{}, err
	}
	defer conn.Close()
	data := map[string]interface{}{"text": text}
	if voice != "" {
		data["voice"] = map[string]interface{}{"name": voice}
	}
	if err := conn.WriteEvent("synthesize", data, nil); err != nil {
		return nil, AudioFormat{}, err
	}
	start, err := conn.readUntil("audio-start")
	if err != nil {
		return nil, AudioFormat{}, err
	}
	format := formatFromData(start.Data)
	var pcm []byte
	for {
		ev, err := conn.readUntil("audio-chunk", "audio-stop")
		if err != nil {
			return nil, AudioFormat{}, err
		}
		if ev.Type == "audio-stop" {
			break
		}
		// A chunk may announce its own format; trust the latest one.
		if chunkFormat := formatFromData(ev.Data); chunkFormat.Rate > 0 {
			format = chunkFormat
		}
		pcm = append(pcm, ev.Payload...)
	}
	if format.Rate == 0 || format.Width == 0 || format.Channels == 0 {
		return nil, AudioFormat{}, errors.New("wyoming: TTS server did not report the audio format")
	}
	return pcm, format, nil
}

// --- service discovery ------------------------------------------------------

// Info lists what a Wyoming server offers.
type Info struct {
	ASR []Program `json:"asr"`
	TTS []Program `json:"tts"`
}

// Program is one ASR or TTS engine on the server.
type Program struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Models  []Model `json:"models,omitempty"`
	Voices  []Model `json:"voices,omitempty"`
}

// Model is an ASR model or a TTS voice.
type Model struct {
	Name      string   `json:"name"`
	Languages []string `json:"languages"`
	Installed bool     `json:"installed"`
}

// Describe asks the server what it offers. It also serves as a connection test.
func Describe(addr string) (*Info, error) {
	conn, err := Dial(addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.WriteEvent("describe", nil, nil); err != nil {
		return nil, err
	}
	ev, err := conn.readUntil("info")
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(ev.Data)
	info := &Info{}
	if err := json.Unmarshal(raw, info); err != nil {
		return nil, err
	}
	return info, nil
}
