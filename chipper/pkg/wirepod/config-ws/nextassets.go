package webserver

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The new web UI (/next) is meant to be light on low-end hardware such as a Pi: the browser keeps its files and asks before reusing
// them (a 304 instead of ~350 KB), and text is gzipped once and kept in memory. The classic UI and the API
// are left exactly as they were.
//
//   - vendor/ and fonts/ are cached for a year. Their URLs carry ?v=<hash of those files>, written into
//     index.html and app.css as they are served, so an upgrade that changes one changes every URL.
//   - everything else is "no-cache": kept, but revalidated with its ETag before each use.
//   - fonts are woff2 already and are sent as they are.

type nextAsset struct {
	modTime   time.Time
	size      int64
	plain, gz []byte
	etag      string
}

type nextAssets struct {
	dir     string       // the web root
	files   http.Handler // serves everything this does not
	version string
	mu      sync.Mutex
	cache   map[string]*nextAsset
}

var (
	textAsset   = map[string]bool{".html": true, ".js": true, ".mjs": true, ".css": true, ".svg": true, ".json": true, ".txt": true, ".md": true}
	versionedAt = regexp.MustCompile(`(vendor/[\w.-]+\.js|fonts/[\w.-]+\.woff2)`)
)

func newNextAssets(dir string, files http.Handler) *nextAssets {
	n := &nextAssets{dir: filepath.Join(dir, "next"), files: files, cache: map[string]*nextAsset{}}
	h := sha256.New()
	for _, sub := range []string{"vendor", "fonts"} {
		filepath.WalkDir(filepath.Join(n.dir, sub), func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				if b, err := os.ReadFile(p); err == nil {
					h.Write([]byte(d.Name()))
					h.Write(b)
				}
			}
			return nil
		})
	}
	n.version = hex.EncodeToString(h.Sum(nil))[:10]
	return n
}

func (n *nextAssets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(r.URL.Path, "/next/")), "/")
	if rel == "" {
		rel = "index.html"
	}
	h := w.Header()
	h.Del("Pragma")
	h.Del("Expires")
	h.Set("Cache-Control", "no-cache")
	if (strings.HasPrefix(rel, "vendor/") || strings.HasPrefix(rel, "fonts/")) && r.URL.Query().Get("v") == n.version {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	var a *nextAsset
	if (r.Method == http.MethodGet || r.Method == http.MethodHead) && textAsset[path.Ext(rel)] {
		a = n.load(rel)
	}
	if a == nil {
		n.files.ServeHTTP(w, r)
		return
	}
	body, etag := a.plain, a.etag
	h.Set("Vary", "Accept-Encoding")
	if acceptsGzip(r) {
		body, etag = a.gz, a.etag[:len(a.etag)-1]+`-gz"`
		h.Set("Content-Encoding", "gzip")
	}
	h.Set("ETag", etag)
	ctype := mime.TypeByExtension(path.Ext(rel))
	if ctype == "" {
		ctype = "text/plain; charset=utf-8" // ServeContent would sniff the compressed bytes
	}
	h.Set("Content-Type", ctype)
	http.ServeContent(w, r, rel, a.modTime, bytes.NewReader(body))
}

// load returns the file ready to send, or nil when there is no such regular file. It is read again when
// its size or modification time changes, so editing the web root while wire-pod runs still works.
func (n *nextAssets) load(rel string) *nextAsset {
	p := filepath.Join(n.dir, filepath.FromSlash(rel))
	info, err := os.Stat(p)
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if a := n.cache[rel]; a != nil && a.size == info.Size() && a.modTime.Equal(info.ModTime()) {
		return a
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	if rel == "index.html" || rel == "app.css" {
		b = versionedAt.ReplaceAll(b, []byte("${1}?v="+n.version))
	}
	var z bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&z, gzip.BestCompression)
	zw.Write(b)
	zw.Close()
	sum := sha256.Sum256(b)
	a := &nextAsset{modTime: info.ModTime(), size: info.Size(), plain: b, gz: z.Bytes(), etag: `"` + hex.EncodeToString(sum[:8]) + `"`}
	n.cache[rel] = a
	return a
}

// acceptsGzip reports whether the request's Accept-Encoding allows gzip.
func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		name, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(name), "gzip") {
			continue
		}
		if q := strings.ReplaceAll(params, " ", ""); strings.HasPrefix(q, "q=") {
			if v, err := strconv.ParseFloat(q[2:], 64); err == nil && v <= 0 {
				return false
			}
		}
		return true
	}
	return false
}
