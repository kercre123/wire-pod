package vars

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func withCleanConfig(t *testing.T) {
	t.Helper()
	restore := APIConfig
	t.Cleanup(func() { APIConfig = restore })
	APIConfig = apiConfig{}
}

// The three new fields must vanish entirely from a default install's config.
// This is the same guarantee as TestZeroConfigMarshalsToBaselineBytes, stated
// against the fields directly so a failure points at the cause.
func TestNewServerFieldsAreOmittedWhenUnset(t *testing.T) {
	withCleanConfig(t)
	NormalizeConfig()

	raw, err := json.Marshal(APIConfig)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	var decoded struct {
		Server map[string]json.RawMessage `json:"server"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decoding server block: %v", err)
	}

	for _, key := range []string{"proxy_mode", "robot_facing_address", "robot_lan_addresses"} {
		if _, present := decoded.Server[key]; present {
			t.Errorf("%q appears in a default config; it must be omitempty", key)
		}
	}
	if len(decoded.Server) != 2 {
		t.Errorf("server block has keys %v, want exactly epconfig and port", decoded.Server)
	}
}

func TestNormalizeTrimsRobotFacingAddress(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.RobotFacingAddress = "  escapepod.local\t"
	NormalizeConfig()

	if got := APIConfig.Server.RobotFacingAddress; got != "escapepod.local" {
		t.Errorf("RobotFacingAddress = %q, want %q", got, "escapepod.local")
	}
}

func TestNormalizeTrimsMapEntriesAndDropsBlanks(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.RobotLANAddresses = map[string]string{
		"  00e20145 ": " 192.168.1.150 ",
		"00e20146":    "   ",
		"   ":         "192.168.1.152",
		"00e20147":    "192.168.1.153",
	}
	NormalizeConfig()

	got := APIConfig.Server.RobotLANAddresses
	want := map[string]string{"00e20145": "192.168.1.150", "00e20147": "192.168.1.153"}

	if len(got) != len(want) {
		t.Fatalf("normalized map = %v, want %v", got, want)
	}
	for esn, addr := range want {
		if got[esn] != addr {
			t.Errorf("entry %q = %q, want %q", esn, got[esn], addr)
		}
	}
}

// An allocated-but-empty map must be reduced to nil, otherwise it still
// marshals as "robot_lan_addresses":{} and breaks byte-identity.
func TestNormalizeReducesEmptyMapToNil(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.RobotLANAddresses = map[string]string{}
	NormalizeConfig()
	if APIConfig.Server.RobotLANAddresses != nil {
		t.Error("empty map was not reduced to nil")
	}

	APIConfig.Server.RobotLANAddresses = map[string]string{"00e20145": "  "}
	NormalizeConfig()
	if APIConfig.Server.RobotLANAddresses != nil {
		t.Error("map holding only blank values was not reduced to nil")
	}

	raw, err := json.Marshal(APIConfig)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	if string(raw) != baselineZeroConfigJSON {
		t.Errorf("a config whose only content was blank entries is not byte-identical "+
			"to a default one.\n got: %s\nwant: %s", raw, baselineZeroConfigJSON)
	}
}

func TestValidateRobotFacingAddressAccepts(t *testing.T) {
	for _, addr := range []string{
		"",
		"escapepod.local",
		"wirepod.lan",
		"wire-pod.example.com",
		"192.0.2.12",
		"10.0.0.1",
		"a",
		"xn--80ak6aa92e.com",
	} {
		if err := ValidateRobotFacingAddress(addr); err != nil {
			t.Errorf("ValidateRobotFacingAddress(%q) rejected a valid address: %v", addr, err)
		}
	}
}

func TestValidateRobotFacingAddressRejects(t *testing.T) {
	for _, tc := range []struct{ addr, why string }{
		{"http://escapepod.local", "scheme"},
		{"https://192.0.2.12", "scheme"},
		{"escapepod.local:443", "embedded port — would produce host:443:443"},
		{"192.0.2.12:443", "embedded port"},
		{"escapepod.local/ok", "path"},
		{"192.0.2.12/ok", "path"},
		{"escape pod.local", "whitespace"},
		{"escapepod.local\n", "trailing newline"},
		{"user@escapepod.local", "userinfo"},
		{"-badlabel.local", "label starts with a hyphen"},
		{"badlabel-.local", "label ends with a hyphen"},
		{"double..dot", "empty label"},
		{"under_score.local", "invalid character for a DNS SAN"},
		{"fe80::1", "IPv6 has no bracket syntax in server_config.json"},
	} {
		if err := ValidateRobotFacingAddress(tc.addr); err == nil {
			t.Errorf("ValidateRobotFacingAddress(%q) accepted an invalid address (%s)",
				tc.addr, tc.why)
		}
	}
}

// Normalization must survive the read path, including the error branches,
// so no un-normalized value can reach an accessor.
func TestReadConfigNormalizesOnUnmarshalFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "apiConfig.json")
	if err := os.WriteFile(path, []byte("{ this is not json"), 0644); err != nil {
		t.Fatalf("seeding bad config: %v", err)
	}

	restorePath, restoreCfg := ApiConfigPath, APIConfig
	t.Cleanup(func() { ApiConfigPath, APIConfig = restorePath, restoreCfg })
	ApiConfigPath = path
	APIConfig = apiConfig{}
	APIConfig.Server.RobotFacingAddress = "  spaced.local  "
	APIConfig.Server.RobotLANAddresses = map[string]string{}

	ReadConfig()

	if got := APIConfig.Server.RobotFacingAddress; got != "spaced.local" {
		t.Errorf("address not normalized on the unmarshal-error path: %q", got)
	}
	if APIConfig.Server.RobotLANAddresses != nil {
		t.Error("empty map not reduced to nil on the unmarshal-error path")
	}
}
