package sdkapp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func setBots(t *testing.T, bots ...[3]string) {
	t.Helper()
	vars.BotInfoMu.Lock()
	old := vars.BotInfo
	vars.BotInfo.Robots = nil
	for _, b := range bots {
		vars.BotInfo.Robots = append(vars.BotInfo.Robots, struct {
			Esn         string `json:"esn"`
			IPAddress   string `json:"ip_address"`
			GUID        string `json:"guid"`
			Activated   bool   `json:"activated"`
			Name        string `json:"name,omitempty"`
			Personality string `json:"personality,omitempty"`
		}{Esn: b[0], IPAddress: b[1], Name: b[2]})
	}
	vars.BotInfoMu.Unlock()
	statusCache.Lock()
	statusCache.m = map[string]*statusEntry{}
	statusCache.Unlock()
	t.Cleanup(func() {
		vars.BotInfoMu.Lock()
		vars.BotInfo = old
		vars.BotInfoMu.Unlock()
	})
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

func TestRobotStatusesProbeInBackground(t *testing.T) {
	setBots(t, [3]string{"00A1", "10.0.0.5:443", "Pip"}, [3]string{"00b2", "", "Bolt"})
	var mu sync.Mutex
	probes := map[string]int{}
	release := make(chan struct{})
	old := probeRobot
	probeRobot = func(esn, ip string) probeResult {
		mu.Lock()
		probes[esn]++
		mu.Unlock()
		<-release
		if esn == "00a1" {
			return probeResult{online: true, battery: &BatteryInfo{Percent: 80}}
		}
		return probeResult{reason: "unreachable"}
	}
	t.Cleanup(func() { probeRobot = old })

	// The first call must return at once, without waiting for any probe.
	start := time.Now()
	first := robotStatuses(time.Now())
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("robotStatuses waited on a probe")
	}
	if len(first) != 2 || first[0].State != "checking" || first[1].State != "checking" {
		t.Fatalf("want two robots still checking, got %+v", first)
	}
	if first[0].ESN != "00a1" || first[0].Name != "Pip" || first[0].IP != "10.0.0.5" {
		t.Fatalf("identity fields wrong: %+v", first[0])
	}
	// A second call while probes run must not start more.
	robotStatuses(time.Now())
	close(release)
	waitFor(t, func() bool {
		s := robotStatuses(time.Now())
		return s[0].State == "online" && s[1].State == "offline"
	})
	s := robotStatuses(time.Now())
	if s[0].Battery == nil || s[0].Battery.Percent != 80 || s[0].LastSeen == 0 {
		t.Fatalf("online robot should carry battery and last_seen: %+v", s[0])
	}
	if s[1].Reason != "unreachable" || s[1].Battery != nil || s[1].LastSeen != 0 {
		t.Fatalf("offline robot wrong: %+v", s[1])
	}
	mu.Lock()
	defer mu.Unlock()
	if probes["00a1"] != 1 || probes["00b2"] != 1 {
		t.Fatalf("each robot should be probed once while fresh, got %v", probes)
	}
}

func TestRobotStatusesReprobeWhenStaleAndKeepLastSeen(t *testing.T) {
	setBots(t, [3]string{"00a1", "10.0.0.5", ""})
	online := true
	var mu sync.Mutex
	old := probeRobot
	probeRobot = func(esn, ip string) probeResult {
		mu.Lock()
		defer mu.Unlock()
		if online {
			return probeResult{online: true}
		}
		return probeResult{reason: "timeout"}
	}
	t.Cleanup(func() { probeRobot = old })
	robotStatuses(time.Now())
	waitFor(t, func() bool { return robotStatuses(time.Now())[0].State == "online" })
	seen := robotStatuses(time.Now())[0].LastSeen
	mu.Lock()
	online = false
	mu.Unlock()
	later := time.Now().Add(2 * statusMaxAge)
	robotStatuses(later)
	waitFor(t, func() bool { return robotStatuses(later)[0].State == "offline" })
	got := robotStatuses(later)[0]
	if got.Reason != "timeout" || got.LastSeen != seen {
		t.Fatalf("offline robot should keep last_seen %d and its reason, got %+v", seen, got)
	}
}

func TestRobotStatusesDropsForgottenRobots(t *testing.T) {
	setBots(t, [3]string{"00a1", "10.0.0.5", ""})
	old := probeRobot
	probeRobot = func(esn, ip string) probeResult { return probeResult{online: true} }
	t.Cleanup(func() { probeRobot = old })
	robotStatuses(time.Now())
	waitFor(t, func() bool { return robotStatuses(time.Now())[0].State == "online" })
	setBots(t)
	if got := robotStatuses(time.Now()); len(got) != 0 {
		t.Fatalf("want no robots, got %+v", got)
	}
	statusCache.Lock()
	defer statusCache.Unlock()
	if len(statusCache.m) != 0 {
		t.Fatal("cache entry for a forgotten robot should be dropped")
	}
}

func TestRobotStatusAPI(t *testing.T) {
	setBots(t)
	rec := httptest.NewRecorder()
	if !handleRobotStatusAPI(rec, httptest.NewRequest("GET", "/api-sdk/robots_status", nil)) {
		t.Fatal("handler should claim the path")
	}
	var body struct {
		Robots []RobotStatus `json:"robots"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Robots == nil || len(body.Robots) != 0 {
		t.Fatalf("no robots should give an empty list, not an error: %v %s", err, rec.Body.String())
	}
	if handleRobotStatusAPI(httptest.NewRecorder(), httptest.NewRequest("GET", "/api-sdk/other", nil)) {
		t.Fatal("handler should ignore other paths")
	}
}

func TestBatteryInfo(t *testing.T) {
	cases := []struct {
		name      string
		volts     float32
		level     int
		charging  bool
		onCharger bool
		suggested float32
		pct, lvl  int
	}{
		{"full voltage", 4.2, 3, false, false, 0, 100, 3},
		{"mid", 3.85, 2, false, false, 0, 80, 2},
		{"empty", 3.4, 1, false, false, 0, 0, 0},
		{"unknown voltage assumes 70", 0, 3, false, false, 0, 70, 3},
		{"level 2 under 20 percent turns red", 3.5, 2, false, false, 0, 0, 0},
		{"level 1 off the charger is capped at 15", 3.8, 1, false, false, 0, 15, 0},
		{"finished charging reads 100", 3.9, 3, false, true, 0, 100, 3},
		{"still charging keeps the curve", 3.9, 3, true, true, 0, 0, 3},
	}
	for _, c := range cases {
		got := batteryInfo(c.volts, c.level, c.charging, c.onCharger, c.suggested)
		if c.name == "still charging keeps the curve" {
			if got.Percent <= 80 || got.Percent >= 100 || !got.Charging {
				t.Errorf("%s: got %+v", c.name, got)
			}
			continue
		}
		if got.Percent != c.pct || got.Level != c.lvl {
			t.Errorf("%s: got %d%% level %d, want %d%% level %d", c.name, got.Percent, got.Level, c.pct, c.lvl)
		}
	}
}

func TestEyeFromJdoc(t *testing.T) {
	if eyeFromJdoc("nobody") != nil {
		t.Fatal("unknown robot has no eye info")
	}
	vars.AddJdoc("vic:00a1", "vic.RobotSettings", vars.AJdoc{JsonDoc: `{"eye_color":4,"custom_eye_color":{"enabled":true,"hue":0.5,"saturation":0.75}}`})
	eye := eyeFromJdoc("00a1")
	if eye == nil || eye.Preset != 4 || !eye.Custom || eye.Hue != 0.5 || eye.Sat != 0.75 {
		t.Fatalf("got %+v", eye)
	}
	vars.AddJdoc("vic:00a1", "vic.RobotSettings", vars.AJdoc{JsonDoc: `{"eye_color":1,"custom_eye_color":{"enabled":false}}`})
	if eye := eyeFromJdoc("00a1"); eye == nil || eye.Preset != 1 || eye.Custom {
		t.Fatalf("custom disabled should read as preset, got %+v", eye)
	}
}

func TestRobotChatsAPI(t *testing.T) {
	setBots(t, [3]string{"00a1", "10.0.0.5", ""})
	old := vars.RememberedChats
	vars.RememberedChats = []vars.RememberedChat{
		{ESN: "00A1", Chats: []openai.ChatCompletionMessage{
			{Role: "system", Content: "secret prompt"},
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi there"},
			{Role: "user", Content: "  "},
		}},
		{ESN: "00b2", Chats: []openai.ChatCompletionMessage{{Role: "user", Content: "not yours"}}},
	}
	t.Cleanup(func() { vars.RememberedChats = old })

	rec := httptest.NewRecorder()
	handleRobotIdentityAPI(rec, httptest.NewRequest("GET", "/api-sdk/get_robot_chats?serial=00a1", nil))
	var lines []ChatLine
	if err := json.Unmarshal(rec.Body.Bytes(), &lines); err != nil {
		t.Fatal(err, rec.Body.String())
	}
	if len(lines) != 2 || lines[0] != (ChatLine{"user", "hello"}) || lines[1] != (ChatLine{"assistant", "hi there"}) {
		t.Fatalf("only this robot's user/assistant lines expected, got %+v", lines)
	}

	rec = httptest.NewRecorder()
	handleRobotIdentityAPI(rec, httptest.NewRequest("GET", "/api-sdk/get_robot_chats?serial=00ff", nil))
	if rec.Code != 404 {
		t.Fatalf("unknown robot should 404, got %d", rec.Code)
	}

	vars.RememberedChats = nil
	rec = httptest.NewRecorder()
	handleRobotIdentityAPI(rec, httptest.NewRequest("GET", "/api-sdk/get_robot_chats?serial=00a1", nil))
	if rec.Body.String() != "[]\n" {
		t.Fatalf("no chats should be an empty list, got %q", rec.Body.String())
	}
}

func TestDescribeRobotLeavesOutTheToken(t *testing.T) {
	info := describeRobot(Robot{ESN: "00a1", GUID: "SECRET-TOKEN", Target: "10.0.0.5:443", BcAssumption: true, ConnTimer: 12})
	b, _ := json.Marshal(info)
	if strings.Contains(string(b), "SECRET") || strings.Contains(string(b), ":443") {
		t.Fatalf("token or port leaked: %s", b)
	}
	if info.Address != "10.0.0.5" || !info.BehaviorControl || info.IdleSeconds != 12 {
		t.Fatalf("wrong info: %+v", info)
	}
}
