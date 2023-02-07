package vars

import (
	"encoding/json"
	"os"

	"github.com/kercre123/chipper/pkg/logger"
)

// there should be a way to create a JSON configuration for wire-pod, rather than using env vars

const ApiConfigPath = "./apiConfig.json"

var APIConfig apiConfig

type apiConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable      bool   `json:"enable"`
		Provider    string `json:"provider"`
		Key         string `json:"key"`
		ID          string `json:"id"`
		IntentGraph bool   `json:"intentgraph"`
	} `json:"knowledge"`
	STT struct {
		Service  string `json:"provider"`
		Language string `json:"language"`
	}
}

func WriteConfigToDisk() {
	logger.Println("Configuration changed, writing to disk")
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func CreateConfigFromEnv() {
	// if no config exists, create it
	if os.Getenv("WEATHERAPI_ENABLED") == "true" {
		APIConfig.Weather.Enable = true
		APIConfig.Weather.Provider = os.Getenv("WEATHERAPI_PROVIDER")
		APIConfig.Weather.Key = os.Getenv("WEATHERAPI_KEY")
		APIConfig.Weather.Unit = os.Getenv("WEATHERAPI_UNIT")
	} else {
		APIConfig.Weather.Enable = false
	}
	if os.Getenv("KNOWLEDGE_ENABLED") == "true" {
		APIConfig.Knowledge.Enable = true
		APIConfig.Knowledge.Provider = os.Getenv("KNOWLEDGE_PROVIDER")
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			APIConfig.Knowledge.ID = os.Getenv("KNOWLEDGE_ID")
		}
		APIConfig.Knowledge.Key = os.Getenv("KNOWLEDGE_KEY")
	} else {
		APIConfig.Knowledge.Enable = false
	}
	WriteSTT()
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func WriteSTT() {
	// was not part of the original code, so this is its own function
	// launched if stt not found in config
	APIConfig.STT.Service = os.Getenv("STT_SERVICE")
	if os.Getenv("STT_SERVICE") == "vosk" {
		APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
	}
}

func ReadConfig() {
	if _, err := os.Stat(ApiConfigPath); err != nil {
		CreateConfigFromEnv()
		logger.Println("API config JSON created")
	} else {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to read API config file")
			logger.Println(err)
			return
		}
		err = json.Unmarshal(configBytes, &APIConfig)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			return
		}
		if APIConfig.STT.Service != os.Getenv("STT_SERVICE") {
			WriteSTT()
			writeBytes, _ := json.Marshal(APIConfig)
			os.WriteFile(ApiConfigPath, writeBytes, 0644)
		}
		logger.Println("API config successfully read")
	}
}
