package wirepod_ttr

import (
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/memories"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func TestCreateAIReqRobotIdentity(t *testing.T) {
	oldBots := vars.BotInfo
	defer func() { vars.BotInfo = oldBots }()
	vars.BotInfo.Robots = nil
	vars.BotInfo.Robots = append(vars.BotInfo.Robots, struct {
		Esn         string `json:"esn"`
		IPAddress   string `json:"ip_address"`
		GUID        string `json:"guid"`
		Activated   bool   `json:"activated"`
		Name        string `json:"name,omitempty"`
		Personality string `json:"personality,omitempty"`
	}{Esn: "00902de7", Name: "Orbit", Personality: "grumpy but kind"})

	sys := func(esn string) string {
		return CreateAIReq("hello", esn, false, false).Messages[0].Content
	}

	got := sys("00902de7")
	if !strings.HasPrefix(got, "Your name is Orbit.") || !strings.Contains(got, "Your personality and identity: grumpy but kind") {
		t.Fatalf("identity missing from system prompt: %q", got)
	}
	if !strings.Contains(got, "called Vector") {
		t.Fatalf("global prompt should follow the identity: %q", got)
	}
	if other := sys("deadbeef"); strings.Contains(other, "Orbit") || strings.Contains(other, "grumpy") {
		t.Fatalf("identity leaked to another robot: %q", other)
	}
	vars.BotInfo.Robots[0].Name, vars.BotInfo.Robots[0].Personality = "", ""
	if cleared := sys("00902de7"); strings.Contains(cleared, "Your name is") || strings.Contains(cleared, "personality") {
		t.Fatalf("cleared identity still present: %q", cleared)
	}
}

func TestMemoriesStayPerRobotInPrompt(t *testing.T) {
	memories.DirOverride = t.TempDir()
	defer func() { memories.DirOverride = "" }()
	if _, _, err := memories.Add("aaaa1111", "Lucas's dog is named Biscuit"); err != nil {
		t.Fatal(err)
	}
	a := CreateAIReq("hello", "aaaa1111", false, false).Messages[0].Content
	b := CreateAIReq("hello", "bbbb2222", false, false).Messages[0].Content
	if !strings.Contains(a, "Things you remember:\n- Lucas's dog is named Biscuit") {
		t.Fatalf("robot A prompt missing its memory: %q", a)
	}
	if strings.Contains(b, "Biscuit") || strings.Contains(b, "Things you remember") {
		t.Fatalf("robot B prompt contains robot A memory: %q", b)
	}
}

func TestRememberFactAction(t *testing.T) {
	acts := GetActionsFromString("Okay! {{rememberFact||Lucas likes tea}} Noted.")
	var found bool
	for _, a := range acts {
		if a.Action == ActionRememberFact && a.Parameter == "Lucas likes tea" {
			found = true
		}
		if a.Action == ActionSayText && strings.Contains(a.Parameter, "rememberFact") {
			t.Fatalf("command leaked into speech: %v", acts)
		}
	}
	if !found {
		t.Fatalf("rememberFact action not parsed: %v", acts)
	}
}
