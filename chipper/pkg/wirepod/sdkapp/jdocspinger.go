package sdkapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/botconn"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/zeroconf"
)

var JdocsPingerBots struct {
	mu     sync.Mutex
	Robots []JdocsPingerRobot
}

type JdocsPingerRobot struct {
	ESN                string `json:"esn"`
	GUID               string `json:"guid"`
	IP                 string `json:"ip"`
	TimeSinceLastCheck int    `json:"timesince"`
	Stopped            bool   `json:"stopped"`
}

// the escape pod CA cert only gets appended to the cert store when a jdocs connection is created
// this doesn't happen at every boot
// this utilizes Vector's connCheck to see if a bot has disconnected from the server for more than 10 seconds
// if it has, it will pull jdocs from the bot which will cause the CA cert to get appended to the store

// setting JDOCS_PINGER_ENABLED=false will disable jdocs pinger
var PingerEnabled bool = true

const jdocsPingDelay = 3 * time.Second

func pingJdocs(target string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	target = vars.HostOnly(target)
	var serial string
	matched := false
	for _, robot := range vars.GetBotInfo().Robots {
		if strings.TrimSpace(strings.ToLower(robot.IPAddress)) == strings.TrimSpace(strings.ToLower(target)) {
			matched = true
			serial = robot.Esn
		}
	}
	if !matched {
		logger.Println("jdocs pinger error: serial did not match any bot in bot json")
		return
	}
	robotTmp, err := botconn.Get(serial)
	if err != nil {
		logger.Println(err)
		return
	}
	_, err = robotTmp.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
	if err != nil {
		var closeRobot func()
		robotTmp, closeRobot, err = botconn.DialWithGUID(serial, vars.GetBotInfo().GlobalGUID)
		if err != nil {
			logger.Println(err)
			logger.Println("Error pinging jdocs")
			return
		}
		defer closeRobot()
		_, err = robotTmp.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			logger.Println("Error pinging jdocs, likely unauthenticated")
			return
		}
	}
	resp, err := robotTmp.Conn.PullJdocs(ctx, &vectorpb.PullJdocsRequest{
		JdocTypes: []vectorpb.JdocType{vectorpb.JdocType_ROBOT_SETTINGS},
	})
	if err != nil {
		logger.Println("Failed to pull jdocs: ", err)
		return
	}
	if len(resp.NamedJdocs) == 0 || resp.NamedJdocs[0].Doc == nil {
		logger.Println("Bot " + serial + " returned no jdocs")
		return
	}
	logger.Println("Successfully got jdocs from " + serial)
	// write to file
	var jdoc vars.AJdoc
	jdoc.DocVersion = resp.NamedJdocs[0].Doc.DocVersion
	jdoc.FmtVersion = resp.NamedJdocs[0].Doc.FmtVersion
	jdoc.ClientMetadata = resp.NamedJdocs[0].Doc.ClientMetadata
	jdoc.JsonDoc = resp.NamedJdocs[0].Doc.JsonDoc
	vars.AddJdoc("vic:"+serial, "vic.RobotSettings", jdoc)
}

func InitJdocsPinger() {
	if os.Getenv("JDOCS_PINGER_ENABLED") == "false" {
		logger.Println("Jdocs pinger is disabled (JDOCS_PINGER_ENABLED=false)")
		PingerEnabled = false
		return
	}
	fmt.Println("Starting jdocs pinger ticker")
	go func() {
		for {
			JdocsPingerBots.mu.Lock()
			for i, bot := range JdocsPingerBots.Robots {
				if !bot.Stopped {
					JdocsPingerBots.Robots[i].TimeSinceLastCheck = JdocsPingerBots.Robots[i].TimeSinceLastCheck + 1
					if JdocsPingerBots.Robots[i].TimeSinceLastCheck > 15 {
						logger.Println("Haven't received a conn check from " + bot.ESN + " in 15 seconds, will ping jdocs on next check")
						JdocsPingerBots.Robots[i].Stopped = true
					}
				}
			}
			JdocsPingerBots.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()
}

func ShouldPingJdocs(target string) bool {
	var esn, guid, botip string
	matched := false
	for _, bot := range vars.GetBotInfo().Robots {
		if target == bot.IPAddress {
			esn = bot.Esn
			guid = bot.GUID
			botip = bot.IPAddress
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	JdocsPingerBots.mu.Lock()
	defer JdocsPingerBots.mu.Unlock()
	for i, bot := range JdocsPingerBots.Robots {
		if esn == bot.ESN {
			if bot.Stopped {
				JdocsPingerBots.Robots[i].TimeSinceLastCheck = 0
				JdocsPingerBots.Robots[i].Stopped = false
				return true
			} else {
				JdocsPingerBots.Robots[i].TimeSinceLastCheck = 0
				return false
			}
		}
	}
	// below will only execute if esn doesn't match bots in list
	newBot := JdocsPingerRobot{
		ESN:                esn,
		GUID:               guid,
		IP:                 botip,
		TimeSinceLastCheck: 0,
		Stopped:            false,
	}
	JdocsPingerBots.Robots = append(JdocsPingerBots.Robots, newBot)
	return true
}

func connCheck(w http.ResponseWriter, r *http.Request) {
	switch {
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	case strings.Contains(r.URL.Path, "/ok"):
		if r.FormValue("runMDNS") == "true" {
			RunMDNS("t")
			fmt.Fprintf(w, "ran")
			return
		}
		robotTarget := vars.HostOnly(r.RemoteAddr)
		vars.MarkBotSeenByIP(robotTarget)
		if PingerEnabled {
			known := false
			for _, bot := range vars.GetBotInfo().Robots {
				if bot.IPAddress == robotTarget {
					known = true
					break
				}
			}
			if known {
				if ShouldPingJdocs(robotTarget) {
					go func() {
						time.Sleep(jdocsPingDelay)
						pingJdocs(robotTarget)
					}()
				}
			} else {
				go RunMDNS(robotTarget)
			}
		}
		fmt.Fprintf(w, "ok")
		return
	}
}

var (
	mdnsAlreadyRunMu sync.Mutex
	mdnsAlreadyRun   []string
)

var RunningMDNS bool

func RunMDNS(botIP string) {
	mdnsAlreadyRunMu.Lock()
	alreadyRun := slices.Contains(mdnsAlreadyRun, botIP)
	mdnsAlreadyRunMu.Unlock()
	if alreadyRun {
		return
	}
	fmt.Println("Running mDNS...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	entries := make(chan *zeroconf.ServiceEntry)
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		logger.Println("mDNS resolver error: ", err)
		return
	}
	resolver.Browse(ctx, "_ankivector._tcp", "local", entries)
	for entry := range entries {
		if len(entry.AddrIPv4) == 0 {
			continue
		}
		robotID := strings.Split(entry.HostName, ".")[0]
		ip := entry.AddrIPv4[0].String()
		rinf, matched := vars.FindRecurringInfoByID(robotID)
		if !matched {
			mdnsAlreadyRunMu.Lock()
			if !slices.Contains(mdnsAlreadyRun, botIP) {
				mdnsAlreadyRun = append(mdnsAlreadyRun, botIP)
			}
			mdnsAlreadyRunMu.Unlock()
			continue
		}
		vars.AddToRInfo(rinf.ESN, robotID, ip)
		fmt.Println("Updating robot " + robotID)
		if err := vars.SetBotIP(rinf.ESN, ip); err != nil {
			logger.Println("Error writing bot info: ", err)
		}
		go func() {
			if vars.Config().Server.EPConfig {
				time.Sleep(time.Second)
			}
			pingJdocs(ip)
		}()
	}
	fmt.Println("Done running mDNS")
}
