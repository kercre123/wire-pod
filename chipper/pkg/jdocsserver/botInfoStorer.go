package jdocsserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/peer"
	"gopkg.in/ini.v1"
)

var PerRobotSDKInfo struct {
	Esn       string `json:"esn"`
	IPAddress string `json:"ip_address"`
}

type RobotSDKInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
		GUID      string `json:"guid"`
	} `json:"robots"`
}

type options struct {
	SerialNo  string
	RobotName string `ini:"name"`
	CertPath  string `ini:"cert"`
	Target    string `ini:"ip"`
	Token     string `ini:"guid"`
}

func IniToJson() {
	var robotSDKInfo RobotSDKInfoStore
	eFileBytes, err := os.ReadFile("./jdocs/botSdkInfo.json")
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	}
	robotSDKInfo.GlobalGUID = "tni1TRsTRTaNSapjo0Y+Sw=="
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
						robot.GUID = cfg.Token
						robot.IPAddress = cfg.Target
					}
				}
				if !matched {
					fmt.Println("Adding " + cfg.SerialNo + " to JSON from INI")
					robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
						Esn       string `json:"esn"`
						IPAddress string `json:"ip_address"`
						GUID      string `json:"guid"`
					}{Esn: cfg.SerialNo, IPAddress: cfg.Target, GUID: cfg.Token})
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
							robot.GUID = cfg.Token
							robot.IPAddress = cfg.Target
						}
					}
					if !matched {
						fmt.Println("Adding " + cfg.SerialNo + " to JSON from INI")
						robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
							Esn       string `json:"esn"`
							IPAddress string `json:"ip_address"`
							GUID      string `json:"guid"`
						}{Esn: cfg.SerialNo, IPAddress: cfg.Target, GUID: cfg.Token})
					}
				}
			}
		} else {
			return
		}
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile("./jdocs/botSdkInfo.json", finalJsonBytes, 0644)
	fmt.Println("Ini to JSON finished")
}

func storeBotInfo(ctx context.Context, thing string) {
	fmt.Println("Storing bot info for later SDK use")
	var appendNew bool = true
	p, _ := peer.FromContext(ctx)
	ipAddr := strings.TrimSpace(strings.Split(p.Addr.String(), ":")[0])
	fmt.Println("Bot IP: `" + ipAddr + "`")
	botEsn := strings.TrimSpace(strings.Split(thing, ":")[1])
	fmt.Println("Bot ESN: `" + botEsn + "`")
	var robotSDKInfo RobotSDKInfoStore
	eFileBytes, err := os.ReadFile("./jdocs/botSdkInfo.json")
	if err == nil {
		json.Unmarshal(eFileBytes, &robotSDKInfo)
	}
	robotSDKInfo.GlobalGUID = "tni1TRsTRTaNSapjo0Y+Sw=="
	iniData, iniErr := ini.Load("../../.anki_vector/sdk_config.ini")
	for num, robot := range robotSDKInfo.Robots {
		if robot.Esn == botEsn {
			appendNew = false
			robotSDKInfo.Robots[num].IPAddress = ipAddr
			if iniErr != nil {
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
	if appendNew {
		robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
		}{Esn: botEsn, IPAddress: ipAddr, GUID: ""})
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile("./jdocs/botSdkInfo.json", finalJsonBytes, 0644)
	fmt.Println(string(finalJsonBytes))
}
