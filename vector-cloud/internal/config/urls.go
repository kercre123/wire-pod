package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// URLs represents a set of URLs where Anki's cloud services can be reached
type URLs struct {
	JDocs          string  `json:"jdocs"`
	Token          string  `json:"tms"`
	Chipper        string  `json:"chipper"`
	Check          string  `json:"check"`
	LogFiles       string  `json:"logfiles"`
	AppKey         string  `json:"appkey"`
	OffboardVision *string `json:"offboard_vision,omitempty"`
}

// DefaultURLs provides a default, hard-coded configuration that can be used
// if an expected configuration is not found on disk
var DefaultURLs = URLs{
	JDocs:    "jdocs-dev.api.anki.com:443",
	Token:    "token-dev.api.anki.com:443",
	Chipper:  "chipper-dev.api.anki.com:443",
	Check:    "conncheck.global.anki-dev-services.com/ok",
	LogFiles: "s3://anki-device-logs-dev/victor",
	AppKey:   "",
}

// Env represents the URLs associated with the most recent successful call to
// SetGlobal. Before this, it has the same values as DefaultURLs.
var Env = DefaultURLs

// SetGlobal sets the public Env variable to the URLs in the given filename. If the given
// filename is blank, a known hardcoded location for server_config.json on the robot is used.
func SetGlobal(filename string) error {
	urls, err := LoadURLs(filename)
	if err != nil {
		return err
	}
	Env = *urls
	return nil
}

var defaultFilename = "/anki/data/assets/cozmo_resources/config/server_config.json"
var wirepodFilename = "/data/data/server_config.json"

// LoadURLs attempts to load a URL config from the given filename. If the given filename
// is blank, a known hardcoded location for server_config.json on the robot is used.
func LoadURLs(filename string) (*URLs, error) {
	if filename == "" {
		if _, err := os.Open(wirepodFilename); err != nil {
			filename = defaultFilename
		} else {
			filename = wirepodFilename
		}
	}
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var urls URLs
	if err := json.Unmarshal(buf, &urls); err != nil {
		return nil, err
	}
	return &urls, nil
}
