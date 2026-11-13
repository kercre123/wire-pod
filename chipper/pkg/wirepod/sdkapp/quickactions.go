package sdkapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// driveOnCharger sends the robot home and answers {"result": "<BehaviorResults>"}, or HTTP 500 "error: ..." when the call fails.
func driveOnCharger(ctx context.Context, w http.ResponseWriter, conn vectorpb.ExternalInterfaceClient) {
	resp, err := conn.DriveOnCharger(ctx, &vectorpb.DriveOnChargerRequest{})
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": resp.GetResult().String()})
}

// cloudIntent sends an AppIntent and answers "done", or HTTP 500 "error: ..." when the robot refuses it.
func cloudIntent(ctx context.Context, w http.ResponseWriter, conn vectorpb.ExternalInterfaceClient, intent string) {
	if _, err := conn.AppIntent(ctx, &vectorpb.AppIntentRequest{Intent: intent}); err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "done")
}
