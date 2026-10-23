// Package wirepod_wyoming sends Vector's microphone audio to an external
// Wyoming speech-to-text server (e.g. wyoming-faster-whisper, the one Home
// Assistant uses). It is not a standalone voice processor: it wraps the
// compiled-in engine (VOSK etc.) and is switched on at runtime in the web UI.
package wirepod_wyoming

import (
	"errors"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	"github.com/kercre123/wire-pod/chipper/pkg/wyoming"
)

var Name string = "wyoming"

// ErrUnavailable means the server could not be reached. It is returned before
// any audio was read from the robot, so the caller can still hand the request
// to the built-in engine.
var ErrUnavailable = errors.New("wyoming STT server unavailable")

// Enabled reports whether Wyoming STT is switched on and configured.
func Enabled() bool {
	c := vars.APIConfig.Wyoming.STT
	return c.Enable && c.Host != "" && c.Port > 0
}

// Language returns the ISO 639-1 code for the configured STT language
// ("pl-PL" -> "pl").
func Language() string {
	return strings.ToLower(strings.Split(vars.APIConfig.STT.Language, "-")[0])
}

// STT streams the request audio to the Wyoming server and returns the text,
// normalized the way the built-in engines return it (lowercase, no punctuation).
func STT(req sr.SpeechRequest) (string, error) {
	c := vars.APIConfig.Wyoming.STT
	addr := wyoming.Address(c.Host, c.Port)
	tr, err := wyoming.StartTranscription(addr, Language(), wyoming.AudioFormat{Rate: 16000, Width: 2, Channels: 1})
	if err != nil {
		logger.Println("(Bot " + req.Device + ", Wyoming) cannot connect to " + addr + ": " + err.Error())
		return "", ErrUnavailable
	}
	logger.Println("(Bot " + req.Device + ", Wyoming) Processing...")

	// audio already decoded from the first request
	if err := tr.SendAudio(req.DecodedMicData); err != nil {
		tr.Abort()
		return "", err
	}
	for {
		chunk, err := req.GetNextStreamChunk()
		if err != nil {
			tr.Abort()
			return "", err
		}
		if err := tr.SendAudio(chunk); err != nil {
			tr.Abort()
			return "", err
		}
		if speechIsDone, _ := req.DetectEndOfSpeech(); speechIsDone {
			break
		}
	}

	text, err := tr.Finish()
	if err != nil {
		logger.Println("(Bot " + req.Device + ", Wyoming) error: " + err.Error())
		return "", err
	}
	text = Normalize(text)
	logger.Println("Bot " + req.Device + " Transcribed text: " + text)
	return text, nil
}

// Normalize lowercases the transcript and drops punctuation, so intent
// matching sees the same kind of text VOSK produces ("Która jest godzina?" ->
// "która jest godzina"). Digits are kept - words2num understands them.
func Normalize(text string) string {
	var b strings.Builder
	lastSpace := true
	for _, r := range strings.ToLower(text) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'':
			b.WriteRune(r)
			lastSpace = false
		case !lastSpace:
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}
