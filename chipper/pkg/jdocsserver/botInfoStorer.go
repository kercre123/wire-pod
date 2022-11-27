package jdocsserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/peer"
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
	} `json:"robots"`
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
	for _, robot := range robotSDKInfo.Robots {
		if robot.Esn == botEsn {
			appendNew = false
			robot.IPAddress = ipAddr
		}
	}
	if appendNew {
		robotSDKInfo.Robots = append(robotSDKInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
		}{Esn: botEsn, IPAddress: ipAddr})
	}
	finalJsonBytes, _ := json.Marshal(robotSDKInfo)
	os.WriteFile("./jdocs/botSdkInfo.json", finalJsonBytes, 0644)
	fmt.Println(string(finalJsonBytes))
}
