package wirepod

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func getWeather(location string) (string, string, string, string, string, string) {
	/*
		This is where you would make a call to a weather API to get the weather.
		You are given `location` which` is the location parsed from the speech
		which needs to be converted to something your API can understand.
		You have to return the following:
		condition = "Cloudy", "Sunny", "Cold", "Rain", "Thunderstorms", or "Windy"
		is_forecast = "true" or "false"
		local_datetime = "2022-06-15 12:21:22.123", UTC ISO 8601 date and time
		speakable_location_string = "New York"
		temperature = "83", degrees
		temperature_unit = "F" or "C"
	*/
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
		if debugLogging == true {
			log.Println("Weather API Enabled")
		}
	} else {
		weatherEnabled = false
		if debugLogging == true {
			log.Println("Weather API not enabled, using placeholder")
			if weatherAPIEnabled == "true" && weatherAPIKey == "" {
				log.Println("Weather API enabled, but Weather API key not set")
			}
		}
	}
	if weatherEnabled == true {
		if weatherAPIUnit != "F" && weatherAPIUnit != "C" {
			if debugLogging == true {
				log.Println("Weather API unit not set, using F")
			}
			weatherAPIUnit = "F"
		}
	}
	if weatherEnabled == true {
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
		body, _ := ioutil.ReadAll(resp.Body)
		weatherResponse := string(body)
		if debugLogging == true {
			log.Println(weatherResponse)
		}
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
		var weatherStruct weatherAPIResponseStruct
		json.Unmarshal([]byte(weatherResponse), &weatherStruct)
		condition = weatherStruct.Current.Condition.Text
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

func weatherParser(speechText string) (string, string, string, string, string, string) {
	var specificLocation bool
	var apiLocation string
	var speechLocation string
	if strings.Contains(speechText, " in ") {
		splitPhrase := strings.SplitAfter(speechText, "in")
		speechLocation = strings.TrimSpace(splitPhrase[1])
		if len(splitPhrase) == 3 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2])
		} else if len(splitPhrase) == 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		} else if len(splitPhrase) > 4 {
			speechLocation = speechLocation + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
		}
		if debugLogging == true {
			log.Println("Location parsed from speech: " + "`" + speechLocation + "`")
		}
		specificLocation = true
	} else {
		if debugLogging == true {
			log.Println("No location parsed from speech")
		}
		specificLocation = false
	}
	if specificLocation == true {
		apiLocation = speechLocation
	} else {
		// jdocs needs to be implemented
		apiLocation = "San Francisco"
	}
	// call to weather API
	condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := getWeather(apiLocation)
	return condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit
}
