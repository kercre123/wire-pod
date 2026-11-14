package jdocsserver

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc/peer"
	"gopkg.in/ini.v1"
)

func IsBotInInfo(esn string) bool {
	for _, robot := range vars.GetBotInfo().Robots {
		if esn == strings.TrimSpace(strings.ToLower(robot.Esn)) {
			return true
		}
	}
	return false
}

// This function write a bot name, esn, guid, and target to sdk_config.ini
// Should only be used for primary auth
// IP should just be "xxx.xxx.xxx.xxx", no port
func WriteToIniPrimary(botName, esn, guid, ip string) {
	userIniData, err := ini.Load(vars.SDKIniPath + "sdk_config.ini")
	if err != nil {
		logger.Println("Creating " + vars.SDKIniPath + " directory")
		os.Mkdir(vars.SDKIniPath, 0755)
		userIniData = ini.Empty()
	}
	matched := false
	for _, section := range userIniData.Sections() {
		if strings.EqualFold(section.Name(), esn) {
			matched = true
			logger.Println("WriteToIniPrimary: bot already in INI matched, setting info")
			section.Key("cert").SetValue(vars.SDKIniPath + botName + "-" + esn + ".cert")
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
		newSection.NewKey("cert", vars.SDKIniPath+botName+"-"+esn+".cert")
		newSection.NewKey("ip", ip)
		newSection.NewKey("name", botName)
		newSection.NewKey("guid", guid)
	}
	userIniData.SaveTo(vars.SDKIniPath + "sdk_config.ini")
	logger.Println("WriteToIniPrimary: successfully wrote INI")
}

// Less information is given with a secondary auth request, we have to get the cert from the Anki servers
// If a cert is already there, it does not get the Anki server cert
func WriteToIniSecondary(esn, guid, ip string) {
	certPath := ""
	botName := ""
	certExists := false
	userIniData, err := ini.Load(vars.SDKIniPath + "sdk_config.ini")
	if err != nil {
		logger.Println("Creating " + vars.SDKIniPath + " directory")
		os.Mkdir(vars.SDKIniPath, 0755)
		userIniData = ini.Empty()
	}
	// see if cert already exists, get name from that if it does
	// if not, get from ANKI servers
	for _, section := range userIniData.Sections() {
		if strings.EqualFold(section.Name(), esn) {
			logger.Println("WriteToIniSecondary: Name found from ESN in INI, setting info")
			botNameKey, _ := section.GetKey("name")
			botName = botNameKey.String()
			certPath = vars.SDKIniPath + botName + "-" + esn + ".cert"
			certExists = true
			// set information
			section.Key("guid").SetValue(guid)
			section.Key("ip").SetValue(ip)
		}
	}
	if !certExists {
		logger.Println("WriteToIniSecondary: getting session cert from DDL server")
		resp, err := http.Get("https://session-certs.token.global.anki-services.com/vic/" + esn)
		if err != nil {
			logger.Println(err)
			logger.Println("The DDL servers are down at the moment. The cert will not be gotten. The Python SDK will not be configured.")
			return
		}
		defer resp.Body.Close()
		certBytesOrig, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Println(err)
			return
		}
		block, _ := pem.Decode(certBytesOrig)
		if block == nil {
			logger.Println("WriteToIniSecondary: session cert from DDL server isn't valid PEM")
			return
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			logger.Println(err)
			return
		}
		botName = cert.Issuer.CommonName
		if !vars.ValidRobotName(botName) {
			logger.Println("WriteToIniSecondary: invalid robot name in session cert: " + botName)
			return
		}
		vars.WriteFileAtomic(filepath.Join(vars.SessionCertPath, esn), certBytesOrig, 0644)
		certPath = filepath.Join(vars.SDKIniPath, botName+"-"+esn+".cert")
		if err := vars.WriteFileAtomic(certPath, certBytesOrig, 0644); err != nil {
			logger.Println(err)
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
	userIniData.SaveTo(vars.SDKIniPath + "sdk_config.ini")
}

func StoreBotInfo(ctx context.Context, thing string) {
	var appendNew bool = true
	p, _ := peer.FromContext(ctx)
	ipAddr := vars.HostOnly(p.Addr.String())
	botEsn := strings.TrimSpace(strings.Split(thing, ":")[1])
	err := vars.UpdateBotInfo(func(b *vars.RobotInfoStore) {
		b.GlobalGUID = "tni1TRsTRTaNSapjo0Y+Sw=="
		for num, robot := range b.Robots {
			if robot.Esn == botEsn {
				appendNew = false
				b.Robots[num].IPAddress = ipAddr
			}
		}
		if appendNew {
			logger.Println("Adding " + botEsn + " to bot info store")
			b.Robots = append(b.Robots, vars.BotInfoRobot{Esn: botEsn, IPAddress: ipAddr})
		}
	})
	if err != nil {
		logger.Println("Error writing bot info: ", err)
	}
}
