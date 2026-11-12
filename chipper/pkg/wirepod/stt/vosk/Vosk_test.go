package wirepod_vosk

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	"github.com/maxhawkins/go-webrtcvad"
)

// pcmStream feeds PCM chunks to STT, then ends the stream with io.EOF
// (what happens when the robot stops the mic stream on its own).
type pcmStream struct {
	pb.ChipperGrpc_StreamingIntentGraphServer
	chunks [][]byte
	next   int
	hitEOF bool
}

func (s *pcmStream) Recv() (*pb.StreamingIntentGraphRequest, error) {
	if s.next >= len(s.chunks) {
		s.hitEOF = true
		return nil, io.EOF
	}
	c := s.chunks[s.next]
	s.next++
	return &pb.StreamingIntentGraphRequest{InputAudio: c}, nil
}

func newTestReq(t *testing.T, pcm []byte) (sr.SpeechRequest, *pcmStream) {
	t.Helper()
	vad, err := webrtcvad.New()
	if err != nil {
		t.Fatal(err)
	}
	vad.SetMode(2)
	chunks := sr.SplitVAD(pcm)
	stream := &pcmStream{chunks: chunks[1:]}
	return sr.SpeechRequest{
		Device:   "test",
		FirstReq: chunks[0],
		Stream:   stream,
		IsIG:     true,
		VADInst:  vad,
	}, stream
}

func initTestModel(t *testing.T) {
	t.Helper()
	dir := os.Getenv("VOSK_TEST_MODELS")
	if dir == "" {
		t.Skip("set VOSK_TEST_MODELS to a vosk models dir (containing en-US/model) to run")
	}
	if _, err := os.Stat(filepath.Join(dir, "en-US", "model")); err != nil {
		t.Skip("no en-US vosk model in " + dir)
	}
	vars.VoskModelPath = dir
	vars.APIConfig.PastInitialSetup = true
	if err := Init(); err != nil {
		t.Fatal(err)
	}
}

// The stream ends (io.EOF) in the middle of speech, before the VAD saw the end.
// The partial transcript must be kept instead of returning an error.
func TestSTTKeepsPartialTranscriptOnEOF(t *testing.T) {
	initTestModel(t)
	pcm, err := os.ReadFile("../../../../stttest.pcm")
	if err != nil {
		t.Skip("stttest.pcm not found")
	}
	// the first second of the sample ("how are you"), cut off before the VAD sees the end of speech
	req, stream := newTestReq(t, pcm[:32000])
	text, err := STT(req)
	if err != nil {
		t.Fatalf("expected the partial transcript, got error: %v", err)
	}
	if !stream.hitEOF {
		t.Fatal("test did not exercise the EOF path")
	}
	if strings.TrimSpace(text) == "" {
		t.Fatal("expected a non-empty partial transcript")
	}
	t.Logf("partial transcript: %q", text)
}

// The stream ends with no speech heard: still an error (intent_system_noaudio).
func TestSTTErrorsOnEOFWithoutSpeech(t *testing.T) {
	initTestModel(t)
	req, _ := newTestReq(t, make([]byte, 32000))
	if _, err := STT(req); err == nil {
		t.Fatal("expected an error when the stream ends without speech")
	}
}
