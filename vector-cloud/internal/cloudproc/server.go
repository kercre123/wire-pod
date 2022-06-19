package cloudproc

import (
	"net/http"

	"github.com/digital-dream-labs/vector-cloud/internal/token"
)

var addHandlerFunc func(func(*http.ServeMux), *token.Server)

func addHandlers(f func(*http.ServeMux), s *token.Server) {
	if addHandlerFunc != nil && f != nil && s != nil {
		addHandlerFunc(f, s)
	}
}
