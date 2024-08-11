package vars

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
)

var ExtAPIConfs ExternalAPIConfigs
var ExternalApiConfigPath = "./externalAPIConfigs.json"
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

type ExternalAPIConfigs struct {
	// uuids
	DefaultWeather string          `json:"defaultweather"`
	DefaultKG      string          `json:"defaultkg"`
	DefaultIG      string          `json:"defaultig"`
	DefaultTTS     string          `json:"defaulttts"`
	Weathers       []WeatherConfig `json:"weathers"`
	KGs            []KGConfig      `json:"kgs"`
	TTSes          []TTSConfig     `json:"ttses"`
}

func SaveExternalAPIConfigs() {
	marshalled, _ := json.Marshal(ExtAPIConfs)
	os.WriteFile(ExternalApiConfigPath, marshalled, 0777)
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

}

func MigrateAPIConfigs() {
	if APIConfig.Knowledge.Enable {
		var hid string
		var hkey string
		var kapikey string
		var igcapable bool
		if APIConfig.Knowledge.ID != "" {
			hid = APIConfig.Knowledge.ID
			hkey = APIConfig.Knowledge.Key
			igcapable = false
		} else if APIConfig.Knowledge.Key != "" {
			kapikey = APIConfig.Knowledge.Key
			igcapable = true
		}
		kgId := AddKGConfig(
			KGConfig{
				Provider:          APIConfig.Knowledge.Provider,
				APIKey:            kapikey,
				HoundifyClientID:  hid,
				HoundifyClientKey: hkey,
				LLMSpecifics: KGLLMSpecifics{
					Model:              APIConfig.Knowledge.Model,
					Endpoint:           APIConfig.Knowledge.Endpoint,
					Prompt:             APIConfig.Knowledge.OpenAIPrompt,
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
	if APIConfig.Weather.Enable {
		wId := AddWeatherConfig(
			WeatherConfig{
				Provider:   APIConfig.Weather.Provider,
				APIKey:     APIConfig.Weather.Key,
				Functional: true,
			},
		)
		SetWeatherDefault(wId)
	}
	if APIConfig.Knowledge.Provider == "openai" {
		ttsId := AddTTSConfig(
			TTSConfig{
				Provider:   "openai",
				APIKey:     APIConfig.Knowledge.Key,
				Voice:      APIConfig.Knowledge.OpenAIVoice,
				Functional: true,
			},
		)
		SetTTSDefault(ttsId)
	}
}
