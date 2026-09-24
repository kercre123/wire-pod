package vars

import (
	"encoding/json"
	"testing"
)

// The fallback is the whole safety property: a call site that previously read
// bot.IPAddress directly must never end up with a worse value than before.
func TestGetRobotDialTargetFallsBackToObserved(t *testing.T) {
	withCleanConfig(t)

	for _, tc := range []struct {
		name string
		m    map[string]string
	}{
		{"nil map", nil},
		{"empty map", map[string]string{}},
		{"other robot only", map[string]string{"00e29999": "192.168.1.99"}},
		{"blank value", map[string]string{"00e20145": ""}},
		{"whitespace-only value", map[string]string{"00e20145": "   "}},
	} {
		APIConfig.Server.RobotLANAddresses = tc.m
		if got := GetRobotDialTarget("00e20145", "172.19.0.1"); got != "172.19.0.1" {
			t.Errorf("%s: GetRobotDialTarget = %q, want the observed address %q",
				tc.name, got, "172.19.0.1")
		}
	}
}

func TestGetRobotDialTargetReturnsConfiguredAddress(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.RobotLANAddresses = map[string]string{"00e20145": "192.0.2.13"}

	if got := GetRobotDialTarget("00e20145", "172.19.0.1"); got != "192.0.2.13" {
		t.Errorf("GetRobotDialTarget = %q, want the configured address %q", got, "192.0.2.13")
	}
}

// ESNs appear in mixed case across the codebase, so matching must not depend
// on how the admin happened to type it.
func TestGetRobotDialTargetMatchesESNCaseInsensitively(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.RobotLANAddresses = map[string]string{"00E20145": " 192.0.2.13 "}

	for _, esn := range []string{"00e20145", "00E20145", " 00e20145 "} {
		if got := GetRobotDialTarget(esn, "172.19.0.1"); got != "192.0.2.13" {
			t.Errorf("GetRobotDialTarget(%q) = %q, want %q", esn, got, "192.0.2.13")
		}
	}
}

func TestSetRobotLANAddressSetsAndClears(t *testing.T) {
	withCleanConfig(t)

	SetRobotLANAddress("00e20145", "192.0.2.13")
	if got := GetRobotDialTarget("00e20145", "obs"); got != "192.0.2.13" {
		t.Fatalf("after set, dial target = %q, want %q", got, "192.0.2.13")
	}

	SetRobotLANAddress("00e20145", "")
	if got := GetRobotDialTarget("00e20145", "obs"); got != "obs" {
		t.Errorf("after clear, dial target = %q, want the observed address", got)
	}
	if APIConfig.Server.RobotLANAddresses != nil {
		t.Error("clearing the only entry should leave a nil map so the field stays omitted")
	}
}

// One robot must never accumulate two entries that disagree.
func TestSetRobotLANAddressReplacesCaseVariantKey(t *testing.T) {
	withCleanConfig(t)

	SetRobotLANAddress("00E20145", "192.0.2.13")
	SetRobotLANAddress("00e20145", "192.0.2.99")

	if n := len(APIConfig.Server.RobotLANAddresses); n != 1 {
		t.Errorf("map holds %d entries %v, want 1", n, APIConfig.Server.RobotLANAddresses)
	}
	if got := GetRobotDialTarget("00e20145", "obs"); got != "192.0.2.99" {
		t.Errorf("dial target = %q, want the most recently set %q", got, "192.0.2.99")
	}
}

func TestSetRobotLANAddressIgnoresBlankESN(t *testing.T) {
	withCleanConfig(t)
	SetRobotLANAddress("   ", "192.0.2.13")
	if APIConfig.Server.RobotLANAddresses != nil {
		t.Errorf("a blank ESN created an entry: %v", APIConfig.Server.RobotLANAddresses)
	}
}

// Feature B must stay invisible to a default install.
func TestConfiguringDialTargetsKeepsDefaultConfigByteIdentical(t *testing.T) {
	withCleanConfig(t)

	SetRobotLANAddress("00e20145", "192.0.2.13")
	SetRobotLANAddress("00e20145", "")
	NormalizeConfig()

	raw, err := marshalConfig()
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	if raw != baselineZeroConfigJSON {
		t.Errorf("setting then clearing an address left the config changed.\n got: %s\nwant: %s",
			raw, baselineZeroConfigJSON)
	}
}

func marshalConfig() (string, error) {
	b, err := json.Marshal(APIConfig)
	return string(b), err
}
