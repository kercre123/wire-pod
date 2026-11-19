package vars

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// a way to create a JSON configuration for wire-pod, rather than the use of env vars

var ApiConfigPath = "./apiConfig.json"

var (
	cfgMu  sync.RWMutex
	apiCfg APIConfig
)

type APIConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable                 bool    `json:"enable"`
		Provider               string  `json:"provider"`
		Key                    string  `json:"key"`
		ID                     string  `json:"id"`
		Model                  string  `json:"model"`
		IntentGraph            bool    `json:"intentgraph"`
		RobotName              string  `json:"robotName"`
		OpenAIPrompt           string  `json:"openai_prompt"`
		OpenAIVoice            string  `json:"openai_voice"`
		OpenAIVoiceWithEnglish bool    `json:"openai_voice_with_english"`
		SaveChat               bool    `json:"save_chat"`
		CommandsEnable         bool    `json:"commands_enable"`
		Endpoint               string  `json:"endpoint"`
		TopP                   float32 `json:"top_p"`
		Temperature            float32 `json:"temp"`
	} `json:"knowledge"`
	STT struct {
		Service       string `json:"provider"`
		Language      string `json:"language"`
		UseLargeModel bool   `json:"use_large_model"`
	} `json:"STT"`
	Server struct {
		// false for ip, true for escape pod
		EPConfig bool   `json:"epconfig"`
		Port     string `json:"port"`
	} `json:"server"`
	HasReadFromEnv   bool `json:"hasreadfromenv"`
	PastInitialSetup bool `json:"pastinitialsetup"`
}

func Config() APIConfig {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return apiCfg
}

func UpdateConfig(fn func(c *APIConfig)) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	fn(&apiCfg)
}

func WriteConfigToDisk() {
	logger.Println("Configuration changed, writing to disk")
	cfgMu.Lock()
	defer cfgMu.Unlock()
	writeBytes, _ := json.Marshal(apiCfg)
	WriteFileAtomic(ApiConfigPath, writeBytes, 0600)
}

func CreateConfigFromEnv() {
	// if no config exists, create it
	if os.Getenv("WEATHERAPI_ENABLED") == "true" {
		apiCfg.Weather.Enable = true
		apiCfg.Weather.Provider = os.Getenv("WEATHERAPI_PROVIDER")
		apiCfg.Weather.Key = os.Getenv("WEATHERAPI_KEY")
		apiCfg.Weather.Unit = os.Getenv("WEATHERAPI_UNIT")
	} else {
		apiCfg.Weather.Enable = false
	}
	if os.Getenv("KNOWLEDGE_ENABLED") == "true" {
		apiCfg.Knowledge.Enable = true
		apiCfg.Knowledge.Provider = os.Getenv("KNOWLEDGE_PROVIDER")
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			apiCfg.Knowledge.ID = os.Getenv("KNOWLEDGE_ID")
		}
		apiCfg.Knowledge.Key = os.Getenv("KNOWLEDGE_KEY")
	} else {
		apiCfg.Knowledge.Enable = false
	}
	WriteSTT()
	apiCfg.HasReadFromEnv = true
	writeBytes, _ := json.Marshal(apiCfg)
	WriteFileAtomic(ApiConfigPath, writeBytes, 0600)
}

func WriteSTT() {
	// was not part of the original code, so this is its own function
	// launched if stt not found in config
	apiCfg.STT.Service = os.Getenv("STT_SERVICE")
	if os.Getenv("STT_SERVICE") == "vosk" || os.Getenv("STT_SERVICE") == "whisper.cpp" {
		apiCfg.STT.Language = os.Getenv("STT_LANGUAGE")
		if os.Getenv("VOSK_LARGE") == "true" {
            apiCfg.STT.UseLargeModel = true
        }
	}
}

func ReadConfig() {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if _, err := os.Stat(ApiConfigPath); err != nil {
		CreateConfigFromEnv()
		logger.Println("API config JSON created")
	} else {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			apiCfg.Knowledge.Enable = false
			apiCfg.Weather.Enable = false
			logger.Println("Failed to read API config file")
			logger.Println(err)
			return
		}
		err = json.Unmarshal(configBytes, &apiCfg)
		if err != nil {
			apiCfg.Knowledge.Enable = false
			apiCfg.Weather.Enable = false
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			return
		}
		// stt service is the only thing controlled by shell
		if apiCfg.STT.Service != os.Getenv("STT_SERVICE") {
			WriteSTT()
		}
		if !apiCfg.HasReadFromEnv {
			if apiCfg.Server.Port != os.Getenv("DDL_RPC_PORT") {
				apiCfg.HasReadFromEnv = true
				apiCfg.PastInitialSetup = true
			}
		}

		if apiCfg.Knowledge.Model == "meta-llama/Llama-2-70b-chat-hf" {
			logger.Println("Setting Together model to Llama3")
			apiCfg.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
		}

		writeBytes, _ := json.Marshal(apiCfg)
		WriteFileAtomic(ApiConfigPath, writeBytes, 0600)
		logger.Println("API config successfully read")
	}
}
