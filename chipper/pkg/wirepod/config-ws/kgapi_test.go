package webserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func postKG(t *testing.T, body string) int {
	t.Helper()
	rec := httptest.NewRecorder()
	handleSetKGAPI(rec, httptest.NewRequest("POST", "/api/set_kg_api", strings.NewReader(body)))
	return rec.Code
}

func kgFields(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	m := map[string]any{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// Turning AI off and back on must not lose the saved LLM setup, whether the client sends only
// {"enable":false} or a full body that carries the saved fields.
func TestSetKGAPIDisableThenEnableKeepsSettings(t *testing.T) {
	oldPath, oldKG := vars.ApiConfigPath, vars.APIConfig.Knowledge
	vars.ApiConfigPath = filepath.Join(t.TempDir(), "apiConfig.json")
	t.Cleanup(func() { vars.ApiConfigPath, vars.APIConfig.Knowledge = oldPath, oldKG })

	full := `{"enable":true,"provider":"custom","key":"lm-studio","endpoint":"http://localhost:1234/v1","model":"qwen/qwen3.6-35b-a3b","openai_prompt":"Be brief.","temp":0.5,"top_p":0.25,"intentgraph":true,"commands_enable":true,"save_chat":true,"robotName":"Vector"}`
	if code := postKG(t, full); code != http.StatusOK {
		t.Fatalf("save: status %d", code)
	}
	want := kgFields(t, []byte(full))

	check := func(step string, enable bool) {
		t.Helper()
		rec := httptest.NewRecorder()
		handleGetKGAPI(rec)
		disk, err := os.ReadFile(vars.ApiConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		var onDisk struct {
			Knowledge json.RawMessage `json:"knowledge"`
		}
		if err := json.Unmarshal(disk, &onDisk); err != nil {
			t.Fatal(err)
		}
		for where, raw := range map[string][]byte{"api": rec.Body.Bytes(), "disk": onDisk.Knowledge} {
			got := kgFields(t, raw)
			if got["enable"] != enable {
				t.Fatalf("%s %s: enable = %v, want %v", step, where, got["enable"], enable)
			}
			for k, v := range want {
				if k != "enable" && !reflect.DeepEqual(got[k], v) {
					t.Fatalf("%s %s: %s = %v, want %v", step, where, k, got[k], v)
				}
			}
		}
	}

	for _, off := range []string{`{"enable":false}`, strings.Replace(full, `"enable":true`, `"enable":false`, 1)} {
		if code := postKG(t, off); code != http.StatusOK {
			t.Fatalf("disable: status %d", code)
		}
		check("disabled", false)
		if code := postKG(t, `{"enable":true}`); code != http.StatusOK {
			t.Fatalf("enable: status %d", code)
		}
		check("re-enabled", true)
	}

	// a bad body must not change anything
	if code := postKG(t, `{"enable":false,"model":`); code != http.StatusBadRequest {
		t.Fatalf("bad body: status %d", code)
	}
	check("after bad body", true)
}
