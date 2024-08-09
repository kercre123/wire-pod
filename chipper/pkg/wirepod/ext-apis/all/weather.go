package extapis

import (
	"encoding/json"
	"errors"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const (
	// conditions
	WEATHER_SUNNY   = "Sunny"
	WEATHER_STARS   = "Stars"
	WEATHER_COUDY   = "Cloudy"
	WEATHER_RAIN    = "Rain"
	WEATHER_SNOW    = "Snow"
	WEATHER_WIND    = "Windy"
	WEATHER_THUNDER = "Thunderstorms"
)

type WeatherAPIStore struct {
	Name         string      `json:"name"`
	Link         string      `json:"dashlink"`
	NeedsPayment bool        `json:"needspayment"`
	APIAddr      string      `json:"apiaddr"`
	GeoAddr      string      `json:"geoaddr"`
	NeedsGeo     bool        `json:"needsgeo"`
	Structure    interface{} `json:"weatherstructure"`
	GeoStructure interface{} `json:"geostructure"`
}

type WeatherAPI interface {
	WeatherAPIStore
	GetWeather(Coordinates) WeatherConditions
	GetCoordinates() Coordinates
	GetWeatherFull(esn string) WeatherConditions
}

// get bot's location from JDocs, usually in the format of:
// San Francisco, California, United States of America
func GetLocation(esn string) (string, error) {
	jdoc, exists := vars.GetJdoc("vic:"+esn, "vic.RobotSettings")
	if !exists {
		return "", errors.New("jdoc does not exist")
	}
	var jdocSettings vars.RobotSettings
	json.Unmarshal([]byte(jdoc.JsonDoc), &jdocSettings)
	return jdocSettings.DefaultLocation, nil
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type WeatherConditions struct {
	Condition string `json:"condition"`
	TempF     int    `json:"tempf"`
	TempC     int    `json:"tempc"`
}
