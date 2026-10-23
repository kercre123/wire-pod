package sdkapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func TestDisplayNameEndpointsPersistWithoutRobotConnection(t *testing.T) {
	originalInfo, originalPath := vars.BotInfo, vars.BotInfoPath
	vars.BotInfo = vars.RobotInfoStore{Robots: []vars.RobotInfo{{Esn: "robot"}}}
	vars.BotInfoPath = filepath.Join(t.TempDir(), "botSdkInfo.json")
	t.Cleanup(func() {
		vars.BotInfo = originalInfo
		vars.BotInfoPath = originalPath
	})

	save := httptest.NewRecorder()
	SdkapiHandler(save, httptest.NewRequest(http.MethodPost, "/api-sdk/set_display_name?serial=robot&name=Sunbeam", nil))
	if save.Code != http.StatusOK {
		t.Fatalf("save status = %d, want %d: %s", save.Code, http.StatusOK, save.Body.String())
	}

	lookup := httptest.NewRecorder()
	SdkapiHandler(lookup, httptest.NewRequest(http.MethodGet, "/api-sdk/get_display_name?serial=robot", nil))
	if lookup.Code != http.StatusOK {
		t.Fatalf("lookup status = %d, want %d: %s", lookup.Code, http.StatusOK, lookup.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(lookup.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := payload["display_name"]; got != "Sunbeam" {
		t.Fatalf("display_name = %q, want Sunbeam", got)
	}

	contents, err := os.ReadFile(vars.BotInfoPath)
	if err != nil {
		t.Fatalf("read persisted robot metadata: %v", err)
	}
	var persisted vars.RobotInfoStore
	if err := json.Unmarshal(contents, &persisted); err != nil {
		t.Fatalf("decode persisted robot metadata: %v", err)
	}
	if got := persisted.Robots[0].DisplayName; got != "Sunbeam" {
		t.Fatalf("persisted display name = %q, want Sunbeam", got)
	}
}
