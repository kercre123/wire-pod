package main

import (
	"net/http"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
)

type WrappedResponseWriter struct {
	http.ResponseWriter
	tag string
}

func (wrap *WrappedResponseWriter) WriteHeader(code int) {
	log.Printf("%s.WriteHeader(): %d\n", wrap.tag, code)
	wrap.ResponseWriter.WriteHeader(code)
}

func (wrap *WrappedResponseWriter) Write(b []byte) (int, error) {
	if wrap.tag == "json" && len(b) > 1 {
		log.Printf("%s.Write(): %s\n", wrap.tag, string(b))
	} else if wrap.tag == "grpc" && len(b) > 1 {
		log.Printf("%s.Write(): [% x]\n", wrap.tag, b)
	}
	return wrap.ResponseWriter.Write(b)
}

func (wrap *WrappedResponseWriter) Header() http.Header {
	h := wrap.ResponseWriter.Header()
	log.Printf("%s.Header(): %+v\n", wrap.tag, h)
	return h
}

func (wrap *WrappedResponseWriter) Flush() {
	wrap.ResponseWriter.(http.Flusher).Flush()
}

func (wrap *WrappedResponseWriter) CloseNotify() <-chan bool {
	return wrap.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func LogRequest(r *http.Request, tag string) {
	log.Printf("%s.Request: %+v\n", tag, r)
}
