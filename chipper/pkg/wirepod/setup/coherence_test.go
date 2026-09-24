package botsetup

import (
	"os"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// captureLog returns whatever WarnIfCertMissesAddress wrote to the log tray.
func captureLog(t *testing.T, fn func()) string {
	t.Helper()
	before := len(logger.LogTrayArray)
	fn()
	return strings.Join(logger.LogTrayArray[before:], "\n")
}

func TestWarnsWhenCertDoesNotCoverConfiguredAddress(t *testing.T) {
	redirectCertPaths(t)

	// Generate a certificate for the outbound IP only...
	vars.APIConfig.Server.RobotFacingAddress = ""
	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	// ...then point the config at a name it cannot possibly cover.
	out := captureLog(t, func() { WarnIfCertMissesAddress("wirepod.lan") })

	if !strings.Contains(out, "does not cover") || !strings.Contains(out, "wirepod.lan") {
		t.Errorf("expected a warning naming the uncovered address, got: %s", out)
	}
	if !strings.Contains(out, "de-pairs") {
		t.Errorf("the warning should say regenerating de-pairs paired robots, got: %s", out)
	}
}

func TestDoesNotWarnWhenCertCoversConfiguredAddress(t *testing.T) {
	redirectCertPaths(t)

	vars.APIConfig.Server.RobotFacingAddress = "wirepod.lan"
	if err := CreateCertCombo(); err != nil {
		t.Fatalf("CreateCertCombo: %v", err)
	}
	out := captureLog(t, func() { WarnIfCertMissesAddress("wirepod.lan") })

	if strings.Contains(out, "WARNING") {
		t.Errorf("warned about an address the certificate does cover: %s", out)
	}
}

// A missing certificate is normal during first-run setup and must stay quiet.
func TestDoesNotWarnWhenNoCertExists(t *testing.T) {
	redirectCertPaths(t)
	os.Remove(vars.CertPath)

	out := captureLog(t, func() { WarnIfCertMissesAddress("wirepod.lan") })
	if strings.Contains(out, "WARNING") {
		t.Errorf("warned when no certificate exists yet: %s", out)
	}
}

func TestWarnsWhenCertIsUnparseable(t *testing.T) {
	redirectCertPaths(t)
	if err := os.WriteFile(vars.CertPath, []byte("not a certificate"), 0644); err != nil {
		t.Fatalf("writing junk cert: %v", err)
	}

	out := captureLog(t, func() { WarnIfCertMissesAddress("wirepod.lan") })
	if !strings.Contains(out, "WARNING") {
		t.Errorf("stayed silent about an unparseable certificate: %s", out)
	}
}

// Valid PEM framing wrapping meaningless DER. TestWarnsWhenCertIsUnparseable
// writes bytes that are not PEM at all, so it stops at pem.Decode and never
// reaches x509.ParseCertificate; this is the only case that does.
func TestWarnsWhenTheCertificateBodyCannotBeParsed(t *testing.T) {
	redirectCertPaths(t)

	const notACert = "-----BEGIN CERTIFICATE-----\nZm9vYmFy\n-----END CERTIFICATE-----\n"
	if err := os.WriteFile(vars.CertPath, []byte(notACert), 0644); err != nil {
		t.Fatalf("writing cert: %v", err)
	}

	out := captureLog(t, func() { WarnIfCertMissesAddress("wirepod.lan") })
	if !strings.Contains(out, "could not parse") {
		t.Errorf("expected the parse failure to be reported, got: %s", out)
	}
}
