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
	Provider          string `json:"provider"`
	ID                string `json:"uuid"`
	APIKey            string `json:"apikey"`
	Functional        bool   `json:"functional"`
	HoundifyClientKey string `json:"houndclientkey"`
	HoundifyClientID  string `json:"hounclientid"`
	LLMSpecifics      struct {
		Model              string `json:"model"`
		Endpoint           string `json:"endpoint"`
		Prompt             string `json:"prompt"`
		IntentGraphCapable bool   `json:"igcapable"`
	} `json:"llmconfig"`
}

type ExternalAPIConfigs struct {
	// uuids
	DefaultWeather string          `json:"defaultweather"`
	DefaultKG      string          `json:"defaultkg"`
	DefaultIG      string          `json:"defaultig"`
	Weathers       []WeatherConfig `json:"weathers"`
	KGs            []KGConfig      `json:"kgs"`
}

func SaveExternalAPIConfigs() {
	ExtAPIMutex.Lock()
	marshalled, _ := json.Marshal(ExtAPIConfs)
	os.WriteFile(ExternalApiConfigPath, marshalled, 0777)
	ExtAPIMutex.Unlock()
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
	return id
}

func AddKGConfig(kgc KGConfig) (uuid string) {
	ExtAPIMutex.Lock()
	defer ExtAPIMutex.Unlock()
	id := GetUUID()
	kgc.ID = id
	kgc.Functional = true
	ExtAPIConfs.KGs = append(ExtAPIConfs.KGs, kgc)
	return id
}
