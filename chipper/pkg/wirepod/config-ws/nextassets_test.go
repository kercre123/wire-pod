package webserver

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

const testCSS = `@font-face{src:url(fonts/a.woff2) format("woff2")}` + "\n" + `body{margin:0}`

func nextFixture(t *testing.T) (string, http.Handler) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "classic home")
	writeFile(t, filepath.Join(dir, "css", "style.css"), "body{}")
	writeFile(t, filepath.Join(dir, "next", "index.html"), `<script type="importmap">{"imports":{"preact":"./vendor/p.module.js"}}</script>`)
	writeFile(t, filepath.Join(dir, "next", "app.css"), testCSS)
	writeFile(t, filepath.Join(dir, "next", "app.js"), strings.Repeat("export const a = 1;\n", 200))
	writeFile(t, filepath.Join(dir, "next", "vendor", "p.module.js"), "export const p = 1;")
	writeFile(t, filepath.Join(dir, "next", "fonts", "a.woff2"), "woff2 bytes")
	return dir, newWebRootHandler(dir)
}

func do(h http.Handler, path string, hdr ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestNextAppFilesRevalidateAndAre304(t *testing.T) {
	_, h := nextFixture(t)
	first := do(h, "/next/app.js")
	if first.Code != 200 || first.Header().Get("Cache-Control") != "no-cache" || first.Header().Get("ETag") == "" || first.Header().Get("Last-Modified") == "" {
		t.Fatalf("app.js: %d %v", first.Code, first.Header())
	}
	if first.Header().Get("Pragma") != "" || first.Header().Get("Expires") != "" {
		t.Fatalf("legacy no-store headers left on app.js: %v", first.Header())
	}
	again := do(h, "/next/app.js", "If-None-Match", first.Header().Get("ETag"))
	if again.Code != http.StatusNotModified || again.Body.Len() != 0 {
		t.Fatalf("revalidation: want 304 with no body, got %d (%d bytes)", again.Code, again.Body.Len())
	}
	if other := do(h, "/next/app.js", "If-None-Match", `"stale"`); other.Code != 200 {
		t.Fatalf("a different ETag must get the file, got %d", other.Code)
	}
}

func TestNextGzipRespectsAcceptEncoding(t *testing.T) {
	_, h := nextFixture(t)
	plain := do(h, "/next/app.js")
	if plain.Header().Get("Content-Encoding") != "" || plain.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("no Accept-Encoding must get the plain file: %v", plain.Header())
	}
	for _, refuse := range []string{"identity", "br", "gzip;q=0", "gzip; q=0.0"} {
		if do(h, "/next/app.js", "Accept-Encoding", refuse).Header().Get("Content-Encoding") != "" {
			t.Errorf("Accept-Encoding %q must not be gzipped", refuse)
		}
	}
	gz := do(h, "/next/app.js", "Accept-Encoding", "br, gzip;q=0.8")
	if gz.Header().Get("Content-Encoding") != "gzip" || gz.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
		t.Fatalf("want gzip javascript, got %v", gz.Header())
	}
	zr, err := gzip.NewReader(gz.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(zr)
	if string(body) != plain.Body.String() || gz.Body.Len() >= plain.Body.Len() {
		t.Fatalf("gzip body must decode to the file and be smaller (%d vs %d)", gz.Body.Len(), plain.Body.Len())
	}
	if gz.Header().Get("ETag") == plain.Header().Get("ETag") {
		t.Fatal("the gzip and plain bodies must not share an ETag")
	}
	if again := do(h, "/next/app.js", "Accept-Encoding", "gzip", "If-None-Match", gz.Header().Get("ETag")); again.Code != 304 {
		t.Fatalf("gzip revalidation: got %d", again.Code)
	}
	// fonts are woff2 already: sent as they are
	if f := do(h, "/next/fonts/a.woff2", "Accept-Encoding", "gzip"); f.Header().Get("Content-Encoding") != "" || f.Body.String() != "woff2 bytes" {
		t.Fatalf("font must not be gzipped: %v", f.Header())
	}
}

func TestNextVendorAndFontsAreImmutableOnlyAtTheirVersionedURL(t *testing.T) {
	dir, h := nextFixture(t)
	page := do(h, "/next/").Body.String()
	v := strings.TrimSuffix(strings.SplitN(page, "?v=", 2)[1][:10], `"`)
	if !strings.Contains(page, "./vendor/p.module.js?v="+v) {
		t.Fatalf("index.html must point at the versioned vendor file: %s", page)
	}
	if css := do(h, "/next/app.css").Body.String(); !strings.Contains(css, "url(fonts/a.woff2?v="+v+")") || !strings.Contains(css, "body{margin:0}") {
		t.Fatalf("app.css must point at the versioned font: %s", css)
	}
	for _, p := range []string{"/next/vendor/p.module.js", "/next/fonts/a.woff2"} {
		if got := do(h, p+"?v="+v).Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Errorf("%s?v=%s: Cache-Control %q", p, v, got)
		}
		if got := do(h, p).Header().Get("Cache-Control"); got != "no-cache" {
			t.Errorf("%s without a version must revalidate, got %q", p, got)
		}
		if got := do(h, p+"?v=other").Header().Get("Cache-Control"); got != "no-cache" {
			t.Errorf("%s with the wrong version must revalidate, got %q", p, got)
		}
	}
	// changing a vendored file changes the version, so an upgrade busts the cache
	writeFile(t, filepath.Join(dir, "next", "vendor", "p.module.js"), "export const p = 2;")
	if page2 := do(newWebRootHandler(dir), "/next/").Body.String(); page2 == page {
		t.Fatal("version did not change after a vendor file did")
	}
}

func TestNextEditedFileIsServedAgain(t *testing.T) {
	dir, h := nextFixture(t)
	old := do(h, "/next/app.js")
	writeFile(t, filepath.Join(dir, "next", "app.js"), "export const changed = true;")
	if got := do(h, "/next/app.js"); got.Body.String() != "export const changed = true;" || got.Header().Get("ETag") == old.Header().Get("ETag") {
		t.Fatalf("an edited file must be re-read: %q", got.Body.String())
	}
}

func TestNextRejectsEscapesAndKeepsMissingAs404(t *testing.T) {
	_, h := nextFixture(t)
	if rec := do(h, "/next/missing.js"); rec.Code != 404 {
		t.Fatalf("missing: %d", rec.Code)
	}
	if rec := do(h, "/next/../index.html"); strings.Contains(rec.Body.String(), "classic home") {
		t.Fatal("escaped /next")
	}
	if rec := do(h, "/next/vendor"); rec.Code == 200 {
		t.Fatalf("a directory is not a file: %d", rec.Code)
	}
}

func TestClassicAndOtherFilesAreUntouched(t *testing.T) {
	_, h := nextFixture(t)
	for _, p := range []string{"/classic", "/index.html", "/css/style.css"} {
		rec := do(h, p, "Accept-Encoding", "gzip")
		if rec.Code != 200 || rec.Header().Get("Cache-Control") != "" || rec.Header().Get("Content-Encoding") != "" || rec.Header().Get("Vary") != "" {
			t.Errorf("%s must be served as before, got %d %v", p, rec.Code, rec.Header())
		}
	}
}

func TestOuterWrapperStillSetsNoStoreForClassic(t *testing.T) {
	_, h := nextFixture(t)
	wrapped := DisableCachingAndSniffing(h)
	if cc := do(wrapped, "/classic").Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Fatalf("classic: %q", cc)
	}
	nx := do(wrapped, "/next/app.js")
	if nx.Header().Get("Cache-Control") != "no-cache" || nx.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("/next: %v", nx.Header())
	}
}
