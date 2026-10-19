package botsetup

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// redirectCertPaths points the package's write targets at a temp dir and
// restores them afterwards. These are package-level vars in vars, so tests in
// this file must not run in parallel.
func redirectCertPaths(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	certs, certPath, keyPath, serverCfg := vars.Certs, vars.CertPath, vars.KeyPath, vars.ServerConfigPath
	cfg := vars.APIConfig
	t.Cleanup(func() {
		vars.Certs, vars.CertPath, vars.KeyPath, vars.ServerConfigPath = certs, certPath, keyPath, serverCfg
		vars.APIConfig = cfg
	})

	vars.Certs = dir
	vars.CertPath = filepath.Join(dir, "cert.crt")
	vars.KeyPath = filepath.Join(dir, "cert.key")
	vars.ServerConfigPath = filepath.Join(dir, "server_config.json")
	return dir
}

func readServerConfig(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(vars.ServerConfigPath)
	if err != nil {
		t.Fatalf("reading server_config.json: %v", err)
	}
	return string(b)
}

// I4: with no proxy configuration, the non-EP server config must be exactly
// what the outbound-IP formula produces today. After Phase 5 this same
// assertion holds because GetRobotFacingAddress falls back to GetOutboundIP.
func TestCreateServerConfigIPModeMatchesOutboundIP(t *testing.T) {
	redirectCertPaths(t)
	vars.APIConfig.Server.EPConfig = false
	vars.APIConfig.Server.Port = "443"

	CreateServerConfig()

	ip := vars.GetOutboundIP().String()
	want := `{"jdocs":"` + ip + `:443","tms":"` + ip + `:443","chipper":"` + ip + `:443",` +
		`"check":"` + ip + `/ok","logfiles":"s3://anki-device-logs-prod/victor",` +
		`"appkey":"oDoa0quieSeir6goowai7f"}`

	if got := readServerConfig(t); got != want {
		t.Errorf("IP-mode server_config.json changed.\n got: %s\nwant: %s", got, want)
	}
}

// EP mode is explicitly out of scope for the robot-facing address, so its
// output is fully static and must never move.
func TestCreateServerConfigEPModeIsUnchanged(t *testing.T) {
	redirectCertPaths(t)
	vars.APIConfig.Server.EPConfig = true
	vars.APIConfig.Server.Port = "443"

	CreateServerConfig()

	const want = `{"jdocs":"escapepod.local:443","tms":"escapepod.local:443",` +
		`"chipper":"escapepod.local:443","check":"escapepod.local/ok",` +
		`"logfiles":"s3://anki-device-logs-prod/victor","appkey":"oDoa0quieSeir6goowai7f"}`

	if got := readServerConfig(t); got != want {
		t.Errorf("EP-mode server_config.json changed; EP mode must stay out of scope.\n"+
			" got: %s\nwant: %s", got, want)
	}
}

// A custom port must land exactly once. CreateServerConfig concatenates
// addr + ":" + port, which is why the robot-facing address needs validation
// against embedded ports.
func TestCreateServerConfigHonorsCustomPort(t *testing.T) {
	redirectCertPaths(t)
	vars.APIConfig.Server.EPConfig = false
	vars.APIConfig.Server.Port = "8443"

	CreateServerConfig()

	ip := vars.GetOutboundIP().String()
	var cfg ClientServerConfig
	if err := json.Unmarshal([]byte(readServerConfig(t)), &cfg); err != nil {
		t.Fatalf("parsing server config: %v", err)
	}
	if want := ip + ":8443"; cfg.Jdocs != want {
		t.Errorf("jdocs = %q, want %q", cfg.Jdocs, want)
	}
	if want := ip + "/ok"; cfg.Check != want {
		t.Errorf("check = %q, want %q (the check endpoint carries no port)", cfg.Check, want)
	}
}

func parseGeneratedCert(t *testing.T) *x509.Certificate {
	t.Helper()
	pemBytes, err := os.ReadFile(vars.CertPath)
	if err != nil {
		t.Fatalf("reading generated cert: %v", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("generated cert is not valid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parsing generated cert: %v", err)
	}
	return cert
}

// I4 for the certificate: with no robot-facing address configured, the SAN set
// is exactly one IP — the outbound IP — and no DNS names at all.
//
// This is the assertion that matters most in the whole suite. The robot pins
// this leaf into its root pool (vector-cloud/cloud/main.go:84), so the SAN set
// is the entire name policy, and changing it de-pairs every robot (I3).
func TestCreateCertComboBaselineSANs(t *testing.T) {
	redirectCertPaths(t)

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	cert := parseGeneratedCert(t)

	if len(cert.DNSNames) != 0 {
		t.Errorf("baseline cert gained DNS SANs %v; today it has none", cert.DNSNames)
	}
	if len(cert.IPAddresses) != 1 {
		t.Fatalf("baseline cert has %d IP SANs %v, want exactly 1",
			len(cert.IPAddresses), cert.IPAddresses)
	}
	if want := vars.GetOutboundIP(); !cert.IPAddresses[0].Equal(want) {
		t.Errorf("cert IP SAN = %v, want the outbound IP %v", cert.IPAddresses[0], want)
	}
}

// The leaf is what gets pinned, so it must be the thing written to CertPath,
// and it must be usable as a self-standing trust anchor the way the robot
// uses it. Guards the I3 reasoning against a well-meaning refactor that starts
// writing the CA instead.
func TestCreateCertComboWritesTheLeafNotTheCA(t *testing.T) {
	redirectCertPaths(t)

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	cert := parseGeneratedCert(t)

	if cert.IsCA {
		t.Error("CertPath now holds a CA certificate; the robot pins whatever is " +
			"here as a trust anchor, so this changes the pairing model")
	}
	if len(cert.IPAddresses) == 0 && len(cert.DNSNames) == 0 {
		t.Error("written certificate has no SANs at all; no robot could verify it")
	}
}

// Regenerating mints a fresh key every time, which is exactly why cert
// generation must never be added to the SSH path or a settings write (I3).
func TestCreateCertComboRegeneratesADifferentKey(t *testing.T) {
	redirectCertPaths(t)

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("first CreateCertCombo: %v", err)
	}
	first := parseGeneratedCert(t)

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("second CreateCertCombo: %v", err)
	}
	second := parseGeneratedCert(t)

	if first.Equal(second) {
		t.Error("two CreateCertCombo calls produced the same certificate; the " +
			"de-pairing consequence in I3 rests on them differing")
	}
}
