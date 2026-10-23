package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func callAPI(method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	apiHandler(rec, req)
	return rec
}

func TestSetGetWyoming(t *testing.T) {
	vars.ApiConfigPath = filepath.Join(t.TempDir(), "apiConfig.json")

	rec := callAPI("POST", "/api/set_wyoming", `{"stt":{"enable":true,"host":" 10.0.0.5 ","port":10300},"tts":{"enable":true,"host":"10.0.0.5","port":10200,"voice":"pl_PL-gosia-medium","volume":80}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set_wyoming: %d %s", rec.Code, rec.Body.String())
	}
	if vars.APIConfig.Wyoming.STT.Host != "10.0.0.5" {
		t.Errorf("host not trimmed: %q", vars.APIConfig.Wyoming.STT.Host)
	}
	saved, _ := os.ReadFile(vars.ApiConfigPath)
	if !strings.Contains(string(saved), `"voice":"pl_PL-gosia-medium"`) {
		t.Errorf("config not written to disk: %s", saved)
	}

	rec = callAPI("GET", "/api/get_wyoming", "")
	if !strings.Contains(rec.Body.String(), `"volume":80`) {
		t.Errorf("get_wyoming = %s", rec.Body.String())
	}
}

func TestSetWyomingValidation(t *testing.T) {
	vars.ApiConfigPath = filepath.Join(t.TempDir(), "apiConfig.json")
	cases := []string{
		`{"stt":{"enable":true,"host":"","port":10300}}`,
		`{"tts":{"enable":true,"host":"x","port":0}}`,
		`{"tts":{"enable":false,"volume":150}}`,
		`not json`,
	}
	for _, body := range cases {
		if rec := callAPI("POST", "/api/set_wyoming", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: code %d, want 400", body, rec.Code)
		}
	}
}

func TestWavFile(t *testing.T) {
	wav := wavFile(make([]byte, 320), 16000)
	if len(wav) != 44+320 || string(wav[:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		t.Errorf("bad wav header: %q", wav[:12])
	}
}

// WYOMING_TEST_HOST=10.20.1.19 go test ./pkg/wirepod/config-ws -run Live -v
func TestLiveDescribeAndPreview(t *testing.T) {
	host := os.Getenv("WYOMING_TEST_HOST")
	if host == "" {
		t.Skip("WYOMING_TEST_HOST not set")
	}
	rec := callAPI("GET", "/api/wyoming_describe?host="+host+"&port=10200", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "pl_PL") {
		t.Fatalf("describe: %d %.200s", rec.Code, rec.Body.String())
	}
	rec = callAPI("GET", "/api/wyoming_tts_preview?host="+host+"&port=10200&voice=pl_PL-gosia-medium&text=Cze%C5%9B%C4%87", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "audio/wav" || rec.Body.Len() < 1000 {
		t.Fatalf("preview: %d %s, %d bytes", rec.Code, rec.Header().Get("Content-Type"), rec.Body.Len())
	}
	t.Logf("preview: %d bytes of WAV", rec.Body.Len())

	rec = callAPI("GET", "/api/wyoming_describe?host=127.0.0.1&port=1", "")
	if rec.Code != http.StatusBadGateway {
		t.Errorf("unreachable server: code %d, want 502", rec.Code)
	}
}
