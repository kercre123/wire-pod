package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	lcztn "github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization"
)

/* TODO:
Create seperate functions for weatherAPI and openweathermap,
create a standard for how weather functions should be created
*/

// *** WEATHERAPI.COM ***

type weatherAPIResponseStruct struct {
	Location struct {
		Name      string `json:"name"`
		Localtime string `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
	} `json:"current"`
}
type weatherAPICladStruct []struct {
	APIValue string `json:"APIValue"`
	CladType string `json:"CladType"`
}

// *** OPENWEATHERMAP.ORG ***

type openWeatherMapAPIGeoCodingStruct struct {
	Name       string            `json:"name"`
	LocalNames map[string]string `json:"local_names"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
}

/*
//3.0 API, requires your credit card even to get 1k free requests per day

type openWeatherMapAPIResponseStruct struct {
    Lat      			float64 `json:"lat"`
	Lon					float64 `json:"lon"`
	timezone		  	string `json:"timezone"`
	timezone_offset	  	string `json:"timezone_offset"`
	Current struct {
		DT	 	 	int     `json:"dt"`
		Sunrise	 	int     `json:"sunrise"`
		Sunset	 	int     `json:"sunset"`
		Temp	    float64 `json:"temp"`
		FeelsLike   float64 `json:"feels_like"`
		Pressure	int     `json:"pressure"`
		Humidity	int     `json:"humidity"`
		DewPoint	float64 `json:"dew_point"`
		UVI	        float64 `json:"uvi"`
		Clouds	 	int     `json:"clouds"`
		Visibility	int     `json:"visibility"`
		WindSpeed	float64 `json:"wind_speed"`
		WindDeg	 	int     `json:"wid_deg"`
		WindGust	float64 `json:"wind_gust"`
		Weather        struct {
			Id	 		int    `json:"id"`
			Main 		string `json:"main"`
			Description string `json:"description"`
			Icon 		string `json:"icon"`
		} `json:"weather"`
	} `json:"current"`
}
*/

//2.5 API

type WeatherStruct struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type openWeatherMapAPIResponseStruct struct {
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Weather []WeatherStruct `json:"weather"`
	Base    string          `json:"base"`
	Main    struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	DT  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		Id      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

type openWeatherMapForecastAPIResponseStruct struct {
	Cod     string                            `json:"cod"`
	Message int                               `json:"message"`
	Cnt     int                               `json:"cnt"`
	List    []openWeatherMapAPIResponseStruct `json:"list"`
}

func getWeather(location string, botUnits string, hoursFromNow int) (string, string, string, string, string, string) {
	var weatherEnabled bool
	var condition string
	var is_forecast string
	var local_datetime string
	var speakable_location_string string
	var temperature string
	var temperature_unit string
	weatherAPIEnabled := vars.APIConfig.Weather.Enable
	weatherAPIKey := vars.APIConfig.Weather.Key
	weatherAPIUnit := vars.APIConfig.Weather.Unit
	weatherAPIProvider := vars.APIConfig.Weather.Provider
	if weatherAPIEnabled && weatherAPIKey != "" {
		weatherEnabled = true
		logger.Println("Weather API enabled")
	} else {
		weatherEnabled = false
		logger.Println("Weather API not enabled, using placeholder")
		if weatherAPIEnabled && weatherAPIKey == "" {
			logger.Println("Weather API enabled, but Weather API key not set")
		}
	}
	if weatherEnabled {
		if botUnits != "" {
			if botUnits == "F" {
				logger.Println("Weather units set to F")
				weatherAPIUnit = "F"
			} else if botUnits == "C" {
				logger.Println("Weather units set to C")
				weatherAPIUnit = "C"
			}
		} else if weatherAPIUnit != "F" && weatherAPIUnit != "C" {
			logger.Println("Weather API unit not set, using F")
			weatherAPIUnit = "F"
		}
	}

	if weatherEnabled {
		if weatherAPIProvider == "weatherapi.com" {
			params := url.Values{}
			params.Add("key", weatherAPIKey)
			params.Add("q", location)
			params.Add("aqi", "no")
			url := "http://api.weatherapi.com/v1/current.json"
			resp, err := http.PostForm(url, params)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			weatherResponse := string(body)
			var weatherAPICladMap weatherAPICladStruct
			mapPath := ""
			if runtime.GOOS == "android" {
				mapPath = vars.AndroidPath + "/weather-map.json"
			} else {
				mapPath = "./weather-map.json"
			}
			jsonFile, _ := os.ReadFile(mapPath)
			json.Unmarshal(jsonFile, &weatherAPICladMap)
			var weatherStruct weatherAPIResponseStruct
			json.Unmarshal([]byte(weatherResponse), &weatherStruct)
			var matchedValue bool
			for _, b := range weatherAPICladMap {
				if b.APIValue == weatherStruct.Current.Condition.Text {
					condition = b.CladType
					logger.Println("API Value: " + b.APIValue + ", Clad Type: " + b.CladType)
					matchedValue = true
					break
				}
			}
			if !matchedValue {
				condition = weatherStruct.Current.Condition.Text
			}
			is_forecast = "false"
			local_datetime = weatherStruct.Current.LastUpdated
			speakable_location_string = weatherStruct.Location.Name
			if weatherAPIUnit == "C" {
				temperature = strconv.Itoa(int(weatherStruct.Current.TempC))
				temperature_unit = "C"
			} else {
				temperature = strconv.Itoa(int(weatherStruct.Current.TempF))
				temperature_unit = "F"
			}
		} else if weatherAPIProvider == "openweathermap.org" {
			// First use geocoding api to convert location into coordinates
			// E.G. http://api.openweathermap.org/geo/1.0/direct?q={city name},{state code},{country code}&limit={limit}&appid={API key}
			url := "http://api.openweathermap.org/geo/1.0/direct?q=" + location + "&limit=1&appid=" + weatherAPIKey
			resp, err := http.Get(url)
			if err != nil {
				logger.Println(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			geoCodingResponse := string(body)

			var geoCodingInfoStruct []openWeatherMapAPIGeoCodingStruct

			err = json.Unmarshal([]byte(geoCodingResponse), &geoCodingInfoStruct)
			if err != nil {
				logger.Println(err)
			}
			if len(geoCodingInfoStruct) == 0 {
				logger.Println("Geo provided no response.")
				condition = "undefined"
				is_forecast = "false"
				local_datetime = "test"              // preferably local time in UTC ISO 8601 format ("2022-06-15 12:21:22.123")
				speakable_location_string = location // preferably the processed location
				temperature = "120"
				temperature_unit = "C"
				return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
			}
			Lat := fmt.Sprintf("%f", geoCodingInfoStruct[0].Lat)
			Lon := fmt.Sprintf("%f", geoCodingInfoStruct[0].Lon)

			logger.Println("Lat: " + Lat + ", Lon: " + Lon)
			logger.Println("Name: " + geoCodingInfoStruct[0].Name)
			logger.Println("Country: " + geoCodingInfoStruct[0].Country)

			// Now that we have Lat and Lon, let's query the weather
			units := "metric"
			if weatherAPIUnit == "F" {
				units = "imperial"
			}
			if hoursFromNow == 0 {
				url = "https://api.openweathermap.org/data/2.5/weather?lat=" + Lat + "&lon=" + Lon + "&units=" + units + "&appid=" + weatherAPIKey
			} else {
				url = "https://api.openweathermap.org/data/2.5/forecast?lat=" + Lat + "&lon=" + Lon + "&units=" + units + "&appid=" + weatherAPIKey
			}
			resp, err = http.Get(url)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ = io.ReadAll(resp.Body)
			weatherResponse := string(body)
			var openWeatherMapAPIResponse openWeatherMapAPIResponseStruct

			if hoursFromNow > 0 {
				// Forecast request: free API results are returned in 3 hours slots
				var openWeatherMapForecastAPIResponse openWeatherMapForecastAPIResponseStruct
				err = json.Unmarshal([]byte(weatherResponse), &openWeatherMapForecastAPIResponse)
				openWeatherMapAPIResponse = openWeatherMapForecastAPIResponse.List[hoursFromNow/3]
			} else {
				// Current weather request
				err = json.Unmarshal([]byte(weatherResponse), &openWeatherMapAPIResponse)
			}

			if err != nil {
				panic(err)
			}

			conditionCode := openWeatherMapAPIResponse.Weather[0].Id
			logger.Println(conditionCode)

			if conditionCode < 300 {
				// Thunderstorm
				condition = "Thunderstorms"
			} else if conditionCode < 400 {
				// Drizzle
				condition = "Rain"
			} else if conditionCode < 600 {
				// Rain
				condition = "Rain"
			} else if conditionCode < 700 {
				// Snow
				condition = "Snow"
			} else if conditionCode < 800 {
				// Athmosphere
				if openWeatherMapAPIResponse.Weather[0].Main == "Mist" ||
					openWeatherMapAPIResponse.Weather[0].Main == "Fog" {
					condition = "Rain"
				} else {
					condition = "Windy"
				}
			} else if conditionCode == 800 {
				// Clear
				if openWeatherMapAPIResponse.DT < openWeatherMapAPIResponse.Sys.Sunset {
					condition = "Sunny"
				} else {
					condition = "Stars"
				}
			} else if conditionCode < 900 {
				// Cloud
				condition = "Cloudy"
			} else {
				condition = openWeatherMapAPIResponse.Weather[0].Main
			}

			is_forecast = "false"
			t := time.Unix(int64(openWeatherMapAPIResponse.DT), 0)
			local_datetime = t.Format(time.RFC850)
			logger.Println(local_datetime)
			speakable_location_string = openWeatherMapAPIResponse.Name
			temperature = fmt.Sprintf("%f", math.Round(openWeatherMapAPIResponse.Main.Temp))
			if weatherAPIUnit == "C" {
				temperature_unit = "C"
			} else {
				temperature_unit = "F"
			}
		}
	} else {
		condition = "Snow"
		is_forecast = "false"
		local_datetime = "test"              // preferably local time in UTC ISO 8601 format ("2022-06-15 12:21:22.123")
		speakable_location_string = location // preferably the processed location
		temperature = "120"
		temperature_unit = "C"
	}
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}

func weatherParser(speechText string, botLocation string, botUnits string) (string, string, string, string, string, string) {
	var specificLocation bool
	var apiLocation string
	var speechLocation string
	var hoursFromNow int
	if strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_IN)) {
		splitPhrase := strings.SplitAfter(speechText, lcztn.GetText(lcztn.STR_WEATHER_IN))
		speechLocation = strings.TrimSpace(splitPhrase[1])
		if len(splitPhrase) == 3 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2])
		} else if len(splitPhrase) == 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		} else if len(splitPhrase) > 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		}
		splitLocation := strings.Split(speechLocation, " ")
		if len(splitLocation) == 2 {
			speechLocation = splitLocation[0] + ", " + splitLocation[1]
		} else if len(splitLocation) == 3 {
			speechLocation = splitLocation[0] + " " + splitLocation[1] + ", " + splitLocation[2]
		}
		logger.Println("Location parsed from speech: " + "`" + speechLocation + "`")
		specificLocation = true
	} else {
		logger.Println("No location parsed from speech")
		specificLocation = false
	}
	hoursFromNow = 0
	hours, _, _ := time.Now().Clock()
	if strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_THIS_AFTERNOON)) {
		if hours < 14 {
			hoursFromNow = 14 - hours
		}
	} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_TONIGHT)) {
		if hours < 20 {
			hoursFromNow = 20 - hours
		}
	} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_THE_DAY_AFTER_TOMORROW)) {
		hoursFromNow = 24 - hours + 24 + 9
	} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_FORECAST)) ||
		strings.Contains(speechText, lcztn.GetText(lcztn.STR_WEATHER_TOMORROW)) {
		hoursFromNow = 24 - hours + 9
	}
	logger.Println("Looking for forecast " + strconv.Itoa(hoursFromNow) + " hours from now...")

	if specificLocation {
		apiLocation = speechLocation
	} else {
		apiLocation = botLocation
	}
	// call to weather API
	condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := getWeather(apiLocation, botUnits, hoursFromNow)
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}
