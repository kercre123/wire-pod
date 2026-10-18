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
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/productivity"
	"github.com/kercre123/wire-pod/chipper/pkg/scripting"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization"
	processreqs "github.com/kercre123/wire-pod/chipper/pkg/wirepod/preqs"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
)

var SttInitFunc func() error

var ProductivityImgPath = "./productivity-images"

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	switch strings.TrimPrefix(r.URL.Path, "/api/") {
	case "add_custom_intent":
		handleAddCustomIntent(w, r)
	case "edit_custom_intent":
		handleEditCustomIntent(w, r)
	case "get_custom_intents_json":
		handleGetCustomIntentsJSON(w)
	case "remove_custom_intent":
		handleRemoveCustomIntent(w, r)
	case "set_weather_api":
		handleSetWeatherAPI(w, r)
	case "get_weather_api":
		handleGetWeatherAPI(w)
	case "set_kg_api":
		handleSetKGAPI(w, r)
	case "get_kg_api":
		handleGetKGAPI(w)
	case "set_stt_info":
		handleSetSTTInfo(w, r)
	case "get_download_status":
		handleGetDownloadStatus(w)
	case "get_stt_info":
		handleGetSTTInfo(w)
	case "get_config":
		handleGetConfig(w)
	case "get_logs":
		handleGetLogs(w)
	case "get_debug_logs":
		handleGetDebugLogs(w)
	case "is_running":
		handleIsRunning(w)
	case "delete_chats":
		handleDeleteChats(w)
	case "get_ota":
		handleGetOTA(w, r)
	case "get_version_info":
		handleGetVersionInfo(w)
	case "generate_certs":
		handleGenerateCerts(w)
	case "set_productivity_api":
		handleSetProductivityAPI(w, r)
	case "get_productivity_api":
		handleGetProductivityAPI(w)
	case "test_productivity_reminder":
		handleTestProductivityReminder(w, r)
	case "is_api_v3":
		fmt.Fprintf(w, "it is!")
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func handleAddCustomIntent(w http.ResponseWriter, r *http.Request) {
	var intent vars.CustomIntent
	if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if anyEmpty(intent.Name, intent.Description, intent.Intent) || len(intent.Utterances) == 0 {
		http.Error(w, "missing required field (name, description, utterances, and intent are required)", http.StatusBadRequest)
		return
	}
	intent.LuaScript = strings.TrimSpace(intent.LuaScript)
	if intent.LuaScript != "" {
		if err := scripting.ValidateLuaScript(intent.LuaScript); err != nil {
			http.Error(w, "lua validation error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	vars.CustomIntentsExist = true
	vars.CustomIntents = append(vars.CustomIntents, intent)
	saveCustomIntents()
	fmt.Fprint(w, "Intent added successfully.")
}

func handleEditCustomIntent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Number int `json:"number"`
		vars.CustomIntent
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if request.Number < 1 || request.Number > len(vars.CustomIntents) {
		http.Error(w, "invalid intent number", http.StatusBadRequest)
		return
	}
	intent := &vars.CustomIntents[request.Number-1]
	if request.Name != "" {
		intent.Name = request.Name
	}
	if request.Description != "" {
		intent.Description = request.Description
	}
	if len(request.Utterances) != 0 {
		intent.Utterances = request.Utterances
	}
	if request.Intent != "" {
		intent.Intent = request.Intent
	}
	if request.Params.ParamName != "" {
		intent.Params.ParamName = request.Params.ParamName
	}
	if request.Params.ParamValue != "" {
		intent.Params.ParamValue = request.Params.ParamValue
	}
	if request.Exec != "" {
		intent.Exec = request.Exec
	}
	if request.LuaScript != "" {
		intent.LuaScript = request.LuaScript
		if err := scripting.ValidateLuaScript(intent.LuaScript); err != nil {
			http.Error(w, "lua validation error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	if len(request.ExecArgs) != 0 {
		intent.ExecArgs = request.ExecArgs
	}
	intent.IsSystemIntent = false
	saveCustomIntents()
	fmt.Fprint(w, "Intent edited successfully.")
}

func handleGetCustomIntentsJSON(w http.ResponseWriter) {
	if !vars.CustomIntentsExist {
		http.Error(w, "you must create an intent first", http.StatusBadRequest)
		return
	}
	customIntentJSONFile, err := os.ReadFile(vars.CustomIntentsPath)
	if err != nil {
		http.Error(w, "could not read custom intents file", http.StatusInternalServerError)
		logger.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(customIntentJSONFile)
}

func handleRemoveCustomIntent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Number int `json:"number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if request.Number < 1 || request.Number > len(vars.CustomIntents) {
		http.Error(w, "invalid intent number", http.StatusBadRequest)
		return
	}
	vars.CustomIntents = append(vars.CustomIntents[:request.Number-1], vars.CustomIntents[request.Number:]...)
	saveCustomIntents()
	fmt.Fprint(w, "Intent removed successfully.")
}

func handleSetWeatherAPI(w http.ResponseWriter, r *http.Request) {
	var config struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if config.Provider == "" {
		vars.APIConfig.Weather.Enable = false
	} else {
		vars.APIConfig.Weather.Enable = true
		vars.APIConfig.Weather.Key = strings.TrimSpace(config.Key)
		vars.APIConfig.Weather.Provider = config.Provider
	}
	vars.WriteConfigToDisk()
	fmt.Fprint(w, "Changes successfully applied.")
}

func handleGetWeatherAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Weather)
}

func handleSetProductivityAPI(w http.ResponseWriter, r *http.Request) {
	// 10MB limit for image uploads
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	provider := r.FormValue("provider")
	key := r.FormValue("key")
	urlVal := r.FormValue("url")
	username := r.FormValue("username")
	password := r.FormValue("password")
	targetRobot := r.FormValue("target_robot")
	manualConfig := r.FormValue("manual_config")

	vars.APIConfig.Productivity.Enable = (provider != "" && provider != "None")
	vars.APIConfig.Productivity.Provider = provider
	vars.APIConfig.Productivity.Key = strings.TrimSpace(key)
	vars.APIConfig.Productivity.Url = strings.TrimSpace(urlVal)
	vars.APIConfig.Productivity.Username = strings.TrimSpace(username)
	vars.APIConfig.Productivity.Password = strings.TrimSpace(password)
	vars.APIConfig.Productivity.TargetRobot = strings.TrimSpace(targetRobot)
	vars.APIConfig.Productivity.ManualConfig = manualConfig

	files := r.MultipartForm.File["files"]
	if len(files) > 0 {
		if _, err := os.Stat(ProductivityImgPath); os.IsNotExist(err) {
			os.MkdirAll(ProductivityImgPath, 0755)
		}

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				logger.Println("Error opening uploaded file:", err)
				continue
			}
			defer file.Close()

			filename := filepath.Base(fileHeader.Filename)
			dstPath := filepath.Join(ProductivityImgPath, filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				logger.Println("Error creating destination file:", err)
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				logger.Println("Error saving file:", err)
			}
		}
	}

	vars.WriteConfigToDisk()
	fmt.Fprint(w, "Productivity settings applied.")
}

func handleGetProductivityAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Productivity)
}

func handleTestProductivityReminder(w http.ResponseWriter, r *http.Request) {
	logger.Println("Received request for /api/test_productivity_reminder")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.Println("Error parsing test form data: " + err.Error())
		http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	targetRobot := r.FormValue("target_robot")
	if targetRobot == "" {
		http.Error(w, "Target robot is required", http.StatusBadRequest)
		return
	}

	configStr := r.FormValue("reminder_config")
	var reminder productivity.ManualReminder
	if err := json.Unmarshal([]byte(configStr), &reminder); err != nil {
		logger.Println("Error parsing reminder JSON: " + err.Error())
		http.Error(w, "Invalid reminder config", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) > 0 {
		logger.Println("Test Request contains image file(s)")
		if _, err := os.Stat(ProductivityImgPath); os.IsNotExist(err) {
			os.MkdirAll(ProductivityImgPath, 0755)
		}
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				logger.Println("Error opening test image: " + err.Error())
				continue
			}
			defer file.Close()
			filename := filepath.Base(fileHeader.Filename)
			dstPath := filepath.Join(ProductivityImgPath, filename)
			dst, err := os.Create(dstPath)
			if err != nil {
				logger.Println("Error creating test image file: " + err.Error())
				continue
			}
			defer dst.Close()
			io.Copy(dst, file)
			logger.Println("Saved test image: " + dstPath)
		}
	} else if reminder.Image != "" {
		logger.Println("Test Request uses existing image: " + reminder.Image)
	}

	task := productivity.Task{
		ID:                  reminder.ID,
		RobotESN:            targetRobot,
		Phrases:             reminder.Phrases,
		Image:               reminder.Image,
		Source:              "test",
		RequireConfirmation: reminder.RequireConfirmation,
		SnoozeMinutes:       reminder.SnoozeMinutes,
	}

	productivity.InjectTestTask(task)
	fmt.Fprint(w, "Test reminder queued.")
}

func handleSetKGAPI(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&vars.APIConfig.Knowledge); err != nil {
		fmt.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	vars.WriteConfigToDisk()
	fmt.Fprint(w, "Changes successfully applied.")
}

func handleGetKGAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Knowledge)
}

func handleSetSTTInfo(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if vars.APIConfig.STT.Service == "vosk" {
		if !isValidLanguage(request.Language, localization.ValidVoskModels) {
			http.Error(w, "language not valid", http.StatusBadRequest)
			return
		}
		if !isDownloadedLanguage(request.Language, vars.DownloadedVoskModels) {
			go localization.DownloadVoskModel(request.Language)
			fmt.Fprint(w, "downloading language model...")
			return
		}
	} else if vars.APIConfig.STT.Service == "whisper.cpp" {
		if !isValidLanguage(request.Language, localization.ValidVoskModels) {
			http.Error(w, "language not valid", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "service must be vosk or whisper", http.StatusBadRequest)
		return
	}
	vars.APIConfig.STT.Language = request.Language
	vars.APIConfig.PastInitialSetup = true
	vars.WriteConfigToDisk()
	processreqs.ReloadVosk()
	logger.Println("Reloaded voice processor successfully")
	fmt.Fprint(w, "Language switched successfully.")
}

func handleGetDownloadStatus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(localization.DownloadStatus))
	if localization.DownloadStatus == "success" || strings.Contains(localization.DownloadStatus, "error") {
		localization.DownloadStatus = "not downloading"
	}
}

func handleGetSTTInfo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.STT)
}

func handleGetConfig(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig)
}

func handleGetLogs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logger.LogList))
}

func handleGetDebugLogs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logger.LogTrayList))
}

func handleIsRunning(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("true"))
}

func handleDeleteChats(w http.ResponseWriter) {
	vars.RememberedChats = []vars.RememberedChat{}
	fmt.Fprint(w, "done")
}

func handleGetOTA(w http.ResponseWriter, r *http.Request) {
	otaName := strings.Split(r.URL.Path, "/")[3]
	targetURL, err := url.Parse("https://archive.org/download/vector-pod-firmware/" + strings.TrimSpace(otaName))
	if err != nil {
		http.Error(w, "failed to parse URL", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest(r.Method, targetURL.String(), nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
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
		http.Error(w, "failed to perform request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "failed to copy response body", http.StatusInternalServerError)
	}
}

func handleGetVersionInfo(w http.ResponseWriter) {
	var installedVer string
	ver, err := os.ReadFile(vars.VersionFile)
	if err == nil {
		installedVer = strings.TrimSpace(string(ver))
	}
	currentVer, err := GetLatestReleaseTag("kercre123", "WirePod")
	if err != nil {
		http.Error(w, "error communicating with github (ver): "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentCommit, err := GetLatestCommitSha()
	if err != nil {
		http.Error(w, "error communicating with github (commit): "+err.Error(), http.StatusInternalServerError)
		return
	}
	type VersionInfo struct {
		FromSource      bool   `json:"fromsource"`
		InstalledVer    string `json:"installedversion"`
		InstalledCommit string `json:"installedcommit"`
		CurrentVer      string `json:"currentversion"`
		CurrentCommit   string `json:"currentcommit"`
		UpdateAvailable bool   `json:"avail"`
	}
	var fromSource bool
	if installedVer == "" {
		fromSource = true
	}
	var uAvail bool
	if fromSource {
		uAvail = vars.CommitSHA != strings.TrimSpace(currentCommit)
	} else {
		uAvail = installedVer != strings.TrimSpace(currentVer)
	}
	verInfo := VersionInfo{
		FromSource:      fromSource,
		InstalledVer:    installedVer,
		InstalledCommit: vars.CommitSHA,
		CurrentVer:      strings.TrimSpace(currentVer),
		CurrentCommit:   strings.TrimSpace(currentCommit),
		UpdateAvailable: uAvail,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verInfo)
}

func handleGenerateCerts(w http.ResponseWriter) {
	if err := botsetup.CreateCertCombo(); err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "done")
}

func saveCustomIntents() {
	customIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
	os.WriteFile(vars.CustomIntentsPath, customIntentJSONFile, 0644)
}

func DisableCachingAndSniffing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func StartWebServer() {
	botsetup.RegisterSSHAPI()
	botsetup.RegisterBLEAPI()
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/session-certs/", certHandler)
	http.Handle("/api/productivity-images/", http.StripPrefix("/api/productivity-images/", http.FileServer(http.Dir(ProductivityImgPath))))

	var webRoot http.Handler
	if runtime.GOOS == "darwin" && vars.Packaged {
		appPath, _ := os.Executable()
		webRoot = http.FileServer(http.Dir(filepath.Dir(appPath) + "/../Frameworks/chipper/webroot"))
	} else if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		webRoot = http.FileServer(http.Dir(vars.AndroidPath + "/static/webroot"))
	} else {
		webRoot = http.FileServer(http.Dir("./webroot"))
	}
	http.Handle("/", DisableCachingAndSniffing(webRoot))
	fmt.Printf("Starting webserver at port " + vars.WebPort + " (http://localhost:" + vars.WebPort + ")\n")
	if err := http.ListenAndServe(":"+vars.WebPort, nil); err != nil {
		logger.Println("Error binding to " + vars.WebPort + ": " + err.Error())
		if vars.Packaged {
			logger.ErrMsg("FATAL: Wire-pod was unable to bind to port " + vars.WebPort + ". Another process is likely using it. Exiting.")
		}
		os.Exit(1)
	}
}

func GetLatestCommitSha() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/kercre123/wire-pod/commits", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get commits: %s", resp.Status)
	}
	type Commit struct {
		Sha string `json:"sha"`
	}
	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found")
	}
	return commits[0].Sha[:7], nil
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
			http.Error(w, "must request a cert by esn (ex. /session-certs/00e20145)", http.StatusBadRequest)
			return
		}
		esn := split[2]
		fileBytes, err := os.ReadFile(path.Join(vars.SessionCertPath, esn))
		if err != nil {
			http.Error(w, "cert does not exist", http.StatusNotFound)
			return
		}
		w.Write(fileBytes)
	}
}

func anyEmpty(values ...string) bool {
	for _, v := range values {
		if v == "" {
			return true
		}
	}
	return false
}

func isValidLanguage(language string, validLanguages []string) bool {
	for _, lang := range validLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

func isDownloadedLanguage(language string, downloadedLanguages []string) bool {
	for _, lang := range downloadedLanguages {
		if lang == language {
			return true
		}
	}
	return false
}
