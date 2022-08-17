package wirepod

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

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

func getWeather(location string, botUnits string) (string, string, string, string, string, string) {
	var weatherEnabled bool
	var condition string
	var is_forecast string
	var local_datetime string
	var speakable_location_string string
	var temperature string
	var temperature_unit string
	weatherAPIEnabled := os.Getenv("WEATHERAPI_ENABLED")
	weatherAPIKey := os.Getenv("WEATHERAPI_KEY")
	weatherAPIUnit := os.Getenv("WEATHERAPI_UNIT")
	if weatherAPIEnabled == "true" && weatherAPIKey != "" {
		weatherEnabled = true
		logger("Weather API Enabled")
	} else {
		weatherEnabled = false
		logger("Weather API not enabled, using placeholder")
		if weatherAPIEnabled == "true" && weatherAPIKey == "" {
			logger("Weather API enabled, but Weather API key not set")
		}
	}
	if weatherEnabled {
		if botUnits != "" {
			if botUnits == "F" {
				logger("Weather units set to F")
				weatherAPIUnit = "F"
			} else if botUnits == "C" {
				logger("Weather units set to C")
				weatherAPIUnit = "C"
			}
		} else if weatherAPIUnit != "F" && weatherAPIUnit != "C" {
			logger("Weather API unit not set, using F")
			weatherAPIUnit = "F"
		}
	}

	if weatherEnabled {
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
		jsonFile, _ := os.ReadFile("./weather-map.json")
		json.Unmarshal(jsonFile, &weatherAPICladMap)
		var weatherStruct weatherAPIResponseStruct
		json.Unmarshal([]byte(weatherResponse), &weatherStruct)
		var matchedValue bool
		for _, b := range weatherAPICladMap {
			if b.APIValue == weatherStruct.Current.Condition.Text {
				condition = b.CladType
				logger("API Value: " + b.APIValue + ", Clad Type: " + b.CladType)
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
	if strings.Contains(speechText, " in ") {
		splitPhrase := strings.SplitAfter(speechText, " in ")
		speechLocation = strings.TrimSpace(splitPhrase[1])
		if len(splitPhrase) == 3 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2])
		} else if len(splitPhrase) == 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		} else if len(splitPhrase) > 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		}
		logger("Location parsed from speech: " + "`" + speechLocation + "`")
		specificLocation = true
	} else {
		logger("No location parsed from speech")
		specificLocation = false
	}
	if specificLocation {
		apiLocation = speechLocation
	} else {
		apiLocation = botLocation
	}
	// call to weather API
	condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := getWeather(apiLocation, botUnits)
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}
