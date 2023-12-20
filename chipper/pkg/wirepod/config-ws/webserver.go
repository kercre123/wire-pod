package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization"
	processreqs "github.com/kercre123/wire-pod/chipper/pkg/wirepod/preqs"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
)

var SttInitFunc func() error

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	case r.URL.Path == "/api/add_custom_intent":
		name := r.FormValue("name")
		description := r.FormValue("description")
		utterances := r.FormValue("utterances")
		intent := r.FormValue("intent")
		paramName := r.FormValue("paramname")
		paramValue := r.FormValue("paramvalue")
		exec := r.FormValue("exec")
		execArgs := r.FormValue("execargs")
		if name == "" || description == "" || utterances == "" || intent == "" {
			fmt.Fprintf(w, "missing required field (name, description, utterances, and intent are required)")
			return
		}
		vars.CustomIntentsExist = true
		vars.CustomIntents = append(vars.CustomIntents, struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Utterances  []string `json:"utterances"`
			Intent      string   `json:"intent"`
			Params      struct {
				ParamName  string `json:"paramname"`
				ParamValue string `json:"paramvalue"`
			} `json:"params"`
			Exec           string   `json:"exec"`
			ExecArgs       []string `json:"execargs"`
			IsSystemIntent bool     `json:"issystem"`
		}{Name: name, Description: description, Utterances: strings.Split(utterances, ","), Intent: intent, Params: struct {
			ParamName  string `json:"paramname"`
			ParamValue string `json:"paramvalue"`
		}{ParamName: paramName, ParamValue: paramValue}, Exec: exec, ExecArgs: strings.Split(execArgs, ","), IsSystemIntent: false})
		customIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
		os.WriteFile(vars.CustomIntentsPath, customIntentJSONFile, 0644)
		fmt.Fprintf(w, "intent added successfully")
		return
	case r.URL.Path == "/api/edit_custom_intent":
		number := r.FormValue("number")
		name := r.FormValue("name")
		description := r.FormValue("description")
		utterances := r.FormValue("utterances")
		intent := r.FormValue("intent")
		paramName := r.FormValue("paramname")
		paramValue := r.FormValue("paramvalue")
		exec := r.FormValue("exec")
		execArgs := r.FormValue("execargs")
		if number == "" {
			fmt.Fprintf(w, "err: a number is required")
			return
		}
		if name == "" && description == "" && utterances == "" && intent == "" && paramName == "" && paramValue == "" && exec == "" {
			fmt.Fprintf(w, "err: an entry must be edited")
			return
		}
		if !vars.CustomIntentsExist {
			fmt.Fprintf(w, "err: you must create an intent first")
			return
		}
		newNumbera, _ := strconv.Atoi(number)
		newNumber := newNumbera - 1
		if newNumber > len(vars.CustomIntents) {
			fmt.Fprintf(w, "err: there are only "+strconv.Itoa(len(vars.CustomIntents))+" intents")
			return
		}
		if name != "" {
			vars.CustomIntents[newNumber].Name = name
		}
		if description != "" {
			vars.CustomIntents[newNumber].Description = description
		}
		if utterances != "" {
			vars.CustomIntents[newNumber].Utterances = strings.Split(utterances, ",")
		}
		if intent != "" {
			vars.CustomIntents[newNumber].Intent = intent
		}
		if paramName != "" {
			vars.CustomIntents[newNumber].Params.ParamName = paramName
		}
		if paramValue != "" {
			vars.CustomIntents[newNumber].Params.ParamValue = paramValue
		}
		if exec != "" {
			vars.CustomIntents[newNumber].Exec = exec
		}
		if execArgs != "" {
			vars.CustomIntents[newNumber].ExecArgs = strings.Split(execArgs, ",")
		}
		vars.CustomIntents[newNumber].IsSystemIntent = false
		newCustomIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
		os.WriteFile(vars.CustomIntentsPath, newCustomIntentJSONFile, 0644)
		fmt.Fprintf(w, "intent edited successfully")
		return
	case r.URL.Path == "/api/get_custom_intents_json":
		if !vars.CustomIntentsExist {
			fmt.Fprintf(w, "error: you must create an intent first")
			return
		}
		customIntentJSONFile, err := os.ReadFile(vars.CustomIntentsPath)
		if err != nil {
			logger.Println(err)
		}
		fmt.Fprint(w, string(customIntentJSONFile))
		return
	case r.URL.Path == "/api/remove_custom_intent":
		number := r.FormValue("number")
		if number == "" {
			fmt.Fprintf(w, "error: a number is required")
			return
		}
		if _, err := os.Stat(vars.CustomIntentsPath); err != nil {
			fmt.Fprintf(w, "error: you must create an intent first")
			return
		}
		newNumbera, _ := strconv.Atoi(number)
		newNumber := newNumbera - 1
		if newNumber > len(vars.CustomIntents) {
			fmt.Fprintf(w, "err: there are only "+strconv.Itoa(len(vars.CustomIntents))+" intents")
			return
		}
		vars.CustomIntents = append(vars.CustomIntents[:newNumber], vars.CustomIntents[newNumber+1:]...)
		newCustomIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
		os.WriteFile("./customIntents.json", newCustomIntentJSONFile, 0644)
		fmt.Fprintf(w, "intent removed successfully")
		return
	case r.URL.Path == "/api/set_weather_api":
		weatherProvider := r.FormValue("provider")
		weatherAPIKey := r.FormValue("api_key")
		if weatherProvider == "" {
			vars.APIConfig.Weather.Enable = false
		} else {
			vars.APIConfig.Weather.Enable = true
			vars.APIConfig.Weather.Key = weatherAPIKey
			vars.APIConfig.Weather.Provider = weatherProvider
		}
		vars.WriteConfigToDisk()
		fmt.Fprintf(w, "Changes successfully applied.")
		return
	case r.URL.Path == "/api/get_weather_api":
		weatherEnabled := false
		weatherProvider := ""
		weatherAPIKey := ""
		if vars.APIConfig.Weather.Enable {
			weatherEnabled = true
			weatherProvider = vars.APIConfig.Weather.Provider
			weatherAPIKey = vars.APIConfig.Weather.Key
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"weatherEnabled\": %t,", weatherEnabled)
		fmt.Fprintf(w, "  \"weatherProvider\": \"%s\",", weatherProvider)
		fmt.Fprintf(w, "  \"weatherApiKey\": \"%s\"", weatherAPIKey)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/set_kg_api":
		kgProvider := r.FormValue("provider")
		kgAPIKey := r.FormValue("api_key")
		// for houndify
		kgAPIID := r.FormValue("api_id")
		kgIntent := r.FormValue("intent_graph")
		// for Together AI Service
		kgModel := r.FormValue("model")

		if kgProvider == "" {
			vars.APIConfig.Knowledge.Enable = false
		} else {
			vars.APIConfig.Knowledge.Enable = true
			vars.APIConfig.Knowledge.Provider = kgProvider
			vars.APIConfig.Knowledge.Key = kgAPIKey
			vars.APIConfig.Knowledge.Model = kgModel
			vars.APIConfig.Knowledge.ID = kgAPIID
		}
		if kgProvider == "openai" && kgIntent == "true" {
			vars.APIConfig.Knowledge.IntentGraph = true
			if r.FormValue("robot_name") == "" {
				vars.APIConfig.Knowledge.RobotName = "Vector"
			} else {
				vars.APIConfig.Knowledge.RobotName = r.FormValue("robot_name")
			}
		} else if kgProvider == "openai" && kgIntent == "false" {
			vars.APIConfig.Knowledge.IntentGraph = false
			vars.APIConfig.Knowledge.RobotName = ""
		}
		vars.WriteConfigToDisk()
		fmt.Fprintf(w, "Changes successfully applied.")
		return
	case r.URL.Path == "/api/get_kg_api":
		kgEnabled := false
		kgProvider := ""
		kgAPIKey := ""
		kgAPIID := ""
		kgModel := ""
		kgIntent := false
		kgRobotName := ""
		if vars.APIConfig.Knowledge.Enable {
			kgEnabled = true
			kgProvider = vars.APIConfig.Knowledge.Provider
			kgAPIKey = vars.APIConfig.Knowledge.Key
			kgModel = vars.APIConfig.Knowledge.Model
			kgAPIID = vars.APIConfig.Knowledge.ID
			kgIntent = vars.APIConfig.Knowledge.IntentGraph
			kgRobotName = vars.APIConfig.Knowledge.RobotName
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"kgEnabled\": %t,", kgEnabled)
		fmt.Fprintf(w, "  \"kgProvider\": \"%s\",", kgProvider)
		fmt.Fprintf(w, "  \"kgApiKey\": \"%s\",", kgAPIKey)
		fmt.Fprintf(w, "  \"kgModel\": \"%s\",", kgModel)
		fmt.Fprintf(w, "  \"kgApiID\": \"%s\",", kgAPIID)
		fmt.Fprintf(w, "  \"kgIntentGraph\": \"%t\",", kgIntent)
		fmt.Fprintf(w, "  \"kgRobotName\": \"%s\"", kgRobotName)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/set_stt_info":
		language := r.FormValue("language")
		if vars.APIConfig.STT.Service == "vosk" {
			// check if language is valid
			matched := false
			for _, lang := range localization.ValidVoskModels {
				if lang == language {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Fprint(w, "error: language not valid")
				return
			}
			// check if language is downloaded already
			matched = false
			for _, lang := range vars.DownloadedVoskModels {
				if lang == language {
					matched = true
					break
				}
			}
			if !matched {
				go localization.DownloadVoskModel(language)
				fmt.Fprint(w, "downloading language model")
			} else {
				vars.APIConfig.STT.Language = language
				vars.APIConfig.PastInitialSetup = true
				vars.WriteConfigToDisk()
				processreqs.ReloadVosk()
				logger.Println("Reloaded voice processor successfully")
				fmt.Fprint(w, "language switched successfully")
			}
		} else if vars.APIConfig.STT.Service == "whisper.cpp" {
			matched := false
			for _, lang := range localization.ValidVoskModels {
				if lang == language {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Fprint(w, "error: language not valid")
				return
			}

			vars.APIConfig.STT.Language = language
			vars.APIConfig.PastInitialSetup = true
			vars.WriteConfigToDisk()
			processreqs.ReloadVosk()
			logger.Println("Reloaded voice processor successfully")
			fmt.Fprint(w, "language switched successfully")
		} else {
			fmt.Fprint(w, "error: service must be vosk or whisper")
		}
		return
	case r.URL.Path == "/api/get_download_status":
		fmt.Fprint(w, localization.DownloadStatus)
		if localization.DownloadStatus == "success" || strings.Contains(localization.DownloadStatus, "error") {
			localization.DownloadStatus = "not downloading"
		}
		return
	case r.URL.Path == "/api/get_stt_info":
		sttLanguage := vars.APIConfig.STT.Language
		sttProvider := vars.APIConfig.STT.Service
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"sttProvider\": \"%s\",", sttProvider)
		fmt.Fprintf(w, "  \"sttLanguage\": \"%s\"", sttLanguage)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/get_config":
		writeBytes, _ := json.Marshal(vars.APIConfig)
		w.Write(writeBytes)
		return
	case r.URL.Path == "/api/get_logs":
		fmt.Fprintf(w, logger.LogList)
		return
	case r.URL.Path == "/api/is_running":
		fmt.Fprintf(w, "true")
		return
	case r.URL.Path == "/api/generate_certs":
		err := botsetup.CreateCertCombo()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
		}
		fmt.Fprint(w, "done")
		return
	}
}

func certHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "/session-certs/"):
		split := strings.Split(r.URL.Path, "/")
		if len(split) < 3 {
			fmt.Fprint(w, "error: must request a cert by esn (ex. /session-certs/00e20145)")
			return
		}
		esn := split[2]
		fileBytes, err := os.ReadFile(vars.SessionCertPath + esn)
		if err != nil {
			fmt.Fprint(w, "error: cert does not exist")
			return
		}
		w.Write(fileBytes)
		return
	}
}

func StartWebServer() {
	botsetup.RegisterSSHAPI()
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/session-certs/", certHandler)
	var webRoot http.Handler
	if runtime.GOOS == "darwin" && vars.Packaged {
		appPath, _ := os.Executable()
		webRoot = http.FileServer(http.Dir(filepath.Dir(appPath) + "/../Frameworks/chipper/webroot"))
	} else if runtime.GOOS == "android" {
		webRoot = http.FileServer(http.Dir(vars.AndroidPath + "/static/webroot"))
	} else {
		webRoot = http.FileServer(http.Dir("./webroot"))
	}
	http.Handle("/", webRoot)
	fmt.Printf("Starting webserver at port " + vars.WebPort + " (http://localhost:" + vars.WebPort + ")\n")
	if err := http.ListenAndServe(":"+vars.WebPort, nil); err != nil {
		logger.Println("Error binding to " + vars.WebPort + ": " + err.Error())
		if vars.Packaged {
			logger.ErrMsg("FATAL: Wire-pod was unable to bind to port " + vars.WebPort + ". Another process is likely using it. Exiting.")
		}
		os.Exit(1)
	}
}
