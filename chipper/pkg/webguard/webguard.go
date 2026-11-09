package webguard

import (
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// keeps random websites in the user's browser from poking the api
//
// WIREPOD_ALLOWED_HOSTS: extra Host names to accept (comma separated)
// WIREPOD_ALLOWED_ORIGINS: extra origins allowed to make cross-origin api calls, ex. "http://myapp.lan:3000"

var localSuffixes = []string{
	".local",
	".localhost",
	".localdomain",
	".lan",
	".home",
	".home.arpa",
	".internal",
}

var extraHosts = splitEnv("WIREPOD_ALLOWED_HOSTS")
var extraOrigins = splitEnv("WIREPOD_ALLOWED_ORIGINS")

func splitEnv(name string) []string {
	var out []string
	for _, s := range strings.Split(os.Getenv(name), ",") {
		s = strings.ToLower(strings.TrimRight(strings.TrimSpace(s), "/"))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stripPort(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	return strings.ToLower(strings.Trim(host, "[]"))
}

// ips, single-label names and non-public tlds are fine
func HostAllowed(hostport string) bool {
	if hostport == "" {
		return true
	}
	host := strings.TrimSuffix(stripPort(hostport), ".")
	if net.ParseIP(host) != nil {
		return true
	}
	if host == "localhost" || !strings.Contains(host, ".") {
		return true
	}
	for _, s := range localSuffixes {
		if strings.HasSuffix(host, s) {
			return true
		}
	}
	for _, h := range extraHosts {
		if host == h {
			return true
		}
	}
	return false
}

func extraOriginAllowed(origin string) bool {
	origin = strings.ToLower(origin)
	for _, o := range extraOrigins {
		if origin == o {
			return true
		}
	}
	return false
}

func originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" {
		if origin == "null" {
			return false
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		return extraOriginAllowed(origin)
	}
	// no Origin (same-origin GETs, <img>, navigations): fall back to fetch metadata
	switch r.Header.Get("Sec-Fetch-Site") {
	case "cross-site", "same-site":
		return false
	}
	return true
}

func isGuardedPath(p string) bool {
	return strings.HasPrefix(p, "/api") || strings.HasPrefix(p, "/cam-stream")
}

func Guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !HostAllowed(r.Host) {
			http.Error(w, "forbidden host. if this is intentional, add it to WIREPOD_ALLOWED_HOSTS", http.StatusForbidden)
			return
		}
		if !isGuardedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		origin := r.Header.Get("Origin")
		if origin != "" && extraOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		if !originAllowed(r) {
			http.Error(w, "cross-origin request blocked. if this is intentional, add the origin to WIREPOD_ALLOWED_ORIGINS", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// json content-type forces a cors preflight, so a cross-site form/text post can't sneak in
func RequireJSON(w http.ResponseWriter, r *http.Request) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		http.Error(w, "content-type must be application/json", http.StatusUnsupportedMediaType)
		return false
	}
	return true
}
