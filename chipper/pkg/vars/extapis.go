package vars

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
)

// TODO:
// Set and Get functions for per-bot configs
// make sure Delete function goes through each per-bot config and resets the default in case the deleted uuid is the bot's default

var ExtAPIConfs ExternalAPIConfigs
var ExternalAPIConfigsPath = "./externalAPIConfigs.json"
var ExtAPIMutex sync.Mutex

type WeatherConfig struct {
	Provider   string `json:"provider"`
	ID         string `json:"uuid"`
	APIKey     string `json:"apikey"`
	GeoKey     string `json:"geokey"`
	Functional bool   `json:"functional"`
}

type KGConfig struct {
	Provider          string         `json:"provider"`
	ID                string         `json:"uuid"`
	APIKey            string         `json:"apikey"`
	Functional        bool           `json:"functional"`
	HoundifyClientKey string         `json:"houndclientkey"`
	HoundifyClientID  string         `json:"hounclientid"`
	LLMSpecifics      KGLLMSpecifics `json:"llmconfig"`
}

type KGLLMSpecifics struct {
	Model              string `json:"model"`
	Endpoint           string `json:"endpoint"`
	Prompt             string `json:"prompt"`
	IntentGraphCapable bool   `json:"igcapable"`
}

type TTSConfig struct {
	Provider string `json:"provider"`
	ID       string `json:"uuid"`
	// if needed
	APIKey     string `json:"apikey"`
	Voice      string `json:"string"`
	Functional bool   `json:"functional"`
}

type PerBotConfig struct {
	ESN       string `json:"esn"`
	KGID      string `json:"kgid"`
	IGID      string `json:"igid"`
	WeatherID string `json:"weatherid"`
	STTID     string `json:"sttid"`
}

type ExternalAPIConfigs struct {
	// uuids
	DefaultWeather string          `json:"defaultweather"`
	DefaultKG      string          `json:"defaultkg"`
	DefaultIG      string          `json:"defaultig"`
	DefaultTTS     string          `json:"defaulttts"`
	Weathers       []WeatherConfig `json:"weathers"`
	KGs            []KGConfig      `json:"kgs"`
	TTSes          []TTSConfig     `json:"ttses"`
	BotConfigs     []PerBotConfig  `json:"botconfigs"`
}

func SaveExternalAPIConfigs() {
	marshalled, _ := json.Marshal(ExtAPIConfs)
	os.WriteFile(ExternalAPIConfigsPath, marshalled, 0777)
}

func GetUUID() string {
	id := uuid.New()
	return id.String()
}

// this should only be run if a Test was successful
func AddWeatherConfig(wc WeatherConfig) (uuid string) {
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	id := GetUUID()
	wc.ID = id
	wc.Functional = true
	ExtAPIConfs.Weathers = append(ExtAPIConfs.Weathers, wc)
	if len(ExtAPIConfs.Weathers) == 1 {
		ExtAPIConfs.DefaultWeather = id
	}
	SaveExternalAPIConfigs()
	return id
}

func AddKGConfig(kgc KGConfig) (uuid string) {
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	id := GetUUID()
	kgc.ID = id
	kgc.Functional = true
	ExtAPIConfs.KGs = append(ExtAPIConfs.KGs, kgc)
	if len(ExtAPIConfs.KGs) == 1 {
		ExtAPIConfs.DefaultKG = id
		if kgc.LLMSpecifics.IntentGraphCapable {
			ExtAPIConfs.DefaultIG = id
		}
	}
	SaveExternalAPIConfigs()
	return id
}

// this should only be run if a Test was successful
func AddTTSConfig(ttsc TTSConfig) (uuid string) {
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	id := GetUUID()
	ttsc.ID = id
	ttsc.Functional = true
	ExtAPIConfs.TTSes = append(ExtAPIConfs.TTSes, ttsc)
	if len(ExtAPIConfs.TTSes) == 1 {
		ExtAPIConfs.DefaultTTS = id
	}
	SaveExternalAPIConfigs()
	return id
}

func SetWeatherDefault(uuid string) {
	ExtAPIMutex.Lock()
	ExtAPIConfs.DefaultWeather = uuid
	ExtAPIMutex.Unlock()
}

func SetKGDefault(uuid string) {
	ExtAPIMutex.Lock()
	ExtAPIConfs.DefaultKG = uuid
	ExtAPIMutex.Unlock()
}

func SetIGDefault(uuid string) {
	ExtAPIMutex.Lock()
	ExtAPIConfs.DefaultIG = uuid
	ExtAPIMutex.Unlock()
}

func SetTTSDefault(uuid string) {
	ExtAPIMutex.Lock()
	ExtAPIConfs.DefaultTTS = uuid
	ExtAPIMutex.Unlock()
}

func SetPerBotKG(esn string, uuid string) {
	ExtAPIMutex.Lock()
	for i, rob := range ExtAPIConfs.BotConfigs {
		if rob.ESN == esn {
			ExtAPIConfs.BotConfigs[i].KGID = uuid
		}
	}
	ExtAPIMutex.Unlock()
}

func GetExtAPIConfigs() ExternalAPIConfigs {
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	return ExtAPIConfs
}

func DeleteAPIConfig(uuid string) {
	var newExtAPIs ExternalAPIConfigs
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	for _, kg := range ExtAPIConfs.KGs {
		if kg.ID != uuid {
			newExtAPIs.KGs = append(newExtAPIs.KGs, kg)
		}
	}
	for _, ttsc := range ExtAPIConfs.TTSes {
		if ttsc.ID != uuid {
			newExtAPIs.TTSes = append(newExtAPIs.TTSes, ttsc)
		}
	}
	for _, wc := range ExtAPIConfs.Weathers {
		if wc.ID != uuid {
			newExtAPIs.Weathers = append(newExtAPIs.Weathers, wc)
		}
	}
	if uuid == ExtAPIConfs.DefaultTTS {
		if len(newExtAPIs.TTSes) > 0 {
			newExtAPIs.DefaultTTS = newExtAPIs.TTSes[0].ID
		} else {
			newExtAPIs.DefaultTTS = ""
		}
	}
	if uuid == ExtAPIConfs.DefaultWeather {
		if len(newExtAPIs.Weathers) > 0 {
			newExtAPIs.DefaultWeather = newExtAPIs.Weathers[0].ID
		} else {
			newExtAPIs.DefaultWeather = ""
		}
	}
	if uuid == ExtAPIConfs.DefaultKG {
		if len(newExtAPIs.KGs) > 0 {
			newExtAPIs.DefaultKG = newExtAPIs.KGs[0].ID
			if newExtAPIs.KGs[0].LLMSpecifics.IntentGraphCapable {
				newExtAPIs.DefaultIG = newExtAPIs.KGs[0].ID
			} else {
				newExtAPIs.DefaultIG = ""
			}
		} else {
			newExtAPIs.DefaultKG = ""
			newExtAPIs.DefaultIG = ""
		}
	}
	for i, rob := range ExtAPIConfs.BotConfigs {
		if rob.IGID == uuid {
			if len(newExtAPIs.KGs) > 0 {
				if ExtAPIConfs.B
				ExtAPIConfs.BotConfigs[i].IGID == newExt
			}
		}
	}

}

func MigrateLegacyConfs() {
	if legacyConf.Knowledge.Enable {
		var hid string
		var hkey string
		var kapikey string
		var igcapable bool
		if legacyConf.Knowledge.ID != "" {
			hid = legacyConf.Knowledge.ID
			hkey = legacyConf.Knowledge.Key
			igcapable = false
		} else if legacyConf.Knowledge.Key != "" {
			kapikey = legacyConf.Knowledge.Key
			igcapable = true
		}
		kgId := AddKGConfig(
			KGConfig{
				Provider:          legacyConf.Knowledge.Provider,
				APIKey:            kapikey,
				HoundifyClientID:  hid,
				HoundifyClientKey: hkey,
				LLMSpecifics: KGLLMSpecifics{
					Model:              legacyConf.Knowledge.Model,
					Endpoint:           legacyConf.Knowledge.Endpoint,
					Prompt:             legacyConf.Knowledge.OpenAIPrompt,
					IntentGraphCapable: igcapable,
				},
				Functional: true,
			},
		)
		SetKGDefault(kgId)
		if igcapable {
			SetIGDefault(kgId)
		}
	}
	if legacyConf.Weather.Enable {
		wId := AddWeatherConfig(
			WeatherConfig{
				Provider:   legacyConf.Weather.Provider,
				APIKey:     legacyConf.Weather.Key,
				Functional: true,
			},
		)
		SetWeatherDefault(wId)
	}
	if legacyConf.Knowledge.Provider == "openai" {
		ttsId := AddTTSConfig(
			TTSConfig{
				Provider:   "openai",
				APIKey:     legacyConf.Knowledge.Key,
				Voice:      legacyConf.Knowledge.OpenAIVoice,
				Functional: true,
			},
		)
		SetTTSDefault(ttsId)
	}
}

type legacyAPIConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable                 bool   `json:"enable"`
		Provider               string `json:"provider"`
		Key                    string `json:"key"`
		ID                     string `json:"id"`
		Model                  string `json:"model"`
		IntentGraph            bool   `json:"intentgraph"`
		RobotName              string `json:"robotName"`
		OpenAIPrompt           string `json:"openai_prompt"`
		OpenAIVoice            string `json:"openai_voice"`
		OpenAIVoiceWithEnglish bool   `json:"openai_voice_with_english"`
		SaveChat               bool   `json:"save_chat"`
		CommandsEnable         bool   `json:"commands_enable"`
		Endpoint               string `json:"endpoint"`
	} `json:"knowledge"`
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
