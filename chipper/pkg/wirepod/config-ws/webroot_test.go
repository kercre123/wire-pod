package webserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWebRootRoutes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "classic home")
	writeFile(t, filepath.Join(dir, "setup.html"), "classic setup")
	writeFile(t, filepath.Join(dir, "initial.html"), "classic initial")
	writeFile(t, filepath.Join(dir, "sdkapp", "settings.html"), "classic bot settings")
	writeFile(t, filepath.Join(dir, "css", "style.css"), "body{}")
	writeFile(t, filepath.Join(dir, "next", "index.html"), "new ui")
	writeFile(t, filepath.Join(dir, "next", "app.js"), "// app")
	h := newWebRootHandler(dir)

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		return rec
	}
	redirect := func(path, want string) {
		t.Helper()
		rec := get(path)
		if rec.Code != http.StatusFound && rec.Code != http.StatusMovedPermanently {
			t.Errorf("%s: want a redirect, got %d", path, rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != want {
			t.Errorf("%s: redirects to %q, want %q", path, loc, want)
		}
	}
	serves := func(path, want string) {
		t.Helper()
		rec := get(path)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Errorf("%s: got %d %q, want 200 containing %q", path, rec.Code, rec.Body.String(), want)
		}
	}

	serves("/", "classic home") // classic is the default front page
	redirect("/next", "next/")  // FileServer redirects relatively
	redirect("/classic/", "/classic")
	serves("/next/", "new ui")
	serves("/next/app.js", "// app")
	// the original UI and every old URL keep working
	serves("/classic", "classic home")
	serves("/index.html", "classic home")
	serves("/setup.html", "classic setup")
	serves("/initial.html", "classic initial")
	serves("/sdkapp/settings.html", "classic bot settings")
	serves("/css/style.css", "body{}")
	if rec := get("/missing.html"); rec.Code != http.StatusNotFound {
		t.Errorf("missing file: got %d", rec.Code)
	}
}

func TestWebRootDefaultFollowsConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "classic home")
	writeFile(t, filepath.Join(dir, "next", "index.html"), "new ui")
	h := newWebRootHandler(dir)
	old := vars.APIConfig.UI.Default
	t.Cleanup(func() { vars.APIConfig.UI.Default = old })

	for _, def := range []string{"", "classic", "something-else"} {
		vars.APIConfig.UI.Default = def
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "classic home") {
			t.Errorf("default %q: got %d %q, want classic", def, rec.Code, rec.Body.String())
		}
	}

	vars.APIConfig.UI.Default = "next"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if loc := rec.Header().Get("Location"); rec.Code != http.StatusFound || loc != "/next/" {
		t.Errorf("default next: got %d to %q, want a redirect to /next/", rec.Code, loc)
	}
	// both stay reachable whichever is the default
	for path, want := range map[string]string{"/classic": "classic home", "/next/": "new ui"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Errorf("%s with default next: got %d %q", path, rec.Code, rec.Body.String())
		}
	}
}

func TestSetUIDefault(t *testing.T) {
	old, oldPath := vars.APIConfig.UI.Default, vars.ApiConfigPath
	vars.ApiConfigPath = filepath.Join(t.TempDir(), "apiConfig.json")
	t.Cleanup(func() { vars.APIConfig.UI.Default, vars.ApiConfigPath = old, oldPath })
	vars.APIConfig.UI.Default = ""

	post := func(body string) int {
		rec := httptest.NewRecorder()
		handleSetUIDefault(rec, httptest.NewRequest("POST", "/api/set_ui_default", strings.NewReader(body)))
		return rec.Code
	}
	if c := post(`{"default":"next"}`); c != 200 || vars.APIConfig.UI.Default != "next" {
		t.Fatalf("set next: %d %q", c, vars.APIConfig.UI.Default)
	}
	if b, err := os.ReadFile(vars.ApiConfigPath); err != nil || !strings.Contains(string(b), `"ui":{"default":"next"}`) {
		t.Fatalf("not saved to the config file: %v %q", err, b)
	}
	for _, bad := range []string{`{"default":"bogus"}`, `{}`, `nope`} {
		if c := post(bad); c != 400 || vars.APIConfig.UI.Default != "next" {
			t.Fatalf("%s: got %d, default now %q", bad, c, vars.APIConfig.UI.Default)
		}
	}
	if c := post(`{"default":"classic"}`); c != 200 || vars.APIConfig.UI.Default != "classic" {
		t.Fatalf("set classic: %d %q", c, vars.APIConfig.UI.Default)
	}
}

func TestOldConfigWithoutUILoads(t *testing.T) {
	old := vars.APIConfig
	t.Cleanup(func() { vars.APIConfig = old })
	var loaded = vars.APIConfig
	loaded.UI.Default = ""
	if err := json.Unmarshal([]byte(`{"weather":{"enable":true},"server":{"port":"8080"}}`), &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.UI.Default != "" || !loaded.Weather.Enable {
		t.Fatalf("old config should load unchanged with no ui setting: %+v", loaded.UI)
	}
}

func TestClassicUIMissingIsNotFound(t *testing.T) {
	h := newWebRootHandler(t.TempDir())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/classic", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}
