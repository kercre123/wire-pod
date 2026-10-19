package tokenserver

import (
	"context"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// withProxyMode sets server.proxy_mode for the duration of a test. The token
// claim is only preferred behind a proxy.
func withProxyMode(t *testing.T, on bool) {
	t.Helper()
	restore := vars.APIConfig
	t.Cleanup(func() { vars.APIConfig = restore })
	vars.APIConfig.Server.ProxyMode = on
}

// Two robots behind one gateway is the case the preference exists for: the peer
// address resolves to whichever is stored first, so without the token claim a
// refresh from one robot is attributed to the other.
func TestProxyModePrefersTheESNTheTokenNames(t *testing.T) {
	withProxyMode(t, true)
	seedBotInfo(t, sharedPeerBotInfo)
	withKnownRobots(t, "00e20145", "00e20146")

	for _, want := range []string{"00e20145", "00e20146"} {
		got, err := resolveTokenRequestESN(ctxWithToken(issueJWT(t, "vic:"+want)), "172.19.0.1")
		if err != nil {
			t.Fatalf("resolveTokenRequestESN: unexpected error %v", err)
		}
		if got != want {
			t.Errorf("resolveTokenRequestESN = %q, want %q; behind a proxy the "+
				"token is the only thing that tells two robots apart", got, want)
		}
	}
}

// Without proxy_mode the peer address wins even when a token names someone
// else. This is the guard that matters: preferring the claim also clears the
// error, and that is what routes CreateJWT into the branch which persists
// SetBotGUID and ChangeGUIDInIni. The claim cannot be verified, so ungated it
// would let anyone who knows an ESN rotate that robot's GUID.
func TestWithoutProxyModeThePeerAddressWins(t *testing.T) {
	withProxyMode(t, false)
	seedBotInfo(t, twoRobotBotInfo)
	withKnownRobots(t, "00e20145", "00e20146")

	// A request genuinely from .150, carrying a token that claims to be .151's robot.
	got, err := resolveTokenRequestESN(ctxWithToken(issueJWT(t, "vic:00e20146")), "192.168.1.150")
	if err != nil {
		t.Fatalf("resolveTokenRequestESN: unexpected error %v", err)
	}
	if got != "00e20145" {
		t.Errorf("resolveTokenRequestESN = %q, want %q; on a direct LAN a token "+
			"claim must not override the address the request actually came from",
			got, "00e20145")
	}
}

// An unknown peer address must keep returning an error when there is no proxy.
// CreateJWT branches on that error, so swallowing it here would silently start
// persisting GUID changes for requests upstream treated as new robots.
func TestWithoutProxyModeAnUnknownAddressStillErrors(t *testing.T) {
	withProxyMode(t, false)
	seedBotInfo(t, twoRobotBotInfo)
	withKnownRobots(t, "00e20145")

	if _, err := resolveTokenRequestESN(ctxWithToken(issueJWT(t, "vic:00e20145")), "10.0.0.99"); err == nil {
		t.Error("an unrecognised peer address must still report an error; CreateJWT " +
			"branches on it to decide whether to persist a new GUID")
	}
}

// Behind a proxy the address is expected not to resolve, and the token is what
// rescues the request. Clearing the error here is deliberate.
func TestProxyModeResolvesEvenWhenTheAddressDoesNot(t *testing.T) {
	withProxyMode(t, true)
	seedBotInfo(t, twoRobotBotInfo)
	withKnownRobots(t, "00e20145")

	got, err := resolveTokenRequestESN(ctxWithToken(issueJWT(t, "vic:00e20145")), "172.19.0.1")
	if err != nil {
		t.Fatalf("resolveTokenRequestESN: unexpected error %v", err)
	}
	if got != "00e20145" {
		t.Errorf("resolveTokenRequestESN = %q, want %q", got, "00e20145")
	}
}

// A token naming a robot wire-pod has never stored must not be trusted even
// behind a proxy: it would otherwise clear the error and invent an identity.
func TestProxyModeIgnoresATokenNamingAnUnknownRobot(t *testing.T) {
	withProxyMode(t, true)
	seedBotInfo(t, twoRobotBotInfo)
	withKnownRobots(t, "00e20145")

	if _, err := resolveTokenRequestESN(ctxWithToken(issueJWT(t, "vic:00ffffff")), "10.0.0.99"); err == nil {
		t.Error("an unknown ESN in the token must fall through to the address path, " +
			"which here cannot resolve either")
	}
}

// No token at all is the primary-association case, and must behave exactly as
// it did before any of this.
func TestNoTokenFallsBackToTheAddress(t *testing.T) {
	withProxyMode(t, true)
	seedBotInfo(t, twoRobotBotInfo)
	withKnownRobots(t, "00e20145")

	got, err := resolveTokenRequestESN(context.Background(), "192.168.1.150")
	if err != nil {
		t.Fatalf("resolveTokenRequestESN: unexpected error %v", err)
	}
	if got != "00e20145" {
		t.Errorf("resolveTokenRequestESN = %q, want %q from the peer address", got, "00e20145")
	}
}

// Two robots whose stored ip_address is the gateway, which is what wire-pod
// records for every robot behind a proxy.
const sharedPeerBotInfo = `{"global_guid":"g","robots":[` +
	`{"esn":"00e20145","ip_address":"172.19.0.1","guid":"g1","activated":true},` +
	`{"esn":"00e20146","ip_address":"172.19.0.1","guid":"g2","activated":true}]}`
