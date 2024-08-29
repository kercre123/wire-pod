package vars

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// a way to create a JSON configuration for wire-pod, rather than the use of env vars

var ApiConfigPath = "./apiConfig.json"

var APIConfig apiConfig
var legacyConf legacyAPIConfig

type apiConfig struct {
	STT struct {
		Service  string `json:"provider"`
		Language string `json:"language"`
	} `json:"STT"`
	Server struct {
		// false for ip, true for escape pod
		EPConfig bool   `json:"epconfig"`
		Port     string `json:"port"`
	} `json:"server"`
	PastInitialSetup bool `json:"pastinitialsetup"`
}

func WriteConfigToDisk() {
	logger.Println("Configuration changed, writing to disk")
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func ReadConfig() {
	if _, err := os.Stat(ApiConfigPath); err == nil {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			logger.Println("Failed to read API config file")
			logger.Println(err)
			return
		}
		// if api config is legacy
		if strings.Contains(string(ApiConfigPath), `{"weather":`) {
			err = json.Unmarshal(configBytes, &legacyConf)
			if err != nil {
				logger.Println("Legacy API configuration detected, but couldn't unmarshal. Errroing")
				logger.Println(err)
				os.Exit(1)
			}
			logger.Println("Legacy API configuration detected. Migrating weather and knowledge API details")
			MigrateLegacyConfs()

		}
		err = json.Unmarshal(configBytes, &APIConfig)
		if err != nil {
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			return
		}
		// stt service is the only thing controlled by shell
		APIConfig.STT.Service = os.Getenv("STT_SERVICE")

		writeBytes, _ := json.Marshal(APIConfig)
		os.WriteFile(ApiConfigPath, writeBytes, 0644)
		logger.Println("API config successfully read")
	}
}
