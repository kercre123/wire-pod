package tokenserver

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc/peer"
)

// withPersistencePaths redirects everything CreateJWT can write. SDKIniPath is
// pointed at an empty directory on purpose: ChangeGUIDInIni logs and returns
// when the ini is absent, which keeps the test off the real SDK config while
// still exercising the branch that calls it.
func withPersistencePaths(t *testing.T, botInfo string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "botSdkInfo.json")
	if err := os.WriteFile(path, []byte(botInfo), 0644); err != nil {
		t.Fatalf("seeding botSdkInfo: %v", err)
	}
	restorePath, restoreIni, restoreInfo := vars.BotInfoPath, vars.SDKIniPath, vars.BotInfo
	t.Cleanup(func() {
		vars.BotInfoPath, vars.SDKIniPath, vars.BotInfo = restorePath, restoreIni, restoreInfo
	})
	vars.BotInfoPath = path
	vars.SDKIniPath = dir + string(os.PathSeparator)
	if err := json.Unmarshal([]byte(botInfo), &vars.BotInfo); err != nil {
		t.Fatalf("loading botSdkInfo: %v", err)
	}
	return path
}

// storedGUID reads a robot's GUID back off disk, which is what the next process
// and the SDK will use.
func storedGUID(t *testing.T, path, esn string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading botSdkInfo: %v", err)
	}
	var store struct {
		Robots []struct {
			Esn  string `json:"esn"`
			GUID string `json:"guid"`
		} `json:"robots"`
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		t.Fatalf("parsing botSdkInfo: %v", err)
	}
	for _, r := range store.Robots {
		if r.Esn == esn {
			return r.GUID
		}
	}
	return ""
}

// ctxFrom builds the context CreateJWT actually receives: a peer address, plus
// the robot's own token as metadata.
func ctxFrom(peerIP, jwtToken string) context.Context {
	base := context.Background()
	if jwtToken != "" {
		base = ctxWithToken(jwtToken)
	}
	return peer.NewContext(base, &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP(peerIP), Port: 41234},
	})
}

const twoRobotsForPersistence = `{"global_guid":"g","robots":[` +
	`{"esn":"00e20145","ip_address":"192.0.2.150","guid":"guid-A","activated":true},` +
	`{"esn":"00e20146","ip_address":"192.0.2.151","guid":"guid-B","activated":true}]}`

// The consequence that matters, tested end to end rather than at the resolver.
//
// On a stock install a request arriving from an address wire-pod does not
// recognise is treated as a new robot: nothing is persisted. If the token's
// claim were honoured here it would clear the error, select the other branch,
// and rewrite a stored robot's GUID -- which invalidates that robot's existing
// SDK and app pairing. Knowing an eight-character ESN would be enough to do it
// to someone else's robot.
func TestStockInstallDoesNotRewriteAGUIDFromATokenClaim(t *testing.T) {
	withProxyMode(t, false)
	path := withPersistencePaths(t, twoRobotsForPersistence)
	before := storedGUID(t, path, "00e20146")

	// Request from an address belonging to no robot, carrying a token that
	// claims to be robot B.
	CreateJWT(ctxFrom("192.0.2.99", issueJWT(t, "vic:00e20146")), false, false)

	if after := storedGUID(t, path, "00e20146"); after != before {
		t.Errorf("robot B's stored GUID changed from %q to %q on a stock install. "+
			"An unverifiable token claim was allowed to re-pair a robot the request "+
			"did not come from", before, after)
	}
}

// Behind a proxy the same request must be honoured, because there the peer
// address is the gateway and the claim is the only way to tell robots apart.
// This is the other half: the gate must not be so tight that the feature stops
// working where it is needed.
func TestProxyInstallDoesRewriteTheGUIDTheTokenNames(t *testing.T) {
	withProxyMode(t, true)
	path := withPersistencePaths(t, twoRobotsForPersistence)
	before := storedGUID(t, path, "00e20146")

	CreateJWT(ctxFrom("172.28.0.1", issueJWT(t, "vic:00e20146")), false, false)

	if after := storedGUID(t, path, "00e20146"); after == before {
		t.Errorf("robot B's GUID was not updated behind a proxy (still %q). The "+
			"token is the only per-robot signal there, so a refresh cannot complete "+
			"without it", before)
	}
}

// A robot refreshing from its own recognised address is the ordinary case and
// must keep working with no proxy configured at all.
func TestStockInstallStillRefreshesFromAKnownAddress(t *testing.T) {
	withProxyMode(t, false)
	path := withPersistencePaths(t, twoRobotsForPersistence)
	before := storedGUID(t, path, "00e20145")

	CreateJWT(ctxFrom("192.0.2.150", ""), false, false)

	if after := storedGUID(t, path, "00e20145"); after == before {
		t.Errorf("a robot refreshing from its own stored address %q did not get a "+
			"new GUID; the ordinary path is broken", before)
	}
}

// The claim must never be able to name a robot wire-pod has never seen, even
// behind a proxy, or a token alone could conjure an identity.
func TestATokenNamingAnUnknownRobotPersistsNothing(t *testing.T) {
	withProxyMode(t, true)
	path := withPersistencePaths(t, twoRobotsForPersistence)
	beforeA := storedGUID(t, path, "00e20145")
	beforeB := storedGUID(t, path, "00e20146")

	CreateJWT(ctxFrom("192.0.2.99", issueJWT(t, "vic:00ffffff")), false, false)

	if storedGUID(t, path, "00e20145") != beforeA || storedGUID(t, path, "00e20146") != beforeB {
		t.Error("a token naming an unknown ESN disturbed a stored robot's GUID")
	}
}
