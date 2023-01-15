package jdocsserver

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kercre123/chipper/pkg/logger"
	tokenserver "github.com/kercre123/chipper/pkg/servers/token"
	"google.golang.org/grpc/peer"
	"gopkg.in/ini.v1"
)

const (
	BotInfoFile = tokenserver.BotInfoFile
	InfoPath    = JdocsPath + BotInfoFile
)

type options struct {
	SerialNo  string
	RobotName string `ini:"name"`
	CertPath  string `ini:"cert"`
	Target    string `ini:"ip"`
	Token     string `ini:"guid"`
}

func IsBotInInfo(esn string) bool {
	esn = strings.TrimSpace(strings.ToLower(esn))
	var botInfo tokenserver.RobotInfoStore
	fileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(fileBytes, &botInfo)
		for _, robot := range botInfo.Robots {
			if esn == strings.TrimSpace(strings.ToLower(robot.Esn)) {
				return true
			}
		}
	}
	return false
}

// This function write a bot name, esn, guid, and target to sdk_config.ini
// Should only be used for primary auth
// IP should just be "xxx.xxx.xxx.xxx", no port
func WriteToIniPrimary(botName, esn, guid, ip string) {
	fullPath, _ := os.Getwd()
	fullPath = strings.TrimSuffix(fullPath, "/wire-pod/chipper") + "/.anki_vector/"
	userIniData, err := ini.Load(fullPath + "sdk_config.ini")
	if err != nil {
		os.Mkdir(fullPath, 0755)
		userIniData = ini.Empty()
	}
	matched := false
	for _, section := range userIniData.Sections() {
		if strings.EqualFold(section.Name(), esn) {
			matched = true
			logger.Println("WriteToIniPrimary: bot already in INI matched, setting info")
			section.Key("cert").SetValue(fullPath + botName + "-" + esn + ".cert")
			section.Key("name").SetValue(botName)
			section.Key("ip").SetValue(ip)
			section.Key("guid").SetValue(guid)
		}
	}
	if !matched {
		logger.Println("WriteToIniPrimary: ESN did not match any section in sdk config file, creating")
		newSection, err := userIniData.NewSection(esn)
		if err != nil {
			logger.Println(err)
		}
		logger.Println("Getting session cert from Anki server")
		newSection.NewKey("cert", fullPath+botName+"-"+esn+".cert")
		newSection.NewKey("ip", ip)
		newSection.NewKey("name", botName)
		newSection.NewKey("guid", guid)
	}
	userIniData.SaveTo(fullPath + "sdk_config.ini")
	logger.Println("WriteToIniPrimary: successfully wrote INI")
}

// Less information is given with a secondary auth request, we have to get the cert from the Anki servers
// If a cert is already there, it does not get the Anki server cert
func WriteToIniSecondary(esn, guid, ip string) {
	certPath := ""
	botName := ""
	certExists := false
	fullPath, _ := os.Getwd()
	fullPath = strings.TrimSuffix(fullPath, "/wire-pod/chipper") + "/.anki_vector/"
	userIniData, err := ini.Load(fullPath + "sdk_config.ini")
	if err != nil {
		os.Mkdir(fullPath, 0755)
		userIniData = ini.Empty()
	}
	// see if cert already exists, get name from that if it does
	// if not, get from ANKI servers
	for _, section := range userIniData.Sections() {
		if strings.EqualFold(section.Name(), esn) {
			logger.Println("WriteToIniSecondary: Name found from ESN in INI, setting info")
			botNameKey, _ := section.GetKey("name")
			botName = botNameKey.String()
			certPath = fullPath + botName + "-" + esn + ".cert"
			certExists = true
			// set information
			section.Key("guid").SetValue(guid)
			section.Key("ip").SetValue(ip)
		}
	}
	if !certExists {
		logger.Println("WriteToIniSecondary: getting session cert from Anki server")
		resp, err := http.Get("https://session-certs.token.global.anki-services.com/vic/" + esn)
		if err != nil {
			logger.Println(err)
		}
		certBytesOrig, _ := io.ReadAll(resp.Body)
		block, _ := pem.Decode(certBytesOrig)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			logger.Println(err)
		}
		botName = cert.Issuer.CommonName
		certPath = fullPath + botName + "-" + esn + ".cert"
		out, err := os.Create(certPath)
		if err == nil {
			out.Write(certBytesOrig)
		}
	}
	logger.Println("WriteToIniSecondary: robot name is " + botName)

	// create an entry
	if !certExists {
		logger.Println("WriteToIniSecondary: creating INI entry")
		newSection, err := userIniData.NewSection(esn)
		if err != nil {
			logger.Println(err)
		}
		newSection.NewKey("cert", certPath)
		newSection.NewKey("ip", ip)
		newSection.NewKey("name", botName)
		newSection.NewKey("guid", guid)
	}
	logger.Println("WriteToIniSecondary complete")
	userIniData.SaveTo("../../.anki_vector/sdk_config.ini")
}

func IniToJson() {
	var robotSDKInfo tokenserver.RobotInfoStore
	eFileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	}
	robotSDKInfo.GlobalGUID = tokenserver.GlobalGUID
	iniData, err := ini.Load("../../.anki_vector/sdk_config.ini")
	if err == nil {
		for _, section := range iniData.Sections() {
			cfg := options{}
			section.MapTo(&cfg)
			cfg.SerialNo = section.Name()
			if cfg.SerialNo != "DEFAULT" {
				matched := false
				for _, robot := range robotSDKInfo.Robots {
					if cfg.SerialNo == robot.Esn {
						matched = true
						if robot.GUID == "" {
							robot.GUID = cfg.Token
							robot.IPAddress = cfg.Target
						}
					}
				}
				if !matched {
					logger.Println("Adding " + cfg.SerialNo + " to JSON from INI")
					robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
						Esn       string `json:"esn"`
						IPAddress string `json:"ip_address"`
						GUID      string `json:"guid"`
						Activated bool   `json:"activated"`
					}{Esn: cfg.SerialNo, IPAddress: cfg.Target, GUID: cfg.Token, Activated: false})
				}
			}
		}
	} else {
		iniData, err = ini.Load("/root/.anki_vector/sdk_config.ini")
		if err == nil {
			for _, section := range iniData.Sections() {
				cfg := options{}
				section.MapTo(&cfg)
				cfg.SerialNo = section.Name()
				if cfg.SerialNo != "DEFAULT" {
					matched := false
					for _, robot := range robotSDKInfo.Robots {
						if cfg.SerialNo == robot.Esn {
							matched = true
							if robot.GUID == "" {
								robot.GUID = cfg.Token
								robot.IPAddress = cfg.Target
							}
						}
					}
					if !matched {
						logger.Println("Adding " + cfg.SerialNo + " to JSON from INI")
						robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
							Esn       string `json:"esn"`
							IPAddress string `json:"ip_address"`
							GUID      string `json:"guid"`
							Activated bool   `json:"activated"`
						}{Esn: cfg.SerialNo, IPAddress: cfg.Target, GUID: cfg.Token, Activated: false})
					}
				}
			}
		} else {
			return
		}
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile(InfoPath, finalJsonBytes, 0644)
	logger.Println("Ini to JSON finished")
}

func StoreBotInfo(ctx context.Context, thing string) {
	var appendNew bool = true
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.TrimSpace(strings.Split(p.Addr.String(), ":")[0])
	botEsn := strings.TrimSpace(strings.Split(thing, ":")[1])
	var robotSDKInfo tokenserver.RobotInfoStore
	eFileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	}
	robotSDKInfo.GlobalGUID = "tni1TRsTRTaNSapjo0Y+Sw=="
	for num, robot := range robotSDKInfo.Robots {
		if robot.Esn == botEsn {
			appendNew = false
			robotSDKInfo.Robots[num].IPAddress = ipAddr
		}
	}
	if appendNew {
		logger.Println("Adding " + botEsn + " to bot info store")
		robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
			Activated bool   `json:"activated"`
		}{Esn: botEsn, IPAddress: ipAddr, GUID: "", Activated: false})
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile(InfoPath, finalJsonBytes, 0644)
}

func StoreBotInfoStrings(target string, botEsn string) {
	logger.Println("Storing bot info for later SDK use")
	var appendNew bool = true
	ipAddr := strings.TrimSpace(strings.Split(target, ":")[0])
	var robotSDKInfo tokenserver.RobotInfoStore
	eFileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	}
	robotSDKInfo.GlobalGUID = "tni1TRsTRTaNSapjo0Y+Sw=="
	for num, robot := range robotSDKInfo.Robots {
		if robot.Esn == botEsn {
			appendNew = false
			robotSDKInfo.Robots[num].IPAddress = ipAddr
		}
	}
	if appendNew {
		logger.Println("Adding " + botEsn + " to bot info store")
		robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
			Activated bool   `json:"activated"`
		}{Esn: botEsn, IPAddress: ipAddr, GUID: "", Activated: false})
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile(InfoPath, finalJsonBytes, 0644)
}
