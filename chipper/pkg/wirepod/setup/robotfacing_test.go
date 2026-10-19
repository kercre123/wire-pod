package botsetup

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func configureRobotFacing(t *testing.T, addr string) {
	t.Helper()
	vars.APIConfig.Server.EPConfig = false
	vars.APIConfig.Server.Port = "443"
	vars.APIConfig.Server.RobotFacingAddress = addr
}

func TestGetRobotFacingAddressFallsBackToOutboundIP(t *testing.T) {
	redirectCertPaths(t)
	vars.APIConfig.Server.RobotFacingAddress = ""

	if got, want := vars.GetRobotFacingAddress(), vars.GetOutboundIP().String(); got != want {
		t.Errorf("GetRobotFacingAddress = %q, want the outbound IP %q", got, want)
	}
}

// A hostname is the case that matters for sub-domain routing: Go only sends
// SNI when ServerName is not an IP literal.
func TestCreateServerConfigUsesConfiguredHostname(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, "escapepod.local")

	CreateServerConfig()

	const want = `{"jdocs":"escapepod.local:443","tms":"escapepod.local:443",` +
		`"chipper":"escapepod.local:443","check":"escapepod.local/ok",` +
		`"logfiles":"s3://anki-device-logs-prod/victor","appkey":"oDoa0quieSeir6goowai7f"}`
	if got := readServerConfig(t); got != want {
		t.Errorf("server config with a hostname.\n got: %s\nwant: %s", got, want)
	}
}

func TestCreateServerConfigUsesConfiguredIP(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, "192.0.2.12")

	CreateServerConfig()

	var cfg ClientServerConfig
	if err := json.Unmarshal([]byte(readServerConfig(t)), &cfg); err != nil {
		t.Fatalf("parsing: %v", err)
	}
	if cfg.Jdocs != "192.0.2.12:443" {
		t.Errorf("jdocs = %q, want %q", cfg.Jdocs, "192.0.2.12:443")
	}
	if cfg.Check != "192.0.2.12/ok" {
		t.Errorf("check = %q, want %q (no port on the check endpoint)", cfg.Check, "192.0.2.12/ok")
	}
}

// The port must land exactly once. This is the doubled-port footgun that
// ValidateRobotFacingAddress exists to prevent reaching here at all.
func TestCreateServerConfigDoesNotDoubleThePort(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, "wirepod.lan")
	vars.APIConfig.Server.Port = "8443"

	CreateServerConfig()

	var cfg ClientServerConfig
	if err := json.Unmarshal([]byte(readServerConfig(t)), &cfg); err != nil {
		t.Fatalf("parsing: %v", err)
	}
	if cfg.Chipper != "wirepod.lan:8443" {
		t.Errorf("chipper = %q, want exactly one port", cfg.Chipper)
	}
}

// EP mode is out of scope: it serves the bundled ep.crt whose only SAN is
// DNS:escapepod.local, so honoring a robot-facing address there would produce
// a config the certificate cannot satisfy.
func TestCreateServerConfigIgnoresRobotFacingAddressInEPMode(t *testing.T) {
	redirectCertPaths(t)
	vars.APIConfig.Server.EPConfig = true
	vars.APIConfig.Server.Port = "443"
	vars.APIConfig.Server.RobotFacingAddress = "wirepod.lan"

	CreateServerConfig()

	const want = `{"jdocs":"escapepod.local:443","tms":"escapepod.local:443",` +
		`"chipper":"escapepod.local:443","check":"escapepod.local/ok",` +
		`"logfiles":"s3://anki-device-logs-prod/victor","appkey":"oDoa0quieSeir6goowai7f"}`
	if got := readServerConfig(t); got != want {
		t.Errorf("EP mode honored a robot-facing address.\n got: %s\nwant: %s", got, want)
	}
}

func TestCreateCertComboPutsHostnameInDNSNames(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, "escapepod.local")

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	cert := parseGeneratedCert(t)

	if len(cert.DNSNames) != 1 || cert.DNSNames[0] != "escapepod.local" {
		t.Errorf("DNSNames = %v, want [escapepod.local]", cert.DNSNames)
	}
	// The outbound IP must still be retained, or an install that previously
	// worked by IP would stop working.
	found := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(vars.GetOutboundIP()) {
			found = true
		}
	}
	if !found {
		t.Errorf("outbound IP %v was dropped from the SANs %v",
			vars.GetOutboundIP(), cert.IPAddresses)
	}

	// And the cert must actually verify for the configured name, which is what
	// the robot's gRPC client does.
	if err := cert.VerifyHostname("escapepod.local"); err != nil {
		t.Errorf("generated cert does not verify for its own configured name: %v", err)
	}
}

func TestCreateCertComboPutsIPInIPAddresses(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, "192.0.2.12")

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	cert := parseGeneratedCert(t)

	if len(cert.DNSNames) != 0 {
		t.Errorf("an IP address produced DNS SANs %v", cert.DNSNames)
	}
	want := net.ParseIP("192.0.2.12")
	found := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(want) {
			found = true
		}
	}
	if !found {
		t.Errorf("configured IP %v missing from SANs %v", want, cert.IPAddresses)
	}
	if err := cert.VerifyHostname("192.0.2.12"); err != nil {
		t.Errorf("generated cert does not verify for its configured IP: %v", err)
	}
}

// Configuring the outbound IP explicitly must not produce it twice.
func TestCreateCertComboDoesNotDuplicateTheOutboundIP(t *testing.T) {
	redirectCertPaths(t)
	configureRobotFacing(t, vars.GetOutboundIP().String())

	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	cert := parseGeneratedCert(t)

	if len(cert.IPAddresses) != 1 {
		t.Errorf("IP SANs = %v, want exactly one entry", cert.IPAddresses)
	}
}

// The certificate and the server config must name the same thing, since the
// robot verifies one against the other.
func TestCertAndServerConfigAgree(t *testing.T) {
	for _, addr := range []string{"escapepod.local", "192.0.2.12", ""} {
		func() {
			redirectCertPaths(t)
			configureRobotFacing(t, addr)

			if err := CreateCertCombo(); err != nil {
				t.Fatalf("CreateCertCombo(%q): %v", addr, err)
			}
			CreateServerConfig()

			var cfg ClientServerConfig
			if err := json.Unmarshal([]byte(readServerConfig(t)), &cfg); err != nil {
				t.Fatalf("parsing: %v", err)
			}
			host, _, err := net.SplitHostPort(cfg.Chipper)
			if err != nil {
				t.Fatalf("server config chipper %q is not host:port: %v", cfg.Chipper, err)
			}
			if err := parseGeneratedCert(t).VerifyHostname(host); err != nil {
				t.Errorf("robot-facing address %q: server config names %q but the "+
					"certificate does not cover it: %v", addr, host, err)
			}
		}()
	}
}
