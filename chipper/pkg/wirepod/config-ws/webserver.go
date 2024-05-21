package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
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
			vars.APIConfig.Weather.Key = strings.TrimSpace(weatherAPIKey)
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
			vars.APIConfig.Knowledge.Key = strings.TrimSpace(kgAPIKey)
			vars.APIConfig.Knowledge.Model = strings.TrimSpace(kgModel)
			vars.APIConfig.Knowledge.ID = strings.TrimSpace(kgAPIID)
		}
		if kgModel == "" && kgProvider == "together" {
			logger.Println("Together model wasn't provided, using default meta-llama/Llama-3-70b-chat-hf")
			vars.APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
		}
		if kgProvider == "openai" || kgProvider == "together" {
			if strings.TrimSpace(r.FormValue("openai_prompt")) != "" {
				vars.APIConfig.Knowledge.OpenAIPrompt = r.FormValue("openai_prompt")
			} else {
				vars.APIConfig.Knowledge.OpenAIPrompt = ""
			}
			if r.FormValue("save_chat") == "true" {
				vars.APIConfig.Knowledge.SaveChat = true
			} else {
				vars.APIConfig.Knowledge.SaveChat = false
			}
			if r.FormValue("commands_enable") == "true" {
				vars.APIConfig.Knowledge.CommandsEnable = true
			} else {
				vars.APIConfig.Knowledge.CommandsEnable = false
			}
		}
		if (kgProvider == "openai" || kgProvider == "together") && kgIntent == "true" {
			vars.APIConfig.Knowledge.IntentGraph = true
			if r.FormValue("robot_name") == "" {
				vars.APIConfig.Knowledge.RobotName = "Vector"
			} else {
				vars.APIConfig.Knowledge.RobotName = strings.TrimSpace(r.FormValue("robot_name"))
			}
		} else if (kgProvider == "openai" || kgProvider == "together") && kgIntent == "false" {
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
		kgOpenAIPrompt := ""
		kgSavePrompt := false
		kgCommandsEnable := false
		if vars.APIConfig.Knowledge.Enable {
			kgEnabled = true
			kgProvider = vars.APIConfig.Knowledge.Provider
			kgAPIKey = vars.APIConfig.Knowledge.Key
			kgModel = vars.APIConfig.Knowledge.Model
			kgAPIID = vars.APIConfig.Knowledge.ID
			kgIntent = vars.APIConfig.Knowledge.IntentGraph
			kgRobotName = vars.APIConfig.Knowledge.RobotName
			kgOpenAIPrompt = vars.APIConfig.Knowledge.OpenAIPrompt
			kgSavePrompt = vars.APIConfig.Knowledge.SaveChat
			kgCommandsEnable = vars.APIConfig.Knowledge.CommandsEnable
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"kgEnabled\": %t,", kgEnabled)
		fmt.Fprintf(w, "  \"kgProvider\": \"%s\",", kgProvider)
		fmt.Fprintf(w, "  \"kgApiKey\": \"%s\",", kgAPIKey)
		fmt.Fprintf(w, "  \"kgModel\": \"%s\",", kgModel)
		fmt.Fprintf(w, "  \"kgApiID\": \"%s\",", kgAPIID)
		fmt.Fprintf(w, "  \"kgIntentGraph\": \"%t\",", kgIntent)
		fmt.Fprintf(w, "  \"kgRobotName\": \"%s\",", kgRobotName)
		fmt.Fprintf(w, "  \"kgOpenAIPrompt\": \"%s\",", kgOpenAIPrompt)
		fmt.Fprintf(w, "  \"kgSaveChat\": \"%t\",", kgSavePrompt)
		fmt.Fprintf(w, "  \"kgCommandsEnable\": \"%t\"", kgCommandsEnable)
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
	case r.URL.Path == "/api/get_debug_logs":
		fmt.Fprintf(w, logger.LogTrayList)
		return
	case r.URL.Path == "/api/is_running":
		fmt.Fprintf(w, "true")
		return
	case r.URL.Path == "/api/delete_chats":
		os.Remove(vars.SavedChatsPath)
		vars.RememberedChats = []vars.RememberedChat{}
		fmt.Fprintf(w, "done")
		return
	case strings.Contains(r.URL.Path, "/api/get_ota"):
		otaName := strings.Split(r.URL.Path, "/")[3]
		//https://archive.org/download/vector-pod-firmware/vicos-2.0.1.6076ep.ota
		targetURL, err := url.Parse("https://archive.org/download/vector-pod-firmware/" + strings.TrimSpace(otaName))
		if err != nil {
			http.Error(w, "Failed to parse URL", http.StatusInternalServerError)
			return
		}
		req, err := http.NewRequest(r.Method, targetURL.String(), nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to perform request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		//w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, "failed to copy response body", http.StatusInternalServerError)
			return
		}
	case r.URL.Path == "/api/get_version_info":
		type VerInfo struct {
			Installed       string `json:"installed"`
			Current         string `json:"current"`
			UpdateAvailable bool   `json:"avail"`
		}
		var verInfo VerInfo
		ver, err := os.ReadFile(vars.VersionFile)
		if err != nil {
			fmt.Fprint(w, "error: version file doesn't exist")
			return
		}
		installedVer := strings.TrimSpace(string(ver))
		currentVer, err := GetLatestReleaseTag("kercre123", "WirePod")
		if err != nil {
			fmt.Fprint(w, "error comming with github: "+err.Error())
			return
		}
		verInfo.Installed = installedVer
		verInfo.Current = strings.TrimSpace(currentVer)
		if installedVer != strings.TrimSpace(currentVer) {
			verInfo.UpdateAvailable = true
		}
		marshalled, _ := json.Marshal(verInfo)
		w.Write(marshalled)
	case r.URL.Path == "/api/generate_certs":
		err := botsetup.CreateCertCombo()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
		}
		fmt.Fprint(w, "done")
		return
	}
}

func GetLatestReleaseTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type Release struct {
		TagName string `json:"tag_name"`
	}
	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
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
		fileBytes, err := os.ReadFile(path.Join(vars.SessionCertPath, esn))
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprint(w, "error: cert does not exist")
			return
		}
		w.Write(fileBytes)
		return
	}
}

func StartWebServer() {
	botsetup.RegisterSSHAPI()
	botsetup.RegisterBLEAPI()
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/session-certs/", certHandler)
	var webRoot http.Handler
	if runtime.GOOS == "darwin" && vars.Packaged {
		appPath, _ := os.Executable()
		webRoot = http.FileServer(http.Dir(filepath.Dir(appPath) + "/../Frameworks/chipper/webroot"))
	} else if runtime.GOOS == "android" || runtime.GOOS == "ios" {
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
