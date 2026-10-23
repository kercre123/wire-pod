package webserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wirepod/tts"
	"github.com/kercre123/wire-pod/chipper/pkg/wyoming"
)

// Settings for external Wyoming voice services (speech-to-text and
// text-to-speech). See the "Voice Services" section of the setup page.

func handleGetWyoming(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Wyoming)
}

func handleSetWyoming(w http.ResponseWriter, r *http.Request) {
	cfg := vars.APIConfig.Wyoming
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	cfg.STT.Host = strings.TrimSpace(cfg.STT.Host)
	cfg.TTS.Host = strings.TrimSpace(cfg.TTS.Host)
	cfg.TTS.Voice = strings.TrimSpace(cfg.TTS.Voice)
	if cfg.STT.Enable && (cfg.STT.Host == "" || cfg.STT.Port <= 0) {
		http.Error(w, "speech-to-text: host and port are required", http.StatusBadRequest)
		return
	}
	if cfg.TTS.Enable && (cfg.TTS.Host == "" || cfg.TTS.Port <= 0) {
		http.Error(w, "text-to-speech: host and port are required", http.StatusBadRequest)
		return
	}
	if cfg.TTS.Volume < 0 || cfg.TTS.Volume > 100 {
		http.Error(w, "volume must be between 0 and 100", http.StatusBadRequest)
		return
	}
	vars.APIConfig.Wyoming = cfg
	vars.WriteConfigToDisk()
	logger.Println(fmt.Sprintf("Wyoming voice services: STT %v (%s:%d), TTS %v (%s:%d, voice %q)",
		cfg.STT.Enable, cfg.STT.Host, cfg.STT.Port, cfg.TTS.Enable, cfg.TTS.Host, cfg.TTS.Port, cfg.TTS.Voice))
	fmt.Fprint(w, "Changes successfully applied.")
}

// handleWyomingDescribe tests the connection to a Wyoming server and returns
// what it offers (ASR models, TTS voices).
func handleWyomingDescribe(w http.ResponseWriter, r *http.Request) {
	host := strings.TrimSpace(r.FormValue("host"))
	port, _ := strconv.Atoi(r.FormValue("port"))
	if host == "" || port <= 0 {
		http.Error(w, "host and port are required", http.StatusBadRequest)
		return
	}
	info, err := wyoming.Describe(wyoming.Address(host, port))
	if err != nil {
		http.Error(w, "cannot connect: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleWyomingTTSPreview synthesizes text with the settings from the request
// and returns it as WAV, exactly as the robot would receive it (16 kHz mono),
// so a voice can be checked in the browser without a robot.
func handleWyomingTTSPreview(w http.ResponseWriter, r *http.Request) {
	host := strings.TrimSpace(r.FormValue("host"))
	port, _ := strconv.Atoi(r.FormValue("port"))
	text := strings.TrimSpace(r.FormValue("text"))
	if host == "" || port <= 0 || text == "" {
		http.Error(w, "host, port and text are required", http.StatusBadRequest)
		return
	}
	if len([]rune(text)) > 600 {
		http.Error(w, "text is too long", http.StatusBadRequest)
		return
	}
	pcm, format, err := wyoming.Synthesize(wyoming.Address(host, port), text, strings.TrimSpace(r.FormValue("voice")))
	if err != nil {
		http.Error(w, "synthesis failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	audio := tts.NormalizePeak(wyoming.ToMono16(pcm, format, 16000))
	w.Header().Set("Content-Type", "audio/wav")
	w.Write(wavFile(audio, 16000))
}

// wavFile wraps 16-bit mono PCM in a WAV header.
func wavFile(pcm []byte, rate int) []byte {
	var b bytes.Buffer
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+len(pcm)))
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(&b, binary.LittleEndian, uint16(1))  // PCM
	binary.Write(&b, binary.LittleEndian, uint16(1))  // mono
	binary.Write(&b, binary.LittleEndian, uint32(rate))
	binary.Write(&b, binary.LittleEndian, uint32(rate*2)) // byte rate
	binary.Write(&b, binary.LittleEndian, uint16(2))      // block align
	binary.Write(&b, binary.LittleEndian, uint16(16))     // bits per sample
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(pcm)))
	b.Write(pcm)
	return b.Bytes()
}
