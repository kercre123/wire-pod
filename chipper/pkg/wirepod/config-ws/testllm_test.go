package webserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTestLLM(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		switch {
		case r.URL.Path != "/v1/models":
			http.NotFound(w, r)
		case gotAuth == "Bearer bad":
			http.Error(w, "nope", http.StatusUnauthorized)
		default:
			w.Write([]byte(`{"data":[{"id":"a"},{"id":"b"}]}`))
		}
	}))
	defer srv.Close()

	res := testLLM("custom", "good", srv.URL+"/v1/")
	if !res.OK || res.Message != "Connected. 2 models available." || gotAuth != "Bearer good" {
		t.Fatalf("got %+v auth %q", res, gotAuth)
	}
	if res := testLLM("custom", "bad", srv.URL+"/v1"); res.OK || !strings.Contains(res.Message, "rejected the key") {
		t.Fatalf("bad key: %+v", res)
	}
	if res := testLLM("custom", "", srv.URL+"/wrong"); res.OK || !strings.Contains(res.Message, "HTTP 404") {
		t.Fatalf("wrong path: %+v", res)
	}
	if res := testLLM("custom", "", "http://127.0.0.1:1/v1"); res.OK || !strings.Contains(res.Message, "Couldn’t reach") {
		t.Fatalf("closed port: %+v", res)
	}
	for _, ep := range []string{"", "localhost:11434/v1", "ftp://x/v1"} {
		if res := testLLM("custom", "", ep); res.OK || !strings.Contains(res.Message, "server address") {
			t.Fatalf("endpoint %q should be refused: %+v", ep, res)
		}
	}
	if res := testLLM("houndify", "", ""); res.OK {
		t.Fatalf("houndify can't be tested: %+v", res)
	}
	if res := testLLM("", "", ""); res.OK {
		t.Fatalf("no provider: %+v", res)
	}
}

func TestHandleTestLLMRejectsBadBody(t *testing.T) {
	rec := httptest.NewRecorder()
	handleTestLLM(rec, httptest.NewRequest("POST", "/api/test_llm", strings.NewReader("not json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d", rec.Code)
	}
}
