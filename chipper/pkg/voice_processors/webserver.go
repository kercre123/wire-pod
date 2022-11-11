package wirepod

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
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
		if _, err := os.Stat("./customIntents.json"); err == nil {
			logger("Found customIntents.json")
			var customIntentJSON intentsStruct
			customIntentJSONFile, _ := os.ReadFile("./customIntents.json")
			json.Unmarshal(customIntentJSONFile, &customIntentJSON)
			logger("Number of custom intents (current): " + strconv.Itoa(len(customIntentJSON)))
			customIntentJSON = append(customIntentJSON, struct {
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Utterances  []string `json:"utterances"`
				Intent      string   `json:"intent"`
				Params      struct {
					ParamName  string `json:"paramname"`
					ParamValue string `json:"paramvalue"`
				} `json:"params"`
				Exec     string   `json:"exec"`
				ExecArgs []string `json:"execargs"`
			}{Name: name, Description: description, Utterances: strings.Split(utterances, ","), Intent: intent, Params: struct {
				ParamName  string `json:"paramname"`
				ParamValue string `json:"paramvalue"`
			}{ParamName: paramName, ParamValue: paramValue}, Exec: exec, ExecArgs: strings.Split(execArgs, ",")})
			customIntentJSONFile, _ = json.Marshal(customIntentJSON)
			os.WriteFile("./customIntents.json", customIntentJSONFile, 0644)
		} else {
			logger("Creating customIntents.json")
			customIntentJSONFile, _ := json.Marshal([]struct {
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Utterances  []string `json:"utterances"`
				Intent      string   `json:"intent"`
				Params      struct {
					ParamName  string `json:"paramname"`
					ParamValue string `json:"paramvalue"`
				} `json:"params"`
				Exec     string   `json:"exec"`
				ExecArgs []string `json:"execargs"`
			}{{Name: name, Description: description, Utterances: strings.Split(utterances, ","), Intent: intent, Params: struct {
				ParamName  string `json:"paramname"`
				ParamValue string `json:"paramvalue"`
			}{ParamName: paramName, ParamValue: paramValue}, Exec: exec, ExecArgs: strings.Split(execArgs, ",")}})
			os.WriteFile("./customIntents.json", customIntentJSONFile, 0644)
		}
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
		if _, err := os.Stat("./customIntents.json"); err == nil {
			// do nothing
		} else {
			fmt.Fprintf(w, "err: you must create an intent first")
			return
		}
		var customIntentJSON intentsStruct
		customIntentJSONFile, err := os.ReadFile("./customIntents.json")
		if err != nil {
			logger(err)
		}
		json.Unmarshal(customIntentJSONFile, &customIntentJSON)
		newNumbera, _ := strconv.Atoi(number)
		newNumber := newNumbera - 1
		if newNumber > len(customIntentJSON) {
			fmt.Fprintf(w, "err: there are only "+strconv.Itoa(len(customIntentJSON))+" intents")
			return
		}
		if name != "" {
			customIntentJSON[newNumber].Name = name
		}
		if description != "" {
			customIntentJSON[newNumber].Description = description
		}
		if utterances != "" {
			customIntentJSON[newNumber].Utterances = strings.Split(utterances, ",")
		}
		if intent != "" {
			customIntentJSON[newNumber].Intent = intent
		}
		if paramName != "" {
			customIntentJSON[newNumber].Params.ParamName = paramName
		}
		if paramValue != "" {
			customIntentJSON[newNumber].Params.ParamValue = paramValue
		}
		if exec != "" {
			customIntentJSON[newNumber].Exec = exec
		}
		if execArgs != "" {
			customIntentJSON[newNumber].ExecArgs = strings.Split(execArgs, ",")
		}
		newCustomIntentJSONFile, _ := json.Marshal(customIntentJSON)
		os.WriteFile("./customIntents.json", newCustomIntentJSONFile, 0644)
		fmt.Fprintf(w, "intent edited successfully")
		return
	case r.URL.Path == "/api/get_custom_intents_json":
		if _, err := os.Stat("./customIntents.json"); err == nil {
			// do nothing
		} else {
			fmt.Fprintf(w, "err: you must create an intent first")
			return
		}
		customIntentJSONFile, err := os.ReadFile("./customIntents.json")
		if err != nil {
			logger(err)
		}
		fmt.Fprint(w, string(customIntentJSONFile))
		return
	case r.URL.Path == "/api/remove_custom_intent":
		number := r.FormValue("number")
		if number == "" {
			fmt.Fprintf(w, "err: a number is required")
			return
		}
		if _, err := os.Stat("./customIntents.json"); err == nil {
			// do nothing
		} else {
			fmt.Fprintf(w, "err: you must create an intent first")
			return
		}
		var customIntentJSON intentsStruct
		customIntentJSONFile, err := os.ReadFile("./customIntents.json")
		if err != nil {
			logger(err)
		}
		json.Unmarshal(customIntentJSONFile, &customIntentJSON)
		newNumbera, _ := strconv.Atoi(number)
		newNumber := newNumbera - 1
		if newNumber > len(customIntentJSON) {
			fmt.Fprintf(w, "err: there are only "+strconv.Itoa(len(customIntentJSON))+" intents")
			return
		}
		customIntentJSON = append(customIntentJSON[:newNumber], customIntentJSON[newNumber+1:]...)
		newCustomIntentJSONFile, _ := json.Marshal(customIntentJSON)
		os.WriteFile("./customIntents.json", newCustomIntentJSONFile, 0644)
		fmt.Fprintf(w, "intent removed successfully")
		return
	case r.URL.Path == "/api/add_bot":
		botESN := r.FormValue("esn")
		botLocation := r.FormValue("location")
		botUnits := r.FormValue("units")
		botFirmwarePrefix := r.FormValue("firmwareprefix")
		var is_early_opus bool
		var use_play_specific bool
		if botESN == "" || botLocation == "" || botUnits == "" || botFirmwarePrefix == "" {
			fmt.Fprintf(w, "err: all fields are required")
			return
		}
		firmwareSplit := strings.Split(botFirmwarePrefix, ".")
		if len(firmwareSplit) != 2 {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		if botUnits != "F" && botUnits != "C" {
			fmt.Fprintf(w, "err: units must be either F or C")
			return
		}
		firmware1, _ := strconv.Atoi(firmwareSplit[0])
		firmware2, err := strconv.Atoi(firmwareSplit[1])
		if err != nil {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		if firmware1 >= 1 && firmware2 < 6 {
			is_early_opus = false
			use_play_specific = true
		} else if firmware1 >= 1 && firmware2 >= 6 {
			is_early_opus = false
			use_play_specific = false
		} else if firmware1 == 0 {
			is_early_opus = true
			use_play_specific = true
		} else {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		type botConfigStruct []struct {
			Esn             string `json:"esn"`
			Location        string `json:"location"`
			Units           string `json:"units"`
			UsePlaySpecific bool   `json:"use_play_specific"`
			IsEarlyOpus     bool   `json:"is_early_opus"`
		}
		var botConfig botConfigStruct
		if _, err := os.Stat("./botConfig.json"); err == nil {
			// read botConfig.json and append to it with the form information
			botConfigFile, err := os.ReadFile("./botConfig.json")
			if err != nil {
				logger(err)
			}
			json.Unmarshal(botConfigFile, &botConfig)
			botConfig = append(botConfig, struct {
				Esn             string `json:"esn"`
				Location        string `json:"location"`
				Units           string `json:"units"`
				UsePlaySpecific bool   `json:"use_play_specific"`
				IsEarlyOpus     bool   `json:"is_early_opus"`
			}{Esn: botESN, Location: botLocation, Units: botUnits, UsePlaySpecific: use_play_specific, IsEarlyOpus: is_early_opus})
			newBotConfigJSONFile, _ := json.Marshal(botConfig)
			os.WriteFile("./botConfig.json", newBotConfigJSONFile, 0644)
		} else {
			botConfig = append(botConfig, struct {
				Esn             string `json:"esn"`
				Location        string `json:"location"`
				Units           string `json:"units"`
				UsePlaySpecific bool   `json:"use_play_specific"`
				IsEarlyOpus     bool   `json:"is_early_opus"`
			}{Esn: botESN, Location: botLocation, Units: botUnits, UsePlaySpecific: use_play_specific, IsEarlyOpus: is_early_opus})
			newBotConfigJSONFile, _ := json.Marshal(botConfig)
			os.WriteFile("./botConfig.json", newBotConfigJSONFile, 0644)
		}
		fmt.Fprintf(w, "bot added successfully")
		return
	case r.URL.Path == "/api/remove_bot":
		number := r.FormValue("number")
		if _, err := os.Stat("./botConfig.json"); err == nil {
			// do nothing
		} else {
			fmt.Fprintf(w, "err: you must create a bot first")
			return
		}
		type botConfigStruct []struct {
			Esn             string `json:"esn"`
			Location        string `json:"location"`
			Units           string `json:"units"`
			UsePlaySpecific bool   `json:"use_play_specific"`
			IsEarlyOpus     bool   `json:"is_early_opus"`
		}
		var botConfigJSON botConfigStruct
		botConfigJSONFile, err := os.ReadFile("./botConfig.json")
		if err != nil {
			logger(err)
		}
		json.Unmarshal(botConfigJSONFile, &botConfigJSON)
		newNumbera, _ := strconv.Atoi(number)
		newNumber := newNumbera - 1
		if newNumber > len(botConfigJSON) {
			fmt.Fprintf(w, "err: there are only "+strconv.Itoa(len(botConfigJSON))+" bots")
			return
		}
		logger(botConfigJSON[newNumber].Esn + " bot is being removed")
		botConfigJSON = append(botConfigJSON[:newNumber], botConfigJSON[newNumber+1:]...)
		newBotConfigJSONFile, _ := json.Marshal(botConfigJSON)
		os.WriteFile("./botConfig.json", newBotConfigJSONFile, 0644)
		fmt.Fprintf(w, "bot removed successfully")
		return
	case r.URL.Path == "/api/edit_bot":
		number := r.FormValue("number")
		botESN := r.FormValue("esn")
		botLocation := r.FormValue("location")
		botUnits := r.FormValue("units")
		botFirmwarePrefix := r.FormValue("firmwareprefix")
		if botESN == "" || botLocation == "" || botUnits == "" || botFirmwarePrefix == "" {
			fmt.Fprintf(w, "err: all fields are required")
			return
		}
		firmwareSplit := strings.Split(botFirmwarePrefix, ".")
		if len(firmwareSplit) != 2 {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		if botUnits != "F" && botUnits != "C" {
			fmt.Fprintf(w, "err: units must be either F or C")
			return
		}
		var is_early_opus bool
		var use_play_specific bool
		firmware1, _ := strconv.Atoi(firmwareSplit[0])
		firmware2, err := strconv.Atoi(firmwareSplit[1])
		if err != nil {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		if firmware1 >= 1 && firmware2 < 6 {
			is_early_opus = false
			use_play_specific = true
		} else if firmware1 >= 1 && firmware2 >= 6 {
			is_early_opus = false
			use_play_specific = false
		} else if firmware1 == 0 {
			is_early_opus = true
			use_play_specific = true
		} else {
			fmt.Fprintf(w, "err: firmware prefix must be in the format: 1.5")
			return
		}
		type botConfigStruct []struct {
			Esn             string `json:"esn"`
			Location        string `json:"location"`
			Units           string `json:"units"`
			UsePlaySpecific bool   `json:"use_play_specific"`
			IsEarlyOpus     bool   `json:"is_early_opus"`
		}
		var botConfig botConfigStruct
		if _, err := os.Stat("./botConfig.json"); err == nil {
			// read botConfig.json and append to it with the form information
			botConfigFile, err := os.ReadFile("./botConfig.json")
			if err != nil {
				logger(err)
			}
			json.Unmarshal(botConfigFile, &botConfig)
			newNumbera, _ := strconv.Atoi(number)
			newNumber := newNumbera - 1
			botConfig[newNumber].Esn = botESN
			botConfig[newNumber].Location = botLocation
			botConfig[newNumber].Units = botUnits
			botConfig[newNumber].UsePlaySpecific = use_play_specific
			botConfig[newNumber].IsEarlyOpus = is_early_opus
			newBotConfigJSONFile, _ := json.Marshal(botConfig)
			os.WriteFile("./botConfig.json", newBotConfigJSONFile, 0644)
		} else {
			fmt.Fprintln(w, "err: you must create a bot first")
			return
		}
		fmt.Fprintf(w, "bot edited successfully")
		return
	case r.URL.Path == "/api/get_bot_json":
		if _, err := os.Stat("./botConfig.json"); err == nil {
			// do nothing
		} else {
			fmt.Fprintf(w, "err: you must add a bot first")
			return
		}
		botConfigJSONFile, err := os.ReadFile("./botConfig.json")
		if err != nil {
			logger(err)
		}
		fmt.Fprint(w, string(botConfigJSONFile))
		return
	}
}

func StartWebServer() {
	var webPort string
	http.HandleFunc("/api/", apiHandler)
	fileServer := http.FileServer(http.Dir("./webroot"))
	http.Handle("/", fileServer)
	if os.Getenv("WEBSERVER_PORT") != "" {
		if _, err := strconv.Atoi(os.Getenv("WEBSERVER_PORT")); err == nil {
			webPort = os.Getenv("WEBSERVER_PORT")
		} else {
			logger("WEBSERVER_PORT contains letters, using default of 8080")
			webPort = "8080"
		}
	} else {
		webPort = "8080"
	}
	fmt.Printf("Starting webserver at port " + webPort + " (http://localhost:" + webPort + ")\n")
	if err := http.ListenAndServe(":"+webPort, nil); err != nil {
		log.Fatal(err)
	}
}
