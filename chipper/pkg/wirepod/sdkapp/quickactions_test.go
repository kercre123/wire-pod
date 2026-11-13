package sdkapp

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"google.golang.org/grpc"
)

// fakeConn stands in for the robot: tests never touch a real one.
type fakeConn struct {
	vectorpb.ExternalInterfaceClient
	res    vectorpb.BehaviorResults
	intent string
	err    error
	n      int
}

func (f *fakeConn) DriveOnCharger(context.Context, *vectorpb.DriveOnChargerRequest, ...grpc.CallOption) (*vectorpb.DriveOnChargerResponse, error) {
	f.n++
	return &vectorpb.DriveOnChargerResponse{Result: f.res}, f.err
}

func (f *fakeConn) AppIntent(_ context.Context, in *vectorpb.AppIntentRequest, _ ...grpc.CallOption) (*vectorpb.AppIntentResponse, error) {
	f.intent = in.Intent
	return &vectorpb.AppIntentResponse{}, f.err
}

func TestCloudIntent(t *testing.T) {
	f := &fakeConn{}
	w := httptest.NewRecorder()
	cloudIntent(context.Background(), w, f, "explore_start")
	if w.Code != 200 || w.Body.String() != "done" || f.intent != "explore_start" {
		t.Fatalf("got %d %q %q", w.Code, w.Body.String(), f.intent)
	}
	f = &fakeConn{err: errors.New("rpc error: busy")}
	w = httptest.NewRecorder()
	cloudIntent(context.Background(), w, f, "x")
	if w.Code != 500 || !strings.HasPrefix(w.Body.String(), "error: rpc error") {
		t.Fatalf("error path: got %d %q", w.Code, w.Body.String())
	}
}

func TestDriveOnCharger(t *testing.T) {
	f := &fakeConn{res: vectorpb.BehaviorResults_BEHAVIOR_COMPLETE_STATE}
	w := httptest.NewRecorder()
	driveOnCharger(context.Background(), w, f)
	if w.Code != 200 || f.n != 1 || !strings.Contains(w.Body.String(), `"result":"BEHAVIOR_COMPLETE_STATE"`) {
		t.Fatalf("got %d %q calls=%d", w.Code, w.Body.String(), f.n)
	}
	f = &fakeConn{err: errors.New("rpc error: unavailable")}
	w = httptest.NewRecorder()
	driveOnCharger(context.Background(), w, f)
	if w.Code != 500 || !strings.HasPrefix(w.Body.String(), "error: rpc error") {
		t.Fatalf("error path: got %d %q", w.Code, w.Body.String())
	}
}
