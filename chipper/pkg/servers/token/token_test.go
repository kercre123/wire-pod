package tokenserver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// GetEsnFromTarget maps an inbound peer address back to an ESN by matching
// ip_address in botSdkInfo.json. This is invariant I1(b): ip_address is an
// identity key, not just a dial target.
//
// The whole proxy-networking design exists to avoid perturbing this. Feature B
// stores the dial override in apiConfig.json specifically so that no admin
// action can change what this function sees.
func seedBotInfo(t *testing.T, contents string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "botSdkInfo.json")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("seeding botSdkInfo: %v", err)
	}
	restore := vars.BotInfoPath
	t.Cleanup(func() { vars.BotInfoPath = restore })
	vars.BotInfoPath = path
}

const twoRobotBotInfo = `{"global_guid":"tni1TRsTRTaNSapjo0Y+Sw==","robots":[` +
	`{"esn":"00e20145","ip_address":"192.168.1.150","guid":"g1","activated":true},` +
	`{"esn":"00e20146","ip_address":"192.168.1.151","guid":"g2","activated":true}]}`

func TestGetEsnFromTargetMatchesObservedAddress(t *testing.T) {
	seedBotInfo(t, twoRobotBotInfo)

	for _, tc := range []struct{ target, wantESN string }{
		{"192.168.1.150", "00e20145"},
		{"192.168.1.151", "00e20146"},
	} {
		got, err := GetEsnFromTarget(tc.target)
		if err != nil {
			t.Errorf("GetEsnFromTarget(%q) returned error: %v", tc.target, err)
			continue
		}
		if got != tc.wantESN {
			t.Errorf("GetEsnFromTarget(%q) = %q, want %q", tc.target, got, tc.wantESN)
		}
	}
}

func TestGetEsnFromTargetFailsForUnknownAddress(t *testing.T) {
	seedBotInfo(t, twoRobotBotInfo)

	if _, err := GetEsnFromTarget("10.0.0.99"); err == nil {
		t.Error("GetEsnFromTarget accepted an address that is in no robot record; " +
			"token auth depends on this failing so it can fall into its new-robot branch")
	}
}

// The proxied deployment relies on this: every robot presents as the bridge
// gateway, the stored ip_address equals that peer address, and identity
// matching therefore still succeeds even though dialing that address fails.
// Feature B must not disturb it.
func TestGetEsnFromTargetWorksWhenAddressIsAProxyPeer(t *testing.T) {
	seedBotInfo(t, `{"global_guid":"g","robots":[`+
		`{"esn":"00e20145","ip_address":"172.19.0.1","guid":"g1","activated":true}]}`)

	got, err := GetEsnFromTarget("172.19.0.1")
	if err != nil {
		t.Fatalf("GetEsnFromTarget on a proxy peer address failed: %v", err)
	}
	if got != "00e20145" {
		t.Errorf("GetEsnFromTarget = %q, want %q", got, "00e20145")
	}
}

// Documents the known limitation in §10 so a future change that appears to
// "fix" it is a deliberate decision rather than an accident: with multiple
// robots behind one proxy, the first match wins.
func TestGetEsnFromTargetReturnsFirstMatchForSharedPeerAddress(t *testing.T) {
	seedBotInfo(t, `{"global_guid":"g","robots":[`+
		`{"esn":"00e20145","ip_address":"172.19.0.1","guid":"g1","activated":true},`+
		`{"esn":"00e20146","ip_address":"172.19.0.1","guid":"g2","activated":true}]}`)

	got, err := GetEsnFromTarget("172.19.0.1")
	if err != nil {
		t.Fatalf("GetEsnFromTarget failed: %v", err)
	}
	if got != "00e20145" {
		t.Errorf("GetEsnFromTarget = %q, want first match %q "+
			"(multi-robot inbound identity is a known limitation, §10)", got, "00e20145")
	}
}

// §7.2, the regression this whole design exists to prevent.
//
// Feature B lets an admin redirect where wire-pod dials a robot. That must be
// invisible to identity matching: GetEsnFromTarget reads botSdkInfo.json from
// disk and matches the observed peer address, and no amount of admin
// configuration may change what it returns.
//
// The naive alternative — pinning ip_address itself — would break exactly this,
// sending token auth into its "act as if this is a new robot" branch.
func TestConfiguredDialTargetDoesNotPerturbGetEsnFromTarget(t *testing.T) {
	seedBotInfo(t, `{"global_guid":"g","robots":[`+
		`{"esn":"00e20145","ip_address":"172.19.0.1","guid":"g1","activated":true}]}`)

	restore := vars.APIConfig
	t.Cleanup(func() { vars.APIConfig = restore })

	// Baseline: the proxied deployment resolves by peer address.
	before, err := GetEsnFromTarget("172.19.0.1")
	if err != nil {
		t.Fatalf("baseline GetEsnFromTarget failed: %v", err)
	}

	// Now point dialing at the robot's real LAN address.
	vars.SetRobotLANAddress("00e20145", "192.0.2.13")

	after, err := GetEsnFromTarget("172.19.0.1")
	if err != nil {
		t.Fatalf("GetEsnFromTarget failed after configuring a dial override: %v", err)
	}
	if after != before {
		t.Errorf("configuring a dial override changed identity resolution: %q -> %q", before, after)
	}

	// The configured LAN address must NOT become an identity key.
	if _, err := GetEsnFromTarget("192.0.2.13"); err == nil {
		t.Error("the configured dial address resolved as an identity key; " +
			"only the observed peer address may do that")
	}

	// And the dial target really did change, so the test is not vacuous.
	if got := vars.GetRobotDialTarget("00e20145", "172.19.0.1"); got != "192.0.2.13" {
		t.Errorf("dial target = %q, want the configured %q", got, "192.0.2.13")
	}
}
