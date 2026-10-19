package botsetup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// withCapturedConfig points the config path at a temp file and clears APIConfig,
// restoring both afterwards. Package-level state in vars, so no t.Parallel.
func withCapturedConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "apiConfig.json")
	restorePath, restoreCfg := vars.ApiConfigPath, vars.APIConfig
	t.Cleanup(func() { vars.ApiConfigPath, vars.APIConfig = restorePath, restoreCfg })
	vars.ApiConfigPath = path
	vars.APIConfig.Server.RobotLANAddresses = nil
	vars.APIConfig.Server.ProxyMode = true
	return path
}

// configuredAddress reads back what was persisted, so the assertion covers the
// write rather than just the in-memory map.
func configuredAddress(t *testing.T, path, esn string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg struct {
		Server struct {
			RobotLANAddresses map[string]string `json:"robot_lan_addresses"`
		} `json:"server"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parsing written config: %v", err)
	}
	return cfg.Server.RobotLANAddresses[esn]
}

// The happy path: the robot names itself and its address over BLE, and both
// end up on disk so wire-pod can dial it after a restart.
func TestRecordsTheAddressTheRobotReported(t *testing.T) {
	path := withCapturedConfig(t)

	if !RecordRobotLANAddress("00e20145", "192.0.2.13") {
		t.Fatal("a valid ESN and address should have been recorded")
	}
	if got := vars.GetRobotDialTarget("00e20145", "172.28.0.1"); got != "192.0.2.13" {
		t.Errorf("dial target = %q, want the captured address", got)
	}
	if got := configuredAddress(t, path, "00e20145"); got != "192.0.2.13" {
		t.Errorf("persisted address = %q, want %q; a capture that is not written "+
			"is lost on restart", got, "192.0.2.13")
	}
}

// The address arrives from the robot over BLE, so it gets the same validation
// as an admin-typed one. A bad value would be concatenated with ":" + port at
// dial time, or point wire-pod at something that is not the robot.
func TestRefusesAnUnusableReportedAddress(t *testing.T) {
	for _, tc := range []struct{ addr, why string }{
		{"", "no address reported at all"},
		{"   ", "whitespace only"},
		{"192.0.2.13:443", "host:port would be concatenated into host:port:port"},
		{"http://192.0.2.13", "a URL is not a bare host"},
		{"fe80::1", "IPv6 cannot be concatenated with :port"},
	} {
		path := withCapturedConfig(t)

		if RecordRobotLANAddress("00e20145", tc.addr) {
			t.Errorf("recorded %q (%s); it would reach the dialler unusable", tc.addr, tc.why)
		}
		if got := configuredAddress(t, path, "00e20145"); got != "" {
			t.Errorf("persisted %q for a rejected address %q", got, tc.addr)
		}
		if got := vars.GetRobotDialTarget("00e20145", "172.28.0.1"); got != "172.28.0.1" {
			t.Errorf("dial target = %q, want the observed address to stand when "+
				"the reported one is rejected", got)
		}
	}
}

// A robot that reports no ESN cannot be keyed, and inventing one would attach
// the address to the wrong robot.
func TestRefusesAnEmptyESN(t *testing.T) {
	path := withCapturedConfig(t)

	for _, esn := range []string{"", "   "} {
		if RecordRobotLANAddress(esn, "192.0.2.13") {
			t.Errorf("recorded an address against ESN %q", esn)
		}
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("nothing should have been written for an unidentifiable robot")
	}
}

// Re-adopting a robot that has moved must replace its address, not accumulate
// entries: lookups take the first case-insensitive match, so a stale duplicate
// could win.
func TestReAdoptingReplacesTheStoredAddress(t *testing.T) {
	withCapturedConfig(t)

	RecordRobotLANAddress("00e20145", "192.0.2.13")
	RecordRobotLANAddress("00E20145", "192.0.2.77")

	if n := len(vars.APIConfig.Server.RobotLANAddresses); n != 1 {
		t.Errorf("stored %d entries for one robot, want 1: %v",
			n, vars.APIConfig.Server.RobotLANAddresses)
	}
	if got := vars.GetRobotDialTarget("00e20145", "172.28.0.1"); got != "192.0.2.77" {
		t.Errorf("dial target = %q, want the address from the most recent adoption", got)
	}
}

// Whitespace around either value is the robot's, not the admin's, and must not
// become part of the key or the dial target.
func TestTrimsWhatTheRobotReported(t *testing.T) {
	withCapturedConfig(t)

	if !RecordRobotLANAddress("  00e20145  ", "  192.0.2.13  ") {
		t.Fatal("padded but valid values should be recorded")
	}
	if got := vars.GetRobotDialTarget("00e20145", "172.28.0.1"); got != "192.0.2.13" {
		t.Errorf("dial target = %q, want the trimmed address", got)
	}
}
