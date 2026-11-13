package sdkapp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func setupBotInfo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldInfo, oldPath := vars.BotInfo, vars.BotInfoPath
	t.Cleanup(func() { vars.BotInfo, vars.BotInfoPath = oldInfo, oldPath })
	vars.BotInfoPath = filepath.Join(dir, "botSdkInfo.json")
	vars.BotInfo = vars.RobotInfoStore{}
	vars.BotInfo.Robots = append(vars.BotInfo.Robots, struct {
		Esn         string `json:"esn"`
		IPAddress   string `json:"ip_address"`
		GUID        string `json:"guid"`
		Activated   bool   `json:"activated"`
		Name        string `json:"name,omitempty"`
		Personality string `json:"personality,omitempty"`
	}{Esn: "aaaa1111", Name: "Orbit", Personality: "calm and kind"})
	return dir
}

func post(path string, form url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleRobotIdentityAPI(w, r)
	return w
}

func TestSetIdentityOnlyChangesSentFields(t *testing.T) {
	setupBotInfo(t)
	if w := post("/api-sdk/set_robot_identity", url.Values{"serial": {"aaaa1111"}, "name": {"Nova"}}); w.Code != 200 {
		t.Fatalf("name-only update failed: %d %s", w.Code, w.Body)
	}
	if got := vars.BotInfo.Robots[0]; got.Name != "Nova" || got.Personality != "calm and kind" {
		t.Fatalf("name-only update changed personality: %+v", got)
	}
	if w := post("/api-sdk/set_robot_identity", url.Values{"serial": {"aaaa1111"}, "personality": {"silly"}}); w.Code != 200 {
		t.Fatalf("personality-only update failed: %d", w.Code)
	}
	if got := vars.BotInfo.Robots[0]; got.Name != "Nova" || got.Personality != "silly" {
		t.Fatalf("personality-only update changed name: %+v", got)
	}
	// an explicitly empty value clears just that field
	post("/api-sdk/set_robot_identity", url.Values{"serial": {"aaaa1111"}, "personality": {""}})
	if got := vars.BotInfo.Robots[0]; got.Name != "Nova" || got.Personality != "" {
		t.Fatalf("empty personality should clear only personality: %+v", got)
	}
	// the legacy name endpoint never touches the personality
	post("/api-sdk/set_robot_identity", url.Values{"serial": {"aaaa1111"}, "personality": {"keep me"}})
	post("/api-sdk/set_robot_name", url.Values{"serial": {"aaaa1111"}, "name": {"Pip"}})
	if got := vars.BotInfo.Robots[0]; got.Name != "Pip" || got.Personality != "keep me" {
		t.Fatalf("set_robot_name touched personality: %+v", got)
	}
	if w := post("/api-sdk/set_robot_identity", url.Values{"serial": {"zzzz9999"}, "name": {"X"}}); w.Code != http.StatusNotFound {
		t.Fatalf("unknown robot: %d", w.Code)
	}
}

func TestFailedSaveRollsBackAndKeepsFile(t *testing.T) {
	dir := setupBotInfo(t)
	original := []byte(`{"global_guid":"g","robots":[]}`)
	if err := os.WriteFile(vars.BotInfoPath, original, 0644); err != nil {
		t.Fatal(err)
	}
	// make the temp file impossible to create next to the store
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0755)
	n := "Nova"
	if err := setRobotIdentity("aaaa1111", &n, nil); err == nil {
		t.Skip("directory permissions not enforced here (running as root?)")
	}
	if vars.BotInfo.Robots[0].Name != "Orbit" {
		t.Fatalf("in-memory change was not rolled back: %+v", vars.BotInfo.Robots[0])
	}
	if b, _ := os.ReadFile(vars.BotInfoPath); string(b) != string(original) {
		t.Fatalf("failed save modified the stored file: %s", b)
	}
}

func TestSaveWritesWholeFileAndLeavesNoTemp(t *testing.T) {
	dir := setupBotInfo(t)
	n := "Nova"
	if err := setRobotIdentity("AAAA1111", &n, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(vars.BotInfoPath)
	if err != nil || !strings.Contains(string(b), `"name":"Nova"`) {
		t.Fatalf("saved file wrong: %s %v", b, err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temp file left behind: %v", entries)
	}
}
