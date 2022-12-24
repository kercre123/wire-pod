package jdocsserver

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/digital-dream-labs/chipper/pkg/tokenserver"
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

func WriteToIni(botName string) {
	// 	[008060ec]
	// cert = /home/kerigan/.anki_vector/Vector-B6H9-008060ec.cert
	// ip = 192.168.1.155
	// name = Vector-B6H9
	// guid = 1YbXk1yrS9C1I78snYy8xA==

	var robotSDKInfo tokenserver.RobotInfoStore
	eFileBytes, err := os.ReadFile(InfoPath)
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	} else {
		return
	}
	userIniData, err := ini.Load("../../.anki_vector/sdk_config.ini")
	fullPath, _ := os.Getwd()
	fullPath = strings.TrimSuffix(fullPath, "/wire-pod/chipper") + "/.anki_vector/"
	if err != nil {
		os.Mkdir(fullPath, 0755)
		userIniData = ini.Empty()
	}
	for _, robot := range robotSDKInfo.Robots {
		matched := false
		for _, section := range userIniData.Sections() {
			if section.Name() == robot.Esn {
				matched = true
				if botName != "" {
					section.Key("cert").SetValue(fullPath + botName + "-" + robot.Esn + ".cert")
					section.Key("name").SetValue(botName)
				}
				section.Key("ip").SetValue(robot.IPAddress)
				if robot.GUID == "" {
					section.Key("guid").SetValue(robotSDKInfo.GlobalGUID)
				} else {
					section.Key("guid").SetValue(robot.GUID)
				}
			}
		}
		if !matched {
			newSection, err := userIniData.NewSection(robot.Esn)
			if err != nil {
				fmt.Println(err)
			}
			setGuid := robot.GUID
			if robot.GUID == "" {
				setGuid = robotSDKInfo.GlobalGUID
			}
			fmt.Println("Getting session cert from Anki server")
			resp, err := http.Get("https://session-certs.token.global.anki-services.com/vic/" + robot.Esn)
			if err != nil {
				fmt.Println(err)
			}
			certBytesOrig, _ := io.ReadAll(resp.Body)
			block, _ := pem.Decode(certBytesOrig)
			certBytes := block.Bytes
			cert, err := x509.ParseCertificate(certBytes)
			if err != nil {
				fmt.Println(err)
			}
			botName = cert.Issuer.CommonName
			out, err := os.Create(fullPath + botName + "-" + robot.Esn + ".cert")
			if err == nil {
				out.Write(certBytesOrig)
			}
			newSection.NewKey("cert", fullPath+botName+"-"+robot.Esn+".cert")
			newSection.NewKey("ip", robot.IPAddress)
			newSection.NewKey("name", botName)
			newSection.NewKey("guid", setGuid)
		}
	}
	fmt.Println("JSON to ini done")
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
					fmt.Println("Adding " + cfg.SerialNo + " to JSON from INI")
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
						fmt.Println("Adding " + cfg.SerialNo + " to JSON from INI")
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
	fmt.Println("Ini to JSON finished")
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
		fmt.Println("Adding " + botEsn + " to bot info store")
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
	fmt.Println("Storing bot info for later SDK use")
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
		fmt.Println("Adding " + botEsn + " to bot info store")
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
