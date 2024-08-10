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

type WeatherAPI struct {
	Name          string
	Link          string
	NeedsPayment  bool
	APIAddr       string
	GeoAddr       string
	NeedsGeo      bool
	Structure     interface{}
	GeoStructure  interface{}
	Meteorologist WeatherAPIer
}

type WeatherAPIer interface {
	/*
		functions you need to define
	*/
	GetWeatherWithLocation(string) WeatherConditions
	GetWeatherWithCoordinates(Coordinates) WeatherConditions
	GetCoordinates() Coordinates
	// test the API, see if API key is correct
	Test() (bool, error)

	/*
		functions you DON'T need to define
	*/
	// actual entry-point function. this is already defined
	// if NeedsGeo is false, it won't use GetWeatherWithCoordinates, vice versa
	GetWeather(string) WeatherConditions
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

func SaveAPIKey(provider string, key string) {

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

func NewWeatherAPI(name, link, apiaddr, geoaddr string, needspayment, needsgeo bool, apistruct, geostruct interface{}) WeatherAPI {
	return WeatherAPI{
		Name:         name,
		Link:         link,
		APIAddr:      apiaddr,
		GeoAddr:      geoaddr,
		NeedsPayment: needspayment,
		NeedsGeo:     needsgeo,
		Structure:    apistruct,
		GeoStructure: geostruct,
	}
}
