// Package tts makes Vector speak with an external Wyoming text-to-speech
// server (e.g. wyoming-piper) instead of its built-in voice, which only
// pronounces English, German, French and Japanese correctly.
//
// The audio is synthesized by the server, converted to 16 kHz mono and played
// through the robot speaker with ExternalAudioStreamPlayback.
package tts

import (
	"context"
	"encoding/binary"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wyoming"
)

// Vector's speaker stream sample rate.
const robotSampleRate = 16000

// Bytes per streamed chunk and the pause between chunks. 1024 bytes is 32 ms
// of audio, sending every 25 ms keeps the robot buffer ahead of playback
// (same values as the OpenAI voice path in ttr).
const chunkSize = 1024
const chunkInterval = 25 * time.Millisecond

// Enabled reports whether Wyoming TTS is switched on and configured.
func Enabled() bool {
	c := vars.APIConfig.Wyoming.TTS
	return c.Enable && c.Host != "" && c.Port > 0
}

// SayText speaks text with the Wyoming voice when it is enabled, otherwise
// with Vector's built-in voice. It is the drop-in replacement for
// robot.Conn.SayText used everywhere wire-pod makes the robot talk. If the
// Wyoming server fails, the built-in voice is used so the robot never stays
// silent.
func SayText(ctx context.Context, robot *vector.Vector, text string, durationScalar float32) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if Enabled() {
		err := Speak(ctx, robot, text)
		if err == nil {
			return nil
		}
		logger.Println("Wyoming TTS failed, using Vector's voice: " + err.Error())
	}
	_, err := robot.Conn.SayText(ctx, &vectorpb.SayTextRequest{
		Text:           text,
		UseVectorVoice: true,
		DurationScalar: durationScalar,
	})
	return err
}

// Speak synthesizes text with the Wyoming server and plays it on the robot.
// It blocks until playback has finished.
func Speak(ctx context.Context, robot *vector.Vector, text string) error {
	audio, err := Synthesize(text)
	if err != nil {
		return err
	}
	return Play(ctx, robot, audio)
}

// Synthesize returns 16 kHz mono 16-bit PCM ready to stream to Vector.
func Synthesize(text string) ([]byte, error) {
	c := vars.APIConfig.Wyoming.TTS
	start := time.Now()
	pcm, format, err := wyoming.Synthesize(wyoming.Address(c.Host, c.Port), text, c.Voice)
	if err != nil {
		return nil, err
	}
	audio := NormalizePeak(wyoming.ToMono16(pcm, format, robotSampleRate))
	logger.Println("Wyoming TTS: synthesized " + duration(audio).String() + " of audio in " + time.Since(start).Round(time.Millisecond).String())
	return audio, nil
}

// Play streams 16 kHz mono PCM to the robot speaker and waits until it has played.
func Play(ctx context.Context, robot *vector.Vector, audio []byte) error {
	stream, err := robot.Conn.ExternalAudioStreamPlayback(ctx)
	if err != nil {
		return err
	}
	err = stream.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: robotSampleRate,
				AudioVolume:    uint32(volume()),
			},
		},
	})
	if err != nil {
		return err
	}
	started := time.Now()
	for off := 0; off < len(audio); off += chunkSize {
		chunk := make([]byte, chunkSize)
		copy(chunk, audio[off:])
		err = stream.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
				AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
					AudioChunkSizeBytes: chunkSize,
					AudioChunkSamples:   chunk,
				},
			},
		})
		if err != nil {
			return err
		}
		time.Sleep(chunkInterval)
	}
	err = stream.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
			AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
		},
	})
	// wait for the rest of the audio to be played before returning
	if remaining := duration(audio) + 100*time.Millisecond - time.Since(started); remaining > 0 {
		select {
		case <-time.After(remaining):
		case <-ctx.Done():
		}
	}
	return err
}

func volume() int {
	v := vars.APIConfig.Wyoming.TTS.Volume
	if v <= 0 || v > 100 {
		return 100
	}
	return v
}

func duration(pcm []byte) time.Duration {
	return time.Duration(len(pcm)/2) * time.Second / robotSampleRate
}

// NormalizePeak scales the audio so its loudest sample reaches ~90% of full
// scale. Vector's speaker is small and TTS voices are often recorded quietly.
func NormalizePeak(pcm []byte) []byte {
	var peak int
	for i := 0; i+1 < len(pcm); i += 2 {
		s := int(int16(binary.LittleEndian.Uint16(pcm[i:])))
		if s < 0 {
			s = -s
		}
		if s > peak {
			peak = s
		}
	}
	if peak == 0 {
		return pcm
	}
	gain := 0.9 * 32767 / float64(peak)
	if gain <= 1 {
		return pcm
	}
	if gain > 8 {
		gain = 8
	}
	for i := 0; i+1 < len(pcm); i += 2 {
		s := float64(int16(binary.LittleEndian.Uint16(pcm[i:]))) * gain
		binary.LittleEndian.PutUint16(pcm[i:], uint16(int16(s)))
	}
	return pcm
}
