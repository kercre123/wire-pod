package wirepod_ttr

import (
	"strings"
	"testing"

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
