package sdkapp

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// This file serves one cached status call for the web UI home screen. Probing a robot can take
// many seconds when it is off or off the network, so the handler never waits on a robot: it
// returns what the last probe found and starts a background probe when that is stale.

const (
	statusMaxAge   = 5 * time.Second
	statusDialWait = 2500 * time.Millisecond
	statusProbeMax = 8 * time.Second
)

// BatteryInfo is the battery state in the terms the UI needs.
type BatteryInfo struct {
	Percent   int     `json:"percent"`
	Level     int     `json:"level"`
	Volts     float32 `json:"volts"`
	Charging  bool    `json:"charging"`
	OnCharger bool    `json:"on_charger"`
}

// EyeInfo is the robot's eye color as last reported in its settings jdoc.
type EyeInfo struct {
	Preset int     `json:"preset"`
	Custom bool    `json:"custom"`
	Hue    float64 `json:"hue,omitempty"`
	Sat    float64 `json:"sat,omitempty"`
}

// RobotStatus is one robot in the /api-sdk/robots_status response.
type RobotStatus struct {
	ESN         string       `json:"esn"`
	Name        string       `json:"name"`
	Personality string       `json:"personality"`
	IP          string       `json:"ip"`
	State       string       `json:"state"`            // online, offline or checking
	Reason      string       `json:"reason,omitempty"` // why it is offline: unreachable, timeout, auth, no_address, error
	Battery     *BatteryInfo `json:"battery,omitempty"`
	Eye         *EyeInfo     `json:"eye,omitempty"`
	LastSeen    int64        `json:"last_seen,omitempty"` // unix seconds, since this server started
	CheckedAt   int64        `json:"checked_at,omitempty"`
}

type probeResult struct {
	online  bool
	reason  string
	battery *BatteryInfo
}

type statusEntry struct {
	probed    bool
	probing   bool
	result    probeResult
	checkedAt time.Time
	lastSeen  time.Time
}

var statusCache = struct {
	sync.Mutex
	m map[string]*statusEntry
}{m: map[string]*statusEntry{}}

// probeRobot is a variable so tests can replace the real network probe.
var probeRobot = realProbe

type knownRobot struct {
	esn, ip, name, personality string
}

func knownRobots() []knownRobot {
	vars.BotInfoMu.RLock()
	defer vars.BotInfoMu.RUnlock()
	list := make([]knownRobot, 0, len(vars.BotInfo.Robots))
	for _, r := range vars.BotInfo.Robots {
		list = append(list, knownRobot{esn: strings.ToLower(r.Esn), ip: r.IPAddress, name: r.Name, personality: r.Personality})
	}
	return list
}

// robotStatuses returns the current status of every known robot and starts probes for stale ones.
func robotStatuses(now time.Time) []RobotStatus {
	robotsKnown := knownRobots()
	out := make([]RobotStatus, 0, len(robotsKnown))
	statusCache.Lock()
	defer statusCache.Unlock()
	live := map[string]bool{}
	for _, k := range robotsKnown {
		live[k.esn] = true
		e := statusCache.m[k.esn]
		if e == nil {
			e = &statusEntry{}
			statusCache.m[k.esn] = e
		}
		if !e.probing && (!e.probed || now.Sub(e.checkedAt) > statusMaxAge) {
			e.probing = true
			go runProbe(probeRobot, k.esn, k.ip)
		}
		s := RobotStatus{ESN: k.esn, Name: k.name, Personality: k.personality, IP: hostOnly(k.ip), State: "checking", Eye: eyeFromJdoc(k.esn)}
		if e.probed {
			s.State = "offline"
			s.Reason = e.result.reason
			if e.result.online {
				s.State = "online"
				s.Reason = ""
				s.Battery = e.result.battery
			}
			s.CheckedAt = e.checkedAt.Unix()
		}
		if !e.lastSeen.IsZero() {
			s.LastSeen = e.lastSeen.Unix()
		}
		out = append(out, s)
	}
	for esn := range statusCache.m {
		if !live[esn] {
			delete(statusCache.m, esn)
		}
	}
	return out
}

func runProbe(probe func(esn, ip string) probeResult, esn, ip string) {
	res := probe(esn, ip)
	statusCache.Lock()
	defer statusCache.Unlock()
	e := statusCache.m[esn]
	if e == nil {
		return
	}
	e.probing = false
	e.probed = true
	e.result = res
	e.checkedAt = time.Now()
	if res.online {
		e.lastSeen = e.checkedAt
	}
}

func hostOnly(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

// realProbe checks that the robot's address answers, then asks it for its battery.
func realProbe(esn, ip string) probeResult {
	if hostOnly(ip) == "" {
		return probeResult{reason: "no_address"}
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostOnly(ip), "443"), statusDialWait)
	if err != nil {
		return probeResult{reason: "unreachable"}
	}
	conn.Close()
	type result struct {
		res probeResult
	}
	done := make(chan result, 1) // buffered so a probe that outlives its timeout doesn't leak
	go func() {
		robot, idx, err := getRobot(esn)
		if err != nil {
			reason := "error"
			if s := strings.ToLower(err.Error()); strings.Contains(s, "unauthenticated") || strings.Contains(s, "token") {
				reason = "auth"
			}
			done <- result{probeResult{reason: reason}}
			return
		}
		ctx, cancel := context.WithTimeout(robot.Ctx, statusProbeMax-time.Second)
		defer cancel()
		resp, err := robot.Vector.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			done <- result{probeResult{reason: "unreachable"}}
			return
		}
		if idx < len(robots) {
			robots[idx].ConnTimer = 0 // like any other API call, a status check keeps the connection open
		}
		done <- result{probeResult{online: true, battery: batteryInfo(resp.BatteryVolts, int(resp.BatteryLevel), resp.IsCharging, resp.IsOnChargerPlatform, resp.SuggestedChargerSec)}}
	}()
	select {
	case r := <-done:
		return r.res
	case <-time.After(statusProbeMax):
		return probeResult{reason: "timeout"}
	}
}

// batteryInfo estimates charge from the pack voltage. The curve and the adjustments to
// the reported level match what the classic UI has always shown.
func batteryInfo(volts float32, level int, charging, onCharger bool, suggestedChargerSec float32) *BatteryInfo {
	const maxV, midV, minV = 4.1, 3.85, 3.5
	v := float64(volts)
	var pct float64
	switch {
	case v >= maxV:
		pct = 100
	case v >= midV:
		pct = 80 + 20*math.Log10(1+(v-midV)/(maxV-midV)*9)
	case v >= minV:
		pct = 80 * math.Log10(1+(v-minV)/(midV-minV)*9)
	case v == 0:
		// a robot switched on off the charger doesn't report voltage; assume a reasonable value
		pct = 70
	default:
		pct = 0
	}
	p := int(math.Round(math.Max(0, math.Min(100, pct))))
	switch {
	case level == 2 && p < 20:
		level = 0
	case level == 2 && p < 50:
		level = 1
	case level == 1 && !onCharger:
		if p > 15 {
			p = 15
		}
		level = 0
	}
	if onCharger && !charging && suggestedChargerSec == 0 {
		p = 100
	}
	return &BatteryInfo{Percent: p, Level: level, Volts: volts, Charging: charging, OnCharger: onCharger}
}

// eyeFromJdoc reads the eye color from the robot settings wire-pod last saw, without contacting the robot.
func eyeFromJdoc(esn string) *EyeInfo {
	doc, ok := vars.GetJdoc("vic:"+esn, "vic.RobotSettings")
	if !ok || doc.JsonDoc == "" {
		return nil
	}
	var settings struct {
		EyeColor *int `json:"eye_color"`
		Custom   *struct {
			Enabled bool    `json:"enabled"`
			Hue     float64 `json:"hue"`
			Sat     float64 `json:"saturation"`
		} `json:"custom_eye_color"`
	}
	if json.Unmarshal([]byte(doc.JsonDoc), &settings) != nil || (settings.EyeColor == nil && settings.Custom == nil) {
		return nil
	}
	eye := &EyeInfo{}
	if settings.EyeColor != nil {
		eye.Preset = *settings.EyeColor
	}
	if settings.Custom != nil && settings.Custom.Enabled {
		eye.Custom, eye.Hue, eye.Sat = true, settings.Custom.Hue, settings.Custom.Sat
	}
	return eye
}

// handleRobotStatusAPI serves /api-sdk/robots_status. It needs no robot connection.
func handleRobotStatusAPI(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/api-sdk/robots_status" {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Robots    []RobotStatus `json:"robots"`
		CheckedAt int64         `json:"checked_at"`
	}{robotStatuses(time.Now()), time.Now().Unix()})
	return true
}
