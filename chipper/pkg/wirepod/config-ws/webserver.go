package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
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
		// Patch source.sh
		lines, err := readLines("source.sh")
		if err == nil {
			var outlines []string
			for _, line := range lines {
				if strings.HasPrefix(line, "export WEATHERAPI_ENABLED") {
					if weatherProvider == "" {
						line = "export WEATHERAPI_ENABLED=false"
					} else {
						line = "export WEATHERAPI_ENABLED=true"
					}
				} else if strings.HasPrefix(line, "export WEATHERAPI_PROVIDER") {
					line = "export WEATHERAPI_PROVIDER=" + weatherProvider
				} else if strings.HasPrefix(line, "export WEATHERAPI_KEY") {
					line = "export WEATHERAPI_KEY=" + weatherAPIKey
				}
				outlines = append(outlines, line)
			}
			writeLines(outlines, "source.sh")
			fmt.Fprintf(w, "Changes saved. Restart needed.")
		}
		return
	case r.URL.Path == "/api/get_weather_api":
		weatherEnabled := 0
		weatherProvider := ""
		weatherAPIKey := ""
		lines, err := readLines("source.sh")
		if err == nil {
			for _, line := range lines {
				if strings.HasPrefix(line, "export WEATHERAPI_ENABLED=true") {
					weatherEnabled = 1
				} else if strings.HasPrefix(line, "export WEATHERAPI_PROVIDER=") {
					weatherProvider = strings.SplitAfter(line, "export WEATHERAPI_PROVIDER=")[1]
				} else if strings.HasPrefix(line, "export WEATHERAPI_KEY=") {
					weatherAPIKey = strings.SplitAfter(line, "export WEATHERAPI_KEY=")[1]
				}
			}
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"weatherEnabled\": %d,", weatherEnabled)
		fmt.Fprintf(w, "  \"weatherProvider\": \"%s\",", weatherProvider)
		fmt.Fprintf(w, "  \"weatherApiKey\": \"%s\"", weatherAPIKey)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/set_kg_api":
		kgProvider := r.FormValue("provider")
		kgAPIKey := r.FormValue("api_key")
		// for houndify
		kgAPIID := r.FormValue("api_id")
		// Patch source.sh
		lines, err := readLines("source.sh")
		var outlines []string
		if err == nil {
			for _, line := range lines {
				if strings.HasPrefix(line, "export KNOWLEDGE_ENABLED") {
					if kgProvider == "" {
						line = "export KNOWLEDGE_ENABLED=false"
					} else {
						line = "export KNOWLEDGE_ENABLED=true"
					}
				} else if strings.HasPrefix(line, "export KNOWLEDGE_PROVIDER") {
					line = "export KNOWLEDGE_PROVIDER=" + kgProvider
				} else if strings.HasPrefix(line, "export KNOWLEDGE_KEY") {
					line = "export KNOWLEDGE_KEY=" + kgAPIKey
				} else if strings.HasPrefix(line, "export KNOWLEDGE_ID") {
					line = "export KNOWLEDGE_ID=" + kgAPIID
				}
				outlines = append(outlines, line)
			}
			writeLines(outlines, "source.sh")
			fmt.Fprintf(w, "Changes saved. Restart needed.")
		}
		return
	case r.URL.Path == "/api/get_kg_api":
		kgEnabled := 0
		kgProvider := ""
		kgAPIKey := ""
		lines, err := readLines("source.sh")
		if err == nil {
			for _, line := range lines {
				if strings.HasPrefix(line, "export KNOWLEDGE_ENABLED=true") {
					kgEnabled = 1
				} else if strings.HasPrefix(line, "export KNOWLEDGE_PROVIDER=") {
					kgProvider = strings.SplitAfter(line, "export KNOWLEDGE_PROVIDER=")[1]
				} else if strings.HasPrefix(line, "export KNOWLEDGE_KEY=") {
					kgAPIKey = strings.SplitAfter(line, "export KNOWLEDGE_KEY=")[1]
				}
			}
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"kgEnabled\": %d,", kgEnabled)
		fmt.Fprintf(w, "  \"kgProvider\": \"%s\",", kgProvider)
		fmt.Fprintf(w, "  \"kgApiKey\": \"%s\"", kgAPIKey)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/set_stt_info":
		language := r.FormValue("language")

		// Patch source.sh
		lines, err := readLines("source.sh")
		var outlines []string
		if err == nil {
			for _, line := range lines {
				if strings.HasPrefix(line, "export STT_LANGUAGE") {
					line = "export STT_LANGUAGE=" + language
				}
				outlines = append(outlines, line)
			}
			writeLines(outlines, "source.sh")
			fmt.Fprintf(w, "Changes saved. Restart needed.")
		}
		return
	case r.URL.Path == "/api/get_stt_info":
		sttLanguage := ""
		sttProvider := ""
		lines, err := readLines("source.sh")
		if err == nil {
			for _, line := range lines {
				if strings.HasPrefix(line, "export STT_SERVICE=") {
					sttProvider = strings.SplitAfter(line, "export STT_SERVICE=")[1]
				} else if strings.HasPrefix(line, "export STT_LANGUAGE=") {
					sttLanguage = strings.SplitAfter(line, "export STT_LANGUAGE=")[1]
				}
			}
		}
		fmt.Fprintf(w, "{ ")
		fmt.Fprintf(w, "  \"sttProvider\": \"%s\",", sttProvider)
		fmt.Fprintf(w, "  \"sttLanguage\": \"%s\"", sttLanguage)
		fmt.Fprintf(w, "}")
		return
	case r.URL.Path == "/api/reset":
		// im not sure if this works... this is only temporary until a way to restart just the voice processor is added
		cmd := exec.Command("/bin/sh", "-c", "sudo systemctl restart wire-pod")
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(w, "%s", err.Error())
			log.Fatal(err)
		}
		return
	}
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
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
			logger.Println("WEBSERVER_PORT contains letters, using default of 8080")
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
