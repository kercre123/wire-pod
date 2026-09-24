package sdkapp

import (
	"net/http/httptest"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func withRobots(t *testing.T, robots ...struct{ esn, ip string }) {
	t.Helper()
	restore := vars.BotInfo
	t.Cleanup(func() { vars.BotInfo = restore })

	vars.BotInfo = vars.RobotInfoStore{GlobalGUID: "g"}
	for _, r := range robots {
		vars.BotInfo.Robots = append(vars.BotInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
			Activated bool   `json:"activated"`
		}{Esn: r.esn, IPAddress: r.ip, GUID: "guid-" + r.esn, Activated: true})
	}
}

// withProxyMode sets server.proxy_mode for the duration of a test. The reported
// ESN is only preferred behind a proxy, so anything exercising emresn needs it.
func withProxyMode(t *testing.T, on bool) {
	t.Helper()
	restore := vars.APIConfig
	t.Cleanup(func() { vars.APIConfig = restore })
	vars.APIConfig.Server.ProxyMode = on
}

// The whole point: behind a proxy every robot presents the same peer address,
// and the ESN the robot itself reports is what disambiguates them.
func TestResolveConnCheckESNPrefersReportedESN(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t,
		struct{ esn, ip string }{"00e20145", "172.19.0.1"},
		struct{ esn, ip string }{"00e20146", "172.19.0.1"},
	)

	for _, want := range []string{"00e20145", "00e20146"} {
		r := httptest.NewRequest("HEAD", "/ok?emresn="+want+"&ankiversion=1.8.1", nil)
		r.RemoteAddr = "172.19.0.1:41234"

		if got := resolveConnCheckESN(r); got != want {
			t.Errorf("two robots behind one gateway: resolveConnCheckESN = %q, want %q "+
				"(the reported ESN must win over the shared peer address)", got, want)
		}
	}
}

// Without emresn the peer address is all there is, and the first match wins --
// the pre-existing behavior this change does not alter.
func TestResolveConnCheckESNFallsBackToPeerAddress(t *testing.T) {
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	r := httptest.NewRequest("HEAD", "/ok", nil)
	r.RemoteAddr = "192.168.1.150:41234"

	if got := resolveConnCheckESN(r); got != "00e20145" {
		t.Errorf("resolveConnCheckESN = %q, want %q from the peer address", got, "00e20145")
	}
}

// An ESN we have never stored must not short-circuit the address path; a
// connection check is not an adoption.
func TestResolveConnCheckESNIgnoresUnknownESN(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	r := httptest.NewRequest("HEAD", "/ok?emresn=00ffffff", nil)
	r.RemoteAddr = "192.168.1.150:41234"

	if got := resolveConnCheckESN(r); got != "00e20145" {
		t.Errorf("resolveConnCheckESN = %q, want the peer-address match %q; an unknown "+
			"reported ESN must fall through rather than invent a robot", got, "00e20145")
	}
}

func TestResolveConnCheckESNMatchesESNCaseInsensitively(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t, struct{ esn, ip string }{"00e20145", "172.19.0.1"})

	r := httptest.NewRequest("HEAD", "/ok?emresn=00E20145", nil)
	r.RemoteAddr = "172.19.0.1:41234"

	if got := resolveConnCheckESN(r); got != "00e20145" {
		t.Errorf("resolveConnCheckESN = %q, want %q", got, "00e20145")
	}
}

// Returning "" is what preserves the existing mDNS-browse branch.
func TestResolveConnCheckESNReturnsEmptyWhenUnidentifiable(t *testing.T) {
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	r := httptest.NewRequest("HEAD", "/ok", nil)
	r.RemoteAddr = "10.0.0.99:41234"

	if got := resolveConnCheckESN(r); got != "" {
		t.Errorf("resolveConnCheckESN = %q, want \"\" so the mDNS fallback still runs", got)
	}
}

// Without proxy_mode the peer address stays authoritative even when the robot
// reports an ESN. That is what keeps the mDNS self-heal alive on a stock
// install: returning "" here is what makes connCheck fall through to the browse
// that rewrites a robot's stored ip_address after DHCP moves it. Preferring
// emresn unconditionally would retire that path without anyone noticing.
func TestResolveConnCheckESNIgnoresReportedESNWithoutProxyMode(t *testing.T) {
	withProxyMode(t, false)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	// The robot moved; the address on file is stale.
	r := httptest.NewRequest("HEAD", "/ok?emresn=00e20145", nil)
	r.RemoteAddr = "192.168.1.77:41234"

	if got := resolveConnCheckESN(r); got != "" {
		t.Errorf("resolveConnCheckESN = %q, want \"\" so connCheck still reaches the "+
			"mDNS browse that repairs the stored address", got)
	}
}

// A proxied robot whose stored ip_address is its real LAN address (because
// Feature B or mDNS recorded it) would previously fail the peer-address match
// entirely and trigger an mDNS browse on every check. emresn fixes that.
func TestReportedESNResolvesWhenPeerAddressDoesNotMatch(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.0.2.13"})

	r := httptest.NewRequest("HEAD", "/ok?emresn=00e20145", nil)
	r.RemoteAddr = "172.19.0.1:41234" // the gateway, not the robot

	if got := resolveConnCheckESN(r); got != "00e20145" {
		t.Errorf("resolveConnCheckESN = %q, want %q; the robot is identifiable even "+
			"though its stored address does not match the peer", got, "00e20145")
	}
}

func TestESNForAddressIgnoresPortAndCase(t *testing.T) {
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	for _, addr := range []string{"192.168.1.150", "192.168.1.150:41234", " 192.168.1.150 "} {
		if got := esnForAddress(addr); got != "00e20145" {
			t.Errorf("esnForAddress(%q) = %q, want %q", addr, got, "00e20145")
		}
	}
	if got := esnForAddress("10.0.0.1"); got != "" {
		t.Errorf("esnForAddress on an unknown address = %q, want \"\"", got)
	}
}
