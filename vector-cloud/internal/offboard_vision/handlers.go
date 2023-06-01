package offboard_vision

import "net/http"

var devHandlers func(*http.ServeMux)

func GetDevHandlers(s *http.ServeMux) {
	if devHandlers != nil {
		devHandlers(s)
	}
}
