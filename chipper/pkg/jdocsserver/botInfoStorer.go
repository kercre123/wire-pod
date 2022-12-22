package jdocsserver

import (
	"context"
	"encoding/json"
	"fmt"
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
					if robot.GUID == "" {
						if cfg.SerialNo == robot.Esn {
							matched = true
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
						if robot.GUID == "" {
							if cfg.SerialNo == robot.Esn {
								matched = true
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
	iniData, iniErr := ini.Load("../../.anki_vector/sdk_config.ini")
	for num, robot := range robotSDKInfo.Robots {
		if robot.Esn == botEsn {
			appendNew = false
			robotSDKInfo.Robots[num].IPAddress = ipAddr
			if robotSDKInfo.Robots[num].GUID == "" {
				if iniErr == nil {
					section := iniData.Section(botEsn)
					if section != nil {
						cfg := options{}
						section.MapTo(&cfg)
						robotSDKInfo.Robots[num].GUID = cfg.Token
						fmt.Println("Found GUID in ini, " + cfg.Token)
					}
				}
			}
		}
	}
	if appendNew {
		robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
			Activated bool   `json:"activated"`
		}{Esn: botEsn, IPAddress: ipAddr, GUID: "", Activated: false})
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile(InfoPath, finalJsonBytes, 0644)
	fmt.Println(string(finalJsonBytes))
}
