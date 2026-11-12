package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	redirect("/", "/next/")
	redirect("/next", "next/") // FileServer redirects relatively
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

func TestClassicUIMissingIsNotFound(t *testing.T) {
	h := newWebRootHandler(t.TempDir())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/classic", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}
