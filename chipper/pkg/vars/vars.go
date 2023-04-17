package vars

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/kercre123/chipper/pkg/logger"
)

// initialize variables so they don't have to be found during runtime

const (
	JdocsPath         string = "./jdocs/jdocs.json"
	CustomIntentsPath string = "./customIntents.json"
	BotConfigsPath    string = "./botConfig.json"
	BotInfoPath       string = "./jdocs/botSdkInfo.json"
	PodName           string = "wire-pod"
	VoskModelPath     string = "../vosk/models/"
)

// /home/name/.anki_vector/
var SDKIniPath string
var BotJdocs []botjdoc
var BotInfo RobotInfoStore
var CustomIntents IntentsStruct
var CustomIntentsExist bool = false
var DownloadedVoskModels []string
var ValidVoskModels []string = []string{"en-US", "it-IT", "es-ES", "fr-FR", "de-DE","pt-BR"}

type RobotInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
		// 192.168.1.150:443
		GUID      string `json:"guid"`
		Activated bool   `json:"activated"`
	} `json:"robots"`
}

type IntentsStruct []struct {
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
}

type botjdoc struct {
	// vic:00000000
	Thing string `json:"thing"`
	// vic.RobotSettings, etc
	Name string `json:"name"`
	// actual jdoc
	Jdoc jdocspb.Jdoc `json:"jdoc"`
}

func Init() {
	logger.Println("Initializing variables")
	// figure out user SDK path, containing sdk_config.ini
	// has to be done like this because wire-pod is running as root
	// path should be /home/name/wire-pod/chipper
	// Split puts an extra / in the beginning of the array
	podPath, _ := os.Getwd()
	podPathSplit := strings.Split(strings.TrimSpace(podPath), "/")
	if podPathSplit[len(podPathSplit)-1] != "chipper" || podPathSplit[len(podPathSplit)-2] != PodName {
		logger.Println("It looks like you may have changed path names of the directories wire-pod is running in. This is not recommended because the SDK implementation depends on relativity in a few spots.")
	}
	if len(podPathSplit) >= 5 {
		SDKIniPath = "/" + podPathSplit[1] + "/" + podPathSplit[2] + "/.anki_vector/"
	} else if strings.EqualFold(podPathSplit[0], "root") {
		SDKIniPath = "/root/.anki_vector/"
	} else if len(podPathSplit) == 4 {
		SDKIniPath = "/" + podPathSplit[1] + "/.anki_vector/"
	} else {
		logger.Println("Unsupported path scenario, printing podPathSplit: ")
		logger.Println(podPathSplit)
		SDKIniPath = "/tmp/.anki_vector/"
	}
	logger.Println("SDK info path: " + SDKIniPath)

	// load api config (config.go)
	ReadConfig()

	// check models folder, add all models to DownloadedVoskModels
	if APIConfig.STT.Service == "vosk" {
		GetDownloadedVoskModels()
	}

	// load jdocs. if there are any in the old format, convert
	jdocsDir, err := os.ReadDir("./jdocs")
	oldJdocsExisted := false
	if err != nil {
		logger.Println("Error reading jdocs directory")
		logger.Println(err)
	} else {
		for _, file := range jdocsDir {
			if strings.Contains(file.Name(), "vic:") {
				oldJdocsExisted = true
				splitName := strings.Split(file.Name(), "-")
				thing := splitName[0]
				jdocname := splitName[1]
				splitJdocName := strings.Split(jdocname, ".")
				jdocname = strings.TrimSpace(splitJdocName[0] + "." + splitJdocName[1])
				jsonBytes, err := os.ReadFile("./jdocs/" + file.Name())
				if err != nil {
					logger.Println(err)
				} else {
					logger.Println("Appending " + file.Name() + " to new jdocs json")
					var newJdoc botjdoc
					var jdoc jdocspb.Jdoc
					newJdoc.Thing = thing
					newJdoc.Name = jdocname
					json.Unmarshal(jsonBytes, &jdoc)
					newJdoc.Jdoc = jdoc
					BotJdocs = append(BotJdocs, newJdoc)
					err := os.Remove("./jdocs/" + file.Name())
					if err != nil {
						logger.Println(err)
					}
				}
			}
		}
		if oldJdocsExisted {
			writeBytes, _ := json.Marshal(BotJdocs)
			os.WriteFile(JdocsPath, writeBytes, 0644)
			logger.Println("New jdocs file written")
		} else {
			if _, err := os.Stat(JdocsPath); err == nil {
				jsonBytes, _ := os.ReadFile(JdocsPath)
				json.Unmarshal(jsonBytes, &BotJdocs)
				logger.Println("Loaded jdocs file")
			}
		}
	}

	// load bot sdk info
	botBytes, err := os.ReadFile(BotInfoPath)
	if err == nil {
		json.Unmarshal(botBytes, &BotInfo)
		var botList []string
		for _, robot := range BotInfo.Robots {
			botList = append(botList, robot.Esn)
		}
		logger.Println("Loaded bot info file, known bots: " + fmt.Sprint(botList))
	}
	LoadCustomIntents()
}

func GetDownloadedVoskModels() {
	array, err := os.ReadDir("../vosk/models/")
	if err != nil {
		logger.Println(err)
		return
	}
	for _, dir := range array {
		DownloadedVoskModels = append(DownloadedVoskModels, dir.Name())
	}
}

func LoadCustomIntents() {
	jsonBytes, err := os.ReadFile(CustomIntentsPath)
	if err == nil {
		json.Unmarshal(jsonBytes, &CustomIntents)
		CustomIntentsExist = true
		logger.Println("Loaded custom intents:")
		for _, intent := range CustomIntents {
			logger.Println(intent.Name)
		}
	}
}

func WriteJdocs() {
	writeBytes, _ := json.Marshal(BotJdocs)
	os.WriteFile(JdocsPath, writeBytes, 0644)
}

// removes a bot from jdocs file
func DeleteData(thing string) {
	var newdocs []botjdoc
	for _, jdocentry := range BotJdocs {
		if jdocentry.Thing != thing {
			newdocs = append(newdocs, jdocentry)
		}
	}
	BotJdocs = newdocs
	WriteJdocs()
}

func GetJdoc(thing, jdocname string) (jdocspb.Jdoc, bool) {
	for _, botJdoc := range BotJdocs {
		if botJdoc.Name == jdocname && botJdoc.Thing == thing {
			return botJdoc.Jdoc, true
		}
	}
	return jdocspb.Jdoc{}, false
}

func AddJdoc(thing string, name string, jdoc jdocspb.Jdoc) uint64 {
	var latestVersion uint64 = 0
	matched := false
	for index, jdocentry := range BotJdocs {
		if jdocentry.Thing == thing && jdocentry.Name == name {
			BotJdocs[index].Jdoc = jdoc
			latestVersion = BotJdocs[index].Jdoc.DocVersion
			matched = true
			break
		}
	}
	if !matched {
		var newbot botjdoc
		newbot.Thing = thing
		newbot.Name = name
		newbot.Jdoc = jdoc
		BotJdocs = append(BotJdocs, newbot)
	}
	WriteJdocs()
	return latestVersion
}
