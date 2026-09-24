package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc/metadata"
)

// issueJWT mints a token the way CreateJWT does: signed with a throwaway key,
// which is exactly why the claims can only ever be parsed unverified.
func issueJWT(t *testing.T, requestorID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"requestor_id": requestorID,
		"token_type":   "user+robot",
		"user_id":      UserId,
	})
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("signing: %v", err)
	}
	return signed
}

func ctxWithToken(jwtToken string) context.Context {
	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("anki-app-key", "oDoa0quieSeir6goowai7f", "anki-access-token", jwtToken))
}

func withKnownRobots(t *testing.T, esns ...string) {
	t.Helper()
	restore := vars.BotInfo
	t.Cleanup(func() { vars.BotInfo = restore })

	vars.BotInfo = vars.RobotInfoStore{GlobalGUID: "g"}
	for _, esn := range esns {
		vars.BotInfo.Robots = append(vars.BotInfo.Robots, struct {
			Esn       string `json:"esn"`
			IPAddress string `json:"ip_address"`
			GUID      string `json:"guid"`
			Activated bool   `json:"activated"`
		}{Esn: esn, IPAddress: "172.19.0.1", GUID: "guid-" + esn, Activated: true})
	}
}

// The point of the change: two robots share a peer address, and the token is
// what tells them apart on a refresh.
func TestESNFromRequestTokenIdentifiesTheRightRobot(t *testing.T) {
	withKnownRobots(t, "00e20145", "00e20146")

	for _, want := range []string{"00e20145", "00e20146"} {
		got := esnFromRequestToken(ctxWithToken(issueJWT(t, "vic:"+want)))
		if got != want {
			t.Errorf("esnFromRequestToken = %q, want %q; the token must disambiguate "+
				"robots sharing a peer address", got, want)
		}
	}
}

func TestESNFromRequestTokenReturnsEmptyWithoutMetadata(t *testing.T) {
	withKnownRobots(t, "00e20145")

	if got := esnFromRequestToken(context.Background()); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\" when there is no metadata", got)
	}
	empty := metadata.NewIncomingContext(context.Background(), metadata.Pairs("anki-app-key", "x"))
	if got := esnFromRequestToken(empty); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\" when no token is present", got)
	}
}

// An unparseable token must fall through to the address path, not error out.
func TestESNFromRequestTokenIgnoresGarbage(t *testing.T) {
	withKnownRobots(t, "00e20145")

	if got := esnFromRequestToken(ctxWithToken("not-a-jwt")); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\" for an unparseable token", got)
	}
}

// A token naming a robot we have never stored must not select anything. This
// is what stops the claim from inventing a robot, and it is also what makes
// the pre-association placeholder requestor_id harmless.
func TestESNFromRequestTokenRejectsUnknownRobot(t *testing.T) {
	withKnownRobots(t, "00e20145")

	if got := esnFromRequestToken(ctxWithToken(issueJWT(t, "vic:00ffffff"))); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\" for a robot not in the store", got)
	}
	// The placeholder wire-pod issues before a robot is associated.
	if got := esnFromRequestToken(ctxWithToken(issueJWT(t, "vic:00601b50"))); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\" for the pre-association "+
			"placeholder requestor_id", got)
	}
}

func TestESNFromRequestTokenHandlesMalformedRequestorID(t *testing.T) {
	withKnownRobots(t, "00e20145")

	for _, requestor := range []string{"", "vic:", "   ", "00e20145"} {
		got := esnFromRequestToken(ctxWithToken(issueJWT(t, requestor)))
		// A bare ESN with no "vic:" prefix is still usable; the others are not.
		want := ""
		if requestor == "00e20145" {
			want = "00e20145"
		}
		if got != want {
			t.Errorf("requestor_id %q: esnFromRequestToken = %q, want %q", requestor, got, want)
		}
	}
}

func TestESNFromRequestTokenMatchesCaseInsensitively(t *testing.T) {
	withKnownRobots(t, "00e20145")

	if got := esnFromRequestToken(ctxWithToken(issueJWT(t, "vic:00E20145"))); got != "00e20145" {
		t.Errorf("esnFromRequestToken = %q, want the stored casing %q", got, "00e20145")
	}
}

// Primary association must be untouched: it carries no usable token, and its
// branch ignores the resolved ESN anyway. This pins that the resolver stays
// silent there rather than relying on the branch to save us.
func TestPrimaryAssociationCarriesNoUsableToken(t *testing.T) {
	withKnownRobots(t, "00e20145", "00e20146")

	// What queue.go:119-131 actually sends on a primary association.
	primary := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("anki-app-key", "oDoa0quieSeir6goowai7f", "anki-user-session", "sess-abc"))

	if got := esnFromRequestToken(primary); got != "" {
		t.Errorf("esnFromRequestToken = %q, want \"\"; a primary association sends no "+
			"access token and must not resolve to any robot", got)
	}
}
