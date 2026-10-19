package sdkapp

import (
	"net/http/httptest"
	"testing"
)

// withMDNSAlreadyRun pre-registers an address so RunMDNS returns immediately
// instead of performing a real browse. Without it these tests would fire a live
// mDNS query into the network.
func withMDNSAlreadyRun(t *testing.T, ip string) {
	t.Helper()
	restore := MDNSAlreadyRun
	t.Cleanup(func() { MDNSAlreadyRun = restore })
	MDNSAlreadyRun = append(MDNSAlreadyRun, ip)
}

// trackRobotAsChecking puts esn in the pinger list in the state a robot that is
// already checking in normally would be: tracked, not stopped. ShouldPingJdocs-
// ForESN then returns false, so the jdocs pull -- which would dial the robot --
// never happens, while still resetting the timer. That reset is the observable
// this file uses to tell which branch of connCheck ran.
func trackRobotAsChecking(t *testing.T, esn string, marker int) {
	t.Helper()
	ShouldPingJdocsForESN(esn)
	JdocsPingerBots.mu.Lock()
	defer JdocsPingerBots.mu.Unlock()
	for i := range JdocsPingerBots.Robots {
		if JdocsPingerBots.Robots[i].ESN == esn {
			JdocsPingerBots.Robots[i].TimeSinceLastCheck = marker
			return
		}
	}
	t.Fatalf("failed to seed pinger state for %s", esn)
}

// okRequest builds the request a Vector actually sends: a bare HEAD /ok, with
// ?emresn= only on the newer firmware path.
func okRequest(emresn, peer string) *httptest.ResponseRecorder {
	path := "/ok"
	if emresn != "" {
		path += "?emresn=" + emresn
	}
	r := httptest.NewRequest("HEAD", path, nil)
	r.RemoteAddr = peer + ":41234"
	w := httptest.NewRecorder()
	connCheck(w, r)
	return w
}

// Behind a proxy the peer address is the gateway and matches no robot, so
// without the reported ESN the request would fall to the mDNS branch on every
// single check. This asserts the whole path -- request in, pinger updated --
// rather than just the resolver's return value.
func TestConnCheckReachesThePingerViaReportedESNBehindAProxy(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.0.2.13"})
	withEmptyPingerList(t)
	withMDNSAlreadyRun(t, "172.28.0.1")
	trackRobotAsChecking(t, "00e20145", 9)

	w := okRequest("00e20145", "172.28.0.1") // the gateway, not the robot

	if body := w.Body.String(); body != "ok" {
		t.Errorf("connCheck body = %q, want \"ok\"; the robot treats anything else as unreachable", body)
	}
	if _, since, _ := pingerState("00e20145"); since != 0 {
		t.Errorf("TimeSinceLastCheck = %d, want 0. The request never reached the "+
			"pinger, so the robot is not being counted as checking in", since)
	}
}

// The regression the gating exists to prevent. On a stock install the peer
// address is authoritative, so a robot whose stored address is stale must NOT
// resolve -- because not resolving is what sends connCheck to the mDNS browse
// that rewrites the stale address. If this starts passing through to the pinger,
// the self-heal is gone and a robot that changes IP stays unreachable forever.
func TestConnCheckFallsToMDNSWhenTheStoredAddressIsStaleAndProxyModeIsOff(t *testing.T) {
	withProxyMode(t, false)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.0.2.13"})
	withEmptyPingerList(t)
	withMDNSAlreadyRun(t, "192.0.2.77")
	trackRobotAsChecking(t, "00e20145", 9)

	// The robot moved to .77; botSdkInfo.json still says .13.
	w := okRequest("00e20145", "192.0.2.77")

	if body := w.Body.String(); body != "ok" {
		t.Errorf("connCheck body = %q, want \"ok\"", body)
	}
	if _, since, _ := pingerState("00e20145"); since != 9 {
		t.Errorf("TimeSinceLastCheck = %d, want the marker 9 left untouched. The "+
			"request reached the pinger, which means it did NOT take the mDNS "+
			"branch -- the path that repairs a stale stored address", since)
	}
}

// With proxy mode off and the address matching, the ordinary path must still
// work end to end. Guards against a gate that fixes the stale case by breaking
// the common one.
func TestConnCheckReachesThePingerViaPeerAddressOnAStockInstall(t *testing.T) {
	withProxyMode(t, false)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.0.2.13"})
	withEmptyPingerList(t)
	withMDNSAlreadyRun(t, "192.0.2.13")
	trackRobotAsChecking(t, "00e20145", 9)

	okRequest("", "192.0.2.13")

	if _, since, _ := pingerState("00e20145"); since != 0 {
		t.Errorf("TimeSinceLastCheck = %d, want 0; a robot checking in from its "+
			"own stored address must still be recognised", since)
	}
}

// An address belonging to no known robot has nothing to resolve to, on any
// configuration, and must reach the mDNS branch so the robot can be discovered.
func TestConnCheckFallsToMDNSForAnUnknownRobot(t *testing.T) {
	withProxyMode(t, true)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.0.2.13"})
	withEmptyPingerList(t)
	withMDNSAlreadyRun(t, "192.0.2.99")
	trackRobotAsChecking(t, "00e20145", 9)

	okRequest("", "192.0.2.99")

	if _, since, _ := pingerState("00e20145"); since != 9 {
		t.Errorf("TimeSinceLastCheck = %d, want the marker 9; an unknown address "+
			"must not be attributed to an existing robot", since)
	}
}
