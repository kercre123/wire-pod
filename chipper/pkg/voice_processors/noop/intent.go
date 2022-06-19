package noop

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"

	opus "github.com/digital-dream-labs/opus-go/opus"
)

var debugLogging bool

var weatherEnabled bool
var weatherAPIKey string

var intent pb.IntentResponse
var matched int = 0
var successMatched int = 0
var intentNum int = 0
var botNum int = 0

var intentParam string
var intentParamValue string
var newIntent string
var isParam bool
var intentParams map[string]string

var specificLocation bool
var apiLocation string
var speechLocation string
var condition string
var is_forecast string
var local_datetime string
var speakable_location_string string
var temperature string
var temperature_unit string

var username string
var nameSplitter string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
	TODO:
	1. Implement jdocs. These are files which are stored on the bot which contain the bot's
	default location, unit settings, etc. Helpful for weather.
		- workaround, ask the user for settings during setup.sh
	3. Overall take shell out of the picture (https://github.com/asticode/go-asticoqui)
	4. Maybe find a way to detect silence in the audio for better end handling.
		- probably unnecessary
*/

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

func paramChecker(req *vtt.IntentRequest, intent string, speechText string) {
	if strings.Contains(intent, "intent_photo_take_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, "me") || strings.Contains(speechText, "self") {
			intentParam = "entity_photo_selfie"
			intentParamValue = "photo_selfie"
		} else {
			intentParam = "entity_photo_selfie"
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_imperative_eyecolor") {
		isParam = true
		newIntent = "intent_imperative_eyecolor_specific_extend"
		intentParam = "eye_color"
		if strings.Contains(speechText, "purple") {
			intentParamValue = "COLOR_PURPLE"
		} else if strings.Contains(speechText, "blue") || strings.Contains(speechText, "sapphire") {
			intentParamValue = "COLOR_BLUE"
		} else if strings.Contains(speechText, "yellow") {
			intentParamValue = "COLOR_YELLOW"
		} else if strings.Contains(speechText, "teal") || strings.Contains(speechText, "tell") {
			intentParamValue = "COLOR_TEAL"
		} else if strings.Contains(speechText, "green") {
			intentParamValue = "COLOR_GREEN"
		} else if strings.Contains(speechText, "orange") {
			intentParamValue = "COLOR_ORANGE"
		} else {
			newIntent = intent
			intentParamValue = ""
			isParam = false
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_weather_extend") {
		isParam = true
		newIntent = intent
		condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := weatherParser(speechText)
		intentParams = map[string]string{"condition": condition, "is_forecast": is_forecast, "local_datetime": local_datetime, "speakable_location_string": speakable_location_string, "temperature": temperature, "temperature_unit": temperature_unit}
	} else if strings.Contains(intent, "intent_imperative_volumelevel_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, "medium lo") || strings.Contains(speechText, "media lo") || strings.Contains(speechText, "medium bo") || strings.Contains(speechText, "media bo") {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_2"
		} else if strings.Contains(speechText, "low") || strings.Contains(speechText, "quiet") {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else if strings.Contains(speechText, "medium high") || strings.Contains(speechText, "media high") || strings.Contains(speechText, "medium hide") || strings.Contains(speechText, "media hide") {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_4"
		} else if strings.Contains(speechText, "medium") || strings.Contains(speechText, "normal") || strings.Contains(speechText, "regular") || strings.Contains(speechText, "media") {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_3"
		} else if strings.Contains(speechText, "high") || strings.Contains(speechText, "loud") || strings.Contains(speechText, "hide") {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_5"
		} else if strings.Contains(speechText, "mute") || strings.Contains(speechText, "nothing") || strings.Contains(speechText, "silent") || strings.Contains(speechText, "off") || strings.Contains(speechText, "zero") || strings.Contains(speechText, "meet") {
			// there is no VOLUME_0 :(
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_names_username_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, "is") {
			nameSplitter = "is"
		} else if strings.Contains(speechText, "'s") {
			nameSplitter = "'s"
		} else if strings.Contains(speechText, "names") {
			nameSplitter = "names"
		}
		if strings.Contains(speechText, "is") || strings.Contains(speechText, "'s") || strings.Contains(speechText, "names") {
			splitPhrase := strings.SplitAfter(speechText, nameSplitter)
			username = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				username = username + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			if debugLogging == true {
				log.Println("Name parsed from speech: " + "`" + username + "`")
			}
			intentParam = "username"
			intentParamValue = username
			intentParams = map[string]string{intentParam: intentParamValue}
		} else {
			if debugLogging == true {
				log.Println("No name parsed from speech")
				intentParam = "username"
				intentParamValue = ""
				intentParams = map[string]string{intentParam: intentParamValue}
			}
		}
	} else if strings.Contains(intent, "intent_clock_settimer_extend") {
		isParam = true
		newIntent = intent
		timerSecs := words2num(speechText)
		if debugLogging == true {
			log.Println("Seconds parsed from speech: " + timerSecs)
		}
		intentParam = "timer_duration"
		intentParamValue = timerSecs
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_global_stop_extend") {
		isParam = true
		newIntent = intent
		intentParam = "what_to_stop"
		intentParamValue = "timer"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else {
		newIntent = intent
		intentParam = ""
		intentParamValue = ""
		isParam = false
		intentParams = map[string]string{intentParam: intentParamValue}
	}
	IntentPass(req, newIntent, speechText, intentParams, isParam)
}

func IntentPass(req *vtt.IntentRequest, intentThing string, speechText string, intentParams map[string]string, isParam bool) (*vtt.IntentResponse, error) {
	intent = pb.IntentResponse{
		IsFinal: true,
		IntentResult: &pb.IntentResult{
			QueryText:  speechText,
			Action:     intentThing,
			Parameters: intentParams,
		},
	}
	if err := req.Stream.Send(&intent); err != nil {
		return nil, err
	}
	r := &vtt.IntentResponse{
		Intent: &intent,
	}
	if debugLogging == true {
		log.Println("Intent Sent: " + intentThing)
		if isParam == true {
			log.Println("Parameters Sent:", intentParams)
		} else {
			log.Println("No Parameters Sent")
		}
	}
	return r, nil
}

func processTextAll(req *vtt.IntentRequest, voiceText string, listOfLists [][]string, intentList []string) {
	intentNum = 0
	matched = 0
	successMatched = 0
	for _, b := range listOfLists {
		for _, c := range b {
			if strings.Contains(voiceText, c) {
				paramChecker(req, intentList[intentNum], voiceText)
				successMatched = 1
				matched = 1
				break
			}
		}
		if matched == 1 {
			matched = 0
			break
		}
		intentNum = intentNum + 1
	}
	return
}

func pcmToWav(pcmFile string, wavFile string) {
	in, err := os.Open(pcmFile)
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(wavFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	e := wav.NewEncoder(out, 16000, 16, 1, 1)
	audioBuf, err := newAudioIntBuffer(in)
	if err != nil {
		log.Fatal(err)
	}
	if err := e.Write(audioBuf); err != nil {
		log.Fatal(err)
	}
	if err := e.Close(); err != nil {
		log.Fatal(err)
	}
}

func newAudioIntBuffer(r io.Reader) (*audio.IntBuffer, error) {
	buf := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  16000,
		},
	}
	for {
		var sample int16
		err := binary.Read(r, binary.LittleEndian, &sample)
		switch {
		case err == io.EOF:
			return &buf, nil
		case err != nil:
			return nil, err
		}
		buf.Data = append(buf.Data, int(sample))
	}
}

func dumpTimer(voiceTimer int) {
	time.Sleep(time.Millisecond * 300)
	for voiceTimer < 7 {
		voiceTimer = voiceTimer + 1
		//log.Println("Voice Timer")
		//log.Println(voiceTimer)
		time.Sleep(time.Second * 1)
	}
	return
}

func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func bytesToInt(stream opus.OggStream, data []byte, numBot int, voiceTimer int, die bool) {
	if die == true {
		return
	}
	f, err := os.Create("/tmp/" + strconv.Itoa(numBot) + "voice.pcm")
	if err != nil {
		log.Println(err)
	}
	n, err := stream.Decode(data)
	f.Write(n)
	if voiceTimer == 1 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice1.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice1.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped1")
		}
	}
	if voiceTimer == 2 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice2.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice2.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped2")
		}
	}
	if voiceTimer == 3 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice3.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice3.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped3")
		}
	}
	if voiceTimer == 4 {
		if _, err := os.Stat("/tmp/" + strconv.Itoa(numBot) + "voice4.wav"); err == nil {
			//
		} else {
			pcmToWav("/tmp/"+strconv.Itoa(numBot)+"voice.pcm", "/tmp/"+strconv.Itoa(numBot)+"voice4.wav")
			os.Create("/tmp/" + strconv.Itoa(numBot) + "dumped4")
		}
	}
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) ProcessIntent(req *vtt.IntentRequest) (*vtt.IntentResponse, error) {
	var voiceTimer int = 0
	var transcription1 string = ""
	var transcription2 string = ""
	var transcription3 string = ""
	var transcription4 string = ""
	var successMatch bool = false
	var transcribedText string
	var die bool = false
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		log.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}
	matched = 0
	botNum = botNum + 1
	if debugLogging == true {
		log.Println("Stream " + strconv.Itoa(botNum) + " opened.")
	}
	data := []byte{}
	data = append(data, req.FirstReq.InputAudio...)
	stream := opus.OggStream{}
	go func() {
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			voiceTimer = voiceTimer + 1
			time.Sleep(time.Second * 1)
		}
	}()
	go func() {
		var processOne bool = false
		var processTwo bool = false
		var processThree bool = false
		var processFour bool = false
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			if processOne == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNum) + "dumped1"); err == nil {
					processOne = true
					process1 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNum)+"voice1.wav")
					process1out, err := process1.Output()
					if err != nil {
						//
					}
					transcription1 = strings.TrimSpace(string(process1out))
					log.Println("1: " + transcription1)
				}
			}
			if processTwo == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNum) + "dumped2"); err == nil {
					processTwo = true
					process2 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNum)+"voice2.wav")
					process2out, err := process2.Output()
					if err != nil {
						//
					}
					transcription2 = strings.TrimSpace(string(process2out))
					log.Println("2: " + transcription2)
				}
			}
			if processThree == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNum) + "dumped3"); err == nil {
					processThree = true
					process3 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNum)+"voice3.wav")
					process3out, err := process3.Output()
					if err != nil {
						//
					}
					transcription3 = strings.TrimSpace(string(process3out))
					log.Println("3: " + transcription3)
				}
			}
			if processFour == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(botNum) + "dumped4"); err == nil {
					processFour = true
					process4 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(botNum)+"voice4.wav")
					process4out, err := process4.Output()
					if err != nil {
						//
					}
					transcription4 = strings.TrimSpace(string(process4out))
					log.Println("4: " + transcription4)
					successMatch = true
				}
			}
		}
	}()
	for {
		chunk, err := req.Stream.Recv()
		if err != nil {
			if err == io.EOF {
				IntentPass(req, "intent_system_noaudio", "EOF error", map[string]string{"error": "EOF"}, true)
				break
			}
		}
		if transcription2 != "" {
			if transcription1 == transcription2 {
				log.Println("Speech stopped, 2: " + transcription1)
				transcribedText = transcription1
				die = true
				break
			} else if transcription2 != "" {
				if transcription2 == transcription3 {
					log.Println("Speech stopped, 3: " + transcription2)
					transcribedText = transcription2
					die = true
					break
				} else if transcription3 != "" {
					if transcription3 == transcription4 {
						log.Println("Speech stopped, 4: " + transcription3)
						transcribedText = transcription3
						die = true
						break
					} else if transcription4 != "" {
						if transcription3 == transcription4 {
							log.Println("Speech stopped, 4: " + transcription4)
							transcribedText = transcription4
							die = true
							break
						} else {
							log.Println("Speech stopped, 4 (nm): " + transcription4)
							transcribedText = transcription4
							die = true
							break
						}
					}
				}
			}
		}
		if transcription4 == "" && successMatch == true {
			transcribedText = ""
			break
		}
		data = append(data, chunk.InputAudio...)
		go bytesToInt(stream, data, botNum, voiceTimer, die)
	}
	processTextAll(req, transcribedText, matchListList, intentsList)
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped4").Run()
	if successMatched == 0 {
		if debugLogging == true {
			log.Println("No intent was matched.")
		}
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false)
	}
	botNum = botNum - 1
	return nil, nil
}
