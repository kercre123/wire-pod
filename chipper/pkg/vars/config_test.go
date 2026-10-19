package vars

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Baseline shape of apiConfig as marshalled with every field at its zero value,
// captured at commit 1f33b63 before any proxy-networking work.
//
// This is the I4 guarantee in its strongest form: a default install's
// apiConfig.json must stay byte-identical. Any new field that is NOT tagged
// omitempty will break this test, which is the entire point — ReadConfig
// rewrites this file on every startup (config.go:132-133), so a new key here
// silently rewrites the config of every existing install on first boot.
const baselineZeroConfigJSON = `{"weather":{"enable":false,"provider":"","key":"","unit":""},` +
	`"knowledge":{"enable":false,"provider":"","key":"","id":"","model":"","intentgraph":false,` +
	`"robotName":"","openai_prompt":"","openai_voice":"","openai_voice_with_english":false,` +
	`"save_chat":false,"commands_enable":false,"endpoint":"","top_p":0,"temp":0},` +
	`"STT":{"provider":"","language":""},` +
	`"server":{"epconfig":false,"port":""},` +
	`"hasreadfromenv":false,"pastinitialsetup":false}`

func TestZeroConfigMarshalsToBaselineBytes(t *testing.T) {
	var cfg apiConfig
	got, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshalling zero apiConfig: %v", err)
	}
	if string(got) != baselineZeroConfigJSON {
		t.Errorf("zero-valued apiConfig no longer marshals to the baseline bytes.\n"+
			"A new field is missing `omitempty`, or an existing field changed.\n"+
			" got: %s\nwant: %s", got, baselineZeroConfigJSON)
	}
}

// A baseline config file must survive read-then-write untouched. ReadConfig
// unconditionally rewrites the file, so this is the round-trip an existing
// install actually experiences on upgrade.
func TestBaselineConfigRoundTripsUnchanged(t *testing.T) {
	original := []byte(baselineZeroConfigJSON)

	var cfg apiConfig
	if err := json.Unmarshal(original, &cfg); err != nil {
		t.Fatalf("unmarshalling baseline config: %v", err)
	}
	got, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("re-marshalling baseline config: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("baseline apiConfig.json did not round-trip.\n got: %s\nwant: %s", got, original)
	}
}

// A populated config must also round-trip, so the guarantee is not an artifact
// of everything being zero.
func TestPopulatedConfigRoundTripsUnchanged(t *testing.T) {
	original := `{"weather":{"enable":true,"provider":"openweathermap","key":"abc","unit":"F"},` +
		`"knowledge":{"enable":true,"provider":"openai","key":"redacted","id":"","model":"gpt-4",` +
		`"intentgraph":false,"robotName":"Vector","openai_prompt":"hi","openai_voice":"nova",` +
		`"openai_voice_with_english":true,"save_chat":true,"commands_enable":false,` +
		`"endpoint":"https://api.openai.com/v1","top_p":0.9,"temp":0.7},` +
		`"STT":{"provider":"vosk","language":"en-US"},` +
		`"server":{"epconfig":true,"port":"443"},` +
		`"hasreadfromenv":true,"pastinitialsetup":true}`

	var cfg apiConfig
	if err := json.Unmarshal([]byte(original), &cfg); err != nil {
		t.Fatalf("unmarshalling populated config: %v", err)
	}
	got, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("re-marshalling populated config: %v", err)
	}
	if string(got) != original {
		t.Errorf("populated apiConfig.json did not round-trip.\n got: %s\nwant: %s", got, original)
	}
}

// ReadConfig is the real upgrade path: it reads, then unconditionally rewrites.
// An existing install's file must come back out identical.
func TestReadConfigDoesNotRewriteBaselineFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "apiConfig.json")

	// Port must match DDL_RPC_PORT or ReadConfig flips HasReadFromEnv /
	// PastInitialSetup (config.go:120-125). Pin both so the only thing under
	// test is field-shape stability.
	t.Setenv("DDL_RPC_PORT", "443")
	t.Setenv("STT_SERVICE", "vosk")

	original := `{"weather":{"enable":false,"provider":"","key":"","unit":""},` +
		`"knowledge":{"enable":false,"provider":"","key":"","id":"","model":"","intentgraph":false,` +
		`"robotName":"","openai_prompt":"","openai_voice":"","openai_voice_with_english":false,` +
		`"save_chat":false,"commands_enable":false,"endpoint":"","top_p":0,"temp":0},` +
		`"STT":{"provider":"vosk","language":"en-US"},` +
		`"server":{"epconfig":true,"port":"443"},` +
		`"hasreadfromenv":true,"pastinitialsetup":true}`

	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("seeding config file: %v", err)
	}

	restorePath, restoreCfg := ApiConfigPath, APIConfig
	t.Cleanup(func() { ApiConfigPath, APIConfig = restorePath, restoreCfg })
	ApiConfigPath = path
	APIConfig = apiConfig{}

	ReadConfig()

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config back: %v", err)
	}
	if string(after) != original {
		t.Errorf("ReadConfig rewrote an existing install's apiConfig.json.\n"+
			"before: %s\n after: %s", original, after)
	}
}
