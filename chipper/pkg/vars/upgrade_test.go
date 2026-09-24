package vars

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withConfigFile points ApiConfigPath at a temp file seeded with contents and
// restores both the path and APIConfig afterwards. Package-level state, so
// nothing here may run in parallel.
func withConfigFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "apiConfig.json")
	if err := os.WriteFile(path, []byte(contents), 0666); err != nil {
		t.Fatalf("seeding config file: %v", err)
	}
	restorePath, restoreCfg := ApiConfigPath, APIConfig
	t.Cleanup(func() { ApiConfigPath, APIConfig = restorePath, restoreCfg })
	ApiConfigPath, APIConfig = path, apiConfig{}
	return path
}

// The on-disk form of the three new fields, pinned. The unset case is covered
// several times over; this is the other half. Without it a renamed json tag
// would silently orphan every proxy install's settings on upgrade and no test
// would notice, because the byte-identity tests only ever see them omitted.
func TestConfiguredNetworkingFieldsMarshalToTheirDocumentedKeys(t *testing.T) {
	withCleanConfig(t)
	APIConfig.Server.ProxyMode = true
	APIConfig.Server.RobotFacingAddress = "wirepod.lan"
	APIConfig.Server.RobotLANAddresses = map[string]string{"00e20145": "192.0.2.13"}
	NormalizeConfig()

	raw, err := json.Marshal(APIConfig)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	var probe struct {
		Server map[string]json.RawMessage `json:"server"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}

	for key, want := range map[string]string{
		"proxy_mode":           `true`,
		"robot_facing_address": `"wirepod.lan"`,
		"robot_lan_addresses":  `{"00e20145":"192.0.2.13"}`,
	} {
		got, ok := probe.Server[key]
		if !ok {
			t.Errorf("server.%s is missing; a configured install would lose this setting", key)
			continue
		}
		if string(got) != want {
			t.Errorf("server.%s = %s, want %s", key, got, want)
		}
	}
}

// The round trip an existing install actually performs when an admin sets an
// address and then clears it again: the file must gain exactly one key and then
// return to its original bytes. WriteConfigToDisk is the only persistence path
// for all of this and is otherwise untested.
func TestSettingAndClearingALANAddressLeavesTheFileAsItWas(t *testing.T) {
	path := withConfigFile(t, baselineZeroConfigJSON)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading seed: %v", err)
	}

	if err := json.Unmarshal(original, &APIConfig); err != nil {
		t.Fatalf("unmarshalling seed: %v", err)
	}

	SetRobotLANAddress("00e20145", "192.0.2.13")
	WriteConfigToDisk()

	withAddress, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading after set: %v", err)
	}
	if string(withAddress) == string(original) {
		t.Fatal("setting a LAN address did not change the file at all")
	}
	var probe struct {
		Server map[string]json.RawMessage `json:"server"`
	}
	if err := json.Unmarshal(withAddress, &probe); err != nil {
		t.Fatalf("unmarshalling after set: %v", err)
	}
	if len(probe.Server) != 3 {
		t.Errorf("server block has %d keys after setting one address, want 3 "+
			"(epconfig, port, robot_lan_addresses); got %v", len(probe.Server), probe.Server)
	}

	SetRobotLANAddress("00e20145", "")
	WriteConfigToDisk()

	cleared, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading after clear: %v", err)
	}
	if string(cleared) != string(original) {
		t.Errorf("clearing the only LAN address did not restore the original bytes.\n"+
			" got: %s\nwant: %s", cleared, original)
	}
}

// Configuring one robot must not disturb another. Every other dial-target test
// uses a single-entry map, so nothing else would catch a setter that rebuilt
// the map instead of editing it.
func TestSettingOneRobotLeavesTheOthersAlone(t *testing.T) {
	withCleanConfig(t)
	SetRobotLANAddress("00e20145", "192.0.2.13")
	SetRobotLANAddress("00e20146", "192.0.2.14")

	SetRobotLANAddress("00e20145", "192.0.2.99")

	for _, tc := range []struct{ esn, want string }{
		{"00e20145", "192.0.2.99"},
		{"00e20146", "192.0.2.14"},
	} {
		if got := GetRobotDialTarget(tc.esn, "observed"); got != tc.want {
			t.Errorf("GetRobotDialTarget(%q) = %q, want %q; editing one robot "+
				"disturbed another", tc.esn, got, tc.want)
		}
	}
}

// Lookups are case-insensitive, so two keys differing only in case would
// otherwise resolve by map iteration order: the UI could show one address while
// wire-pod dialled the other, and the winner could change on restart. A
// hand-edited config is the only way to produce this, which is exactly why the
// setter's own de-duplication does not cover it.
func TestNormalizeCollapsesCaseVariantESNKeysDeterministically(t *testing.T) {
	const seed = `{"server":{"robot_lan_addresses":` +
		`{"00E20145":"192.0.2.13","00e20145":"192.0.2.99"}}}`

	first := ""
	for attempt := 0; attempt < 20; attempt++ {
		withCleanConfig(t)
		if err := json.Unmarshal([]byte(seed), &APIConfig); err != nil {
			t.Fatalf("unmarshalling seed: %v", err)
		}
		NormalizeConfig()

		if n := len(APIConfig.Server.RobotLANAddresses); n != 1 {
			t.Fatalf("case-variant keys collapsed to %d entries, want 1: %v",
				n, APIConfig.Server.RobotLANAddresses)
		}
		got := GetRobotDialTarget("00e20145", "observed")
		if first == "" {
			first = got
		}
		if got != first {
			t.Fatalf("dial target changed between loads of an identical config: "+
				"%q then %q; the survivor must not depend on map iteration order", first, got)
		}
	}
}

// The setters validate, but a hand-edited apiConfig.json does not go through
// them. CreateServerConfig concatenates ":" + port onto whatever it finds, so a
// stored "host:443" would become "host:443:443" and produce a robot that cannot
// connect. Clearing it falls back to the outbound IP: wrong, but reachable.
func TestNormalizeClearsAnUnusableRobotFacingAddressFromDisk(t *testing.T) {
	for _, tc := range []struct{ stored, why string }{
		{"192.168.1.10:443", "host:port would be concatenated into host:port:port"},
		{"https://wirepod.lan", "a URL is not a bare host"},
		{"wirepod.lan/ok", "a path is not a bare host"},
		{"fe80::1", "IPv6 cannot be concatenated with :port"},
	} {
		withCleanConfig(t)
		APIConfig.Server.RobotFacingAddress = tc.stored
		NormalizeConfig()

		if APIConfig.Server.RobotFacingAddress != "" {
			t.Errorf("a stored %q survived normalization (%s); it would reach "+
				"server_config.json unusable", tc.stored, tc.why)
		}
	}
}

// The counterpart: a legitimate stored value must not be discarded.
func TestNormalizeKeepsAUsableRobotFacingAddressFromDisk(t *testing.T) {
	for _, stored := range []string{"wirepod.lan", "escapepod.local", "192.0.2.12"} {
		withCleanConfig(t)
		APIConfig.Server.RobotFacingAddress = stored
		NormalizeConfig()

		if APIConfig.Server.RobotFacingAddress != stored {
			t.Errorf("normalization discarded the valid stored address %q (got %q)",
				stored, APIConfig.Server.RobotFacingAddress)
		}
	}
}

// A realistic pre-fork config: past setup, port and STT matching the env below,
// so ReadConfig's env reconciliation does not mutate anything and the only
// variable under test is the networking field.
const preforkConfigJSON = `{"weather":{"enable":false,"provider":"","key":"","unit":""},` +
	`"knowledge":{"enable":false,"provider":"","key":"","id":"","model":"","intentgraph":false,` +
	`"robotName":"","openai_prompt":"","openai_voice":"","openai_voice_with_english":false,` +
	`"save_chat":false,"commands_enable":false,"endpoint":"","top_p":0,"temp":0},` +
	`"STT":{"provider":"vosk","language":"en-US"},` +
	`"server":{"epconfig":true,"port":"443"},` +
	`"hasreadfromenv":true,"pastinitialsetup":true}`

// The path a restart actually takes: an address is set through the setter, the
// config is written, the process comes back with nothing in memory, ReadConfig
// loads from disk, and the dial target must still be the configured address.
//
// The write and the read are each covered above but never joined, and a robot
// only stays reachable across a restart if the whole path holds. A serialiser
// that wrote the field under the wrong key, or a loader that dropped it, would
// satisfy every individual test here and still lose the setting on reboot.
func TestAConfiguredAddressSurvivesAWriteAndReload(t *testing.T) {
	t.Setenv("DDL_RPC_PORT", "443")
	t.Setenv("STT_SERVICE", "vosk")
	withConfigFile(t, preforkConfigJSON)

	ReadConfig()
	if got := GetRobotDialTarget("00e20145", "172.28.0.1"); got != "172.28.0.1" {
		t.Fatalf("precondition: an unconfigured robot should dial its observed "+
			"address, got %q", got)
	}

	SetRobotLANAddress("00e20145", "192.0.2.13")
	WriteConfigToDisk()

	// The restart: nothing survives in memory.
	APIConfig = apiConfig{}
	ReadConfig()

	if got := GetRobotDialTarget("00e20145", "172.28.0.1"); got != "192.0.2.13" {
		t.Errorf("after write and reload the dial target is %q, want %q. The "+
			"address was lost somewhere between the setter and the loader, so a "+
			"robot behind a proxy becomes unreachable on the next restart", got, "192.0.2.13")
	}
}

// Clearing must survive a reload too, or an address an admin removed comes back
// from disk and wire-pod keeps dialling somewhere the robot no longer is.
func TestAClearedAddressStaysClearedAcrossAReload(t *testing.T) {
	t.Setenv("DDL_RPC_PORT", "443")
	t.Setenv("STT_SERVICE", "vosk")
	withConfigFile(t, preforkConfigJSON)

	ReadConfig()
	SetRobotLANAddress("00e20145", "192.0.2.13")
	WriteConfigToDisk()

	SetRobotLANAddress("00e20145", "")
	WriteConfigToDisk()

	APIConfig = apiConfig{}
	ReadConfig()

	if got := GetRobotDialTarget("00e20145", "172.28.0.1"); got != "172.28.0.1" {
		t.Errorf("after clearing and reloading the dial target is %q, want the "+
			"observed address %q", got, "172.28.0.1")
	}
}

// WriteConfigToDisk normalises before marshalling, and that is load-bearing
// rather than belt-and-braces. Other paths legitimately mutate APIConfig
// without going through SetRobotLANAddress -- the first-run wizard, the config
// handlers, a future caller -- and whatever reaches disk is what the next
// process reads back and dials.
//
// Found by mutation: deleting the NormalizeConfig call from WriteConfigToDisk
// previously broke no test at all, because the only covered write path happened
// to normalise itself on the way in.
func TestWriteNormalisesWhatReachesDisk(t *testing.T) {
	path := withConfigFile(t, preforkConfigJSON)

	// Assign directly, as code outside the setter does.
	APIConfig.Server.RobotFacingAddress = "  wirepod.lan  "
	APIConfig.Server.RobotLANAddresses = map[string]string{"00E20145": "  192.0.2.13  "}
	WriteConfigToDisk()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}
	var cfg struct {
		Server struct {
			RobotFacingAddress string            `json:"robot_facing_address"`
			RobotLANAddresses  map[string]string `json:"robot_lan_addresses"`
		} `json:"server"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parsing written config: %v", err)
	}

	if cfg.Server.RobotFacingAddress != "wirepod.lan" {
		t.Errorf("robot_facing_address written as %q, want it trimmed to %q; "+
			"CreateServerConfig concatenates \":\"+port onto this",
			cfg.Server.RobotFacingAddress, "wirepod.lan")
	}
	if got, ok := cfg.Server.RobotLANAddresses["00e20145"]; !ok || got != "192.0.2.13" {
		t.Errorf("robot_lan_addresses written as %v, want the key lowercased and "+
			"the value trimmed; lookups are case-insensitive, so a stored variant "+
			"resolves by map iteration order", cfg.Server.RobotLANAddresses)
	}
}

// An emptied map must be omitted, not written as {}. omitempty elides a nil map
// but not an empty non-nil one, so without the normalise-on-write a default
// install that set and removed an address would no longer match upstream byte
// for byte -- the guarantee the whole feature rests on.
func TestAnEmptiedMapIsOmittedFromDisk(t *testing.T) {
	path := withConfigFile(t, preforkConfigJSON)

	APIConfig.Server.RobotLANAddresses = map[string]string{}
	WriteConfigToDisk()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}
	if strings.Contains(string(raw), "robot_lan_addresses") {
		t.Errorf("an empty map was written to disk as a key; a default install's "+
			"config must stay byte-identical to upstream.\n  got: %s", raw)
	}
}
