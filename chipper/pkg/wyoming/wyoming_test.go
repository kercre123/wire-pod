package wyoming

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"testing"
)

// fakeServer accepts one connection and runs handler on it.
func fakeServer(t *testing.T, handler func(c *Conn)) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		raw, err := ln.Accept()
		if err != nil {
			return
		}
		c := &Conn{conn: raw, reader: newReader(raw)}
		defer c.Close()
		handler(c)
	}()
	return ln.Addr().String()
}

func TestTranscription(t *testing.T) {
	var gotLanguage string
	var gotAudio []byte
	addr := fakeServer(t, func(c *Conn) {
		for {
			ev, err := c.ReadEvent()
			if err != nil {
				return
			}
			switch ev.Type {
			case "transcribe":
				gotLanguage, _ = ev.Data["language"].(string)
			case "audio-chunk":
				gotAudio = append(gotAudio, ev.Payload...)
			case "audio-stop":
				c.WriteEvent("transcript", map[string]interface{}{"text": "Która jest godzina?"}, nil)
				return
			}
		}
	})

	tr, err := StartTranscription(addr, "pl", AudioFormat{Rate: 16000, Width: 2, Channels: 1})
	if err != nil {
		t.Fatal(err)
	}
	tr.SendAudio([]byte{1, 2, 3, 4})
	tr.SendAudio([]byte{5, 6})
	text, err := tr.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Która jest godzina?" {
		t.Errorf("transcript = %q", text)
	}
	if gotLanguage != "pl" {
		t.Errorf("language = %q, want pl", gotLanguage)
	}
	if string(gotAudio) != string([]byte{1, 2, 3, 4, 5, 6}) {
		t.Errorf("audio = %v", gotAudio)
	}
}

func TestSynthesize(t *testing.T) {
	var gotText, gotVoice string
	addr := fakeServer(t, func(c *Conn) {
		ev, err := c.ReadEvent()
		if err != nil || ev.Type != "synthesize" {
			return
		}
		gotText, _ = ev.Data["text"].(string)
		if v, ok := ev.Data["voice"].(map[string]interface{}); ok {
			gotVoice, _ = v["name"].(string)
		}
		format := map[string]interface{}{"rate": 22050, "width": 2, "channels": 1}
		c.WriteEvent("audio-start", format, nil)
		c.WriteEvent("audio-chunk", format, []byte{1, 0, 2, 0})
		c.WriteEvent("audio-chunk", format, []byte{3, 0})
		c.WriteEvent("audio-stop", nil, nil)
	})

	pcm, format, err := Synthesize(addr, "Cześć", "pl_PL-gosia-medium")
	if err != nil {
		t.Fatal(err)
	}
	if gotText != "Cześć" || gotVoice != "pl_PL-gosia-medium" {
		t.Errorf("server got text=%q voice=%q", gotText, gotVoice)
	}
	if format.Rate != 22050 || format.Width != 2 || format.Channels != 1 {
		t.Errorf("format = %+v", format)
	}
	if len(pcm) != 6 {
		t.Errorf("pcm length = %d, want 6", len(pcm))
	}
}

func TestServerErrorEvent(t *testing.T) {
	addr := fakeServer(t, func(c *Conn) {
		c.ReadEvent()
		c.WriteEvent("error", map[string]interface{}{"text": "voice not found"}, nil)
	})
	_, _, err := Synthesize(addr, "x", "nope")
	if err == nil || !strings.Contains(err.Error(), "voice not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestInlineData(t *testing.T) {
	// Older servers put data inside the header line instead of a separate block.
	addr := fakeServer(t, func(c *Conn) {
		c.ReadEvent()
		c.conn.Write([]byte(`{"type":"info","data":{"tts":[{"name":"piper","voices":[{"name":"pl_PL-gosia-medium","languages":["pl_PL"],"installed":true}]}]}}` + "\n"))
	})
	info, err := Describe(addr)
	if err != nil {
		t.Fatal(err)
	}
	if len(info.TTS) != 1 || len(info.TTS[0].Voices) != 1 || info.TTS[0].Voices[0].Name != "pl_PL-gosia-medium" {
		t.Errorf("info = %+v", info)
	}
}

func TestToMono16(t *testing.T) {
	// stereo 16-bit, 2 frames: (100, 300) and (-100, -300) -> mono 200, -200
	in := make([]byte, 8)
	for i, v := range []int16{100, 300, -100, -300} {
		binary.LittleEndian.PutUint16(in[i*2:], uint16(v))
	}
	out := ToMono16(in, AudioFormat{Rate: 16000, Width: 2, Channels: 2}, 16000)
	if len(out) != 4 {
		t.Fatalf("len = %d", len(out))
	}
	if a, b := int16(binary.LittleEndian.Uint16(out)), int16(binary.LittleEndian.Uint16(out[2:])); a != 200 || b != -200 {
		t.Errorf("samples = %d, %d", a, b)
	}

	// 22050 Hz -> 16000 Hz keeps duration
	one := make([]byte, 22050*2)
	out = ToMono16(one, AudioFormat{Rate: 22050, Width: 2, Channels: 1}, 16000)
	if len(out) != 16000*2 {
		t.Errorf("resampled len = %d, want %d", len(out), 16000*2)
	}
}

// Integration test against real servers:
//
//	WYOMING_TEST_HOST=10.20.1.19 WYOMING_TEST_VOICE=pl_PL-gosia-medium go test ./pkg/wyoming -run Live -v
func TestLiveRoundTrip(t *testing.T) {
	host := os.Getenv("WYOMING_TEST_HOST")
	if host == "" {
		t.Skip("WYOMING_TEST_HOST not set")
	}
	voice := os.Getenv("WYOMING_TEST_VOICE")
	pcm, format, err := Synthesize(Address(host, 10200), "Która jest godzina?", voice)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("piper: %d bytes, %+v", len(pcm), format)
	mono := ToMono16(pcm, format, 16000)

	tr, err := StartTranscription(Address(host, 10300), "pl", AudioFormat{Rate: 16000, Width: 2, Channels: 1})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(mono); i += 1024 {
		end := i + 1024
		if end > len(mono) {
			end = len(mono)
		}
		tr.SendAudio(mono[i:end])
	}
	text, err := tr.Finish()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("whisper: %q", text)
	if !strings.Contains(strings.ToLower(text), "godzin") {
		t.Errorf("unexpected transcript %q", text)
	}
}
