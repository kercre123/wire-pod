package vars

import (
	"encoding/json"
	"testing"
)

// botSdkInfo.json is machine-observed state and its on-disk shape must not
// change. Admin intent lives in apiConfig.json precisely so this file, the
// anonymous struct literal at jdocs/botInfoStorer.go:144, and token.go's
// duplicate RobotInfoStore all stay untouched.
//
// GetEsnFromTarget matches ip_address in these bytes to identify a robot, so a
// changed key here changes identity resolution, not just storage.
const baselineBotInfoJSON = `{"global_guid":"tni1TRsTRTaNSapjo0Y+Sw==",` +
	`"robots":[{"esn":"00e20145","ip_address":"192.168.1.150","guid":"abc123","activated":true}]}`

func TestBotInfoRoundTripsUnchanged(t *testing.T) {
	var info RobotInfoStore
	if err := json.Unmarshal([]byte(baselineBotInfoJSON), &info); err != nil {
		t.Fatalf("unmarshalling baseline botSdkInfo: %v", err)
	}
	got, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("re-marshalling baseline botSdkInfo: %v", err)
	}
	if string(got) != baselineBotInfoJSON {
		t.Errorf("botSdkInfo.json did not round-trip. A field was added to "+
			"RobotInfoStore; it belongs in apiConfig.json instead.\n got: %s\nwant: %s",
			got, baselineBotInfoJSON)
	}
}

func TestBotInfoHasNoAddressOverrideField(t *testing.T) {
	var info RobotInfoStore
	if err := json.Unmarshal([]byte(baselineBotInfoJSON), &info); err != nil {
		t.Fatalf("unmarshalling baseline botSdkInfo: %v", err)
	}
	raw, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}

	var decoded struct {
		Robots []map[string]json.RawMessage `json:"robots"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decoding robots: %v", err)
	}
	if len(decoded.Robots) != 1 {
		t.Fatalf("expected 1 robot, got %d", len(decoded.Robots))
	}

	want := map[string]bool{"esn": true, "ip_address": true, "guid": true, "activated": true}
	for key := range decoded.Robots[0] {
		if !want[key] {
			t.Errorf("unexpected key %q in a robot record; per-robot admin state "+
				"belongs in apiConfig.json, not botSdkInfo.json", key)
		}
	}
	for key := range want {
		if _, ok := decoded.Robots[0][key]; !ok {
			t.Errorf("baseline key %q disappeared from the robot record", key)
		}
	}
}
