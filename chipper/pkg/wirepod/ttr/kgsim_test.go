package wirepod_ttr

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// fakeRobot answers just enough of the robot API for StreamingKGSim and
// reports every behavior control release it receives.
type fakeRobot struct {
	vectorpb.UnimplementedExternalInterfaceServer
	released chan struct{}
}

func (f *fakeRobot) BatteryState(context.Context, *vectorpb.BatteryStateRequest) (*vectorpb.BatteryStateResponse, error) {
	return &vectorpb.BatteryStateResponse{}, nil
}

func (f *fakeRobot) PlayAnimation(context.Context, *vectorpb.PlayAnimationRequest) (*vectorpb.PlayAnimationResponse, error) {
	time.Sleep(20 * time.Millisecond)
	return &vectorpb.PlayAnimationResponse{}, nil
}

func (f *fakeRobot) SayText(context.Context, *vectorpb.SayTextRequest) (*vectorpb.SayTextResponse, error) {
	return &vectorpb.SayTextResponse{}, nil
}

func (f *fakeRobot) BehaviorControl(stream vectorpb.ExternalInterface_BehaviorControlServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil
		}
		if req.GetControlRequest() != nil {
			stream.Send(&vectorpb.BehaviorControlResponse{
				ResponseType: &vectorpb.BehaviorControlResponse_ControlGrantedResponse{
					ControlGrantedResponse: &vectorpb.ControlGrantedResponse{},
				},
			})
		}
		if req.GetControlRelease() != nil {
			select {
			case f.released <- struct{}{}:
			default:
			}
		}
	}
}

// startFakeRobot serves fakeRobot over TLS on a unix socket and registers it
// in the bot info under esn. botconn dials "<ip>:443", so the socket path
// carries that suffix.
func startFakeRobot(t *testing.T, esn string) *fakeRobot {
	dir, err := os.MkdirTemp("", "kgsim")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	certSrv := httptest.NewUnstartedServer(nil)
	certSrv.StartTLS()
	cert := certSrv.TLS.Certificates[0]
	certSrv.Close()

	sock := filepath.Join(dir, "robot.sock")
	lis, err := net.Listen("unix", sock+":443")
	if err != nil {
		t.Fatal(err)
	}
	robot := &fakeRobot{released: make(chan struct{}, 1)}
	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})))
	vectorpb.RegisterExternalInterfaceServer(srv, robot)
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)

	vars.BotInfoPath = filepath.Join(dir, "botSdkInfo.json")
	if err := vars.UpdateBotInfo(func(b *vars.RobotInfoStore) {
		b.Robots = []vars.BotInfoRobot{{Esn: esn, IPAddress: "unix://" + sock, GUID: "guid"}}
	}); err != nil {
		t.Fatal(err)
	}
	return robot
}

func TestStreamingKGSimReleasesControlWhenLLMFails(t *testing.T) {
	cases := []struct {
		name string
		llm  http.HandlerFunc
	}{
		{
			name: "empty response",
			llm: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"\"}}]}\n\n")
				fmt.Fprint(w, "data: [DONE]\n\n")
			},
		},
		{
			name: "endpoint error",
			llm: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, `{"error":{"message":"model failed to load"}}`, http.StatusInternalServerError)
			},
		},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			esn := fmt.Sprintf("0dd1c6c%d", i)
			robot := startFakeRobot(t, esn)

			llm := httptest.NewServer(tc.llm)
			t.Cleanup(llm.Close)
			vars.UpdateConfig(func(c *vars.APIConfig) {
				c.Knowledge.Provider = "custom"
				c.Knowledge.Endpoint = llm.URL
				c.Knowledge.Model = "gemma4:e4b"
				c.Knowledge.Key = "key"
			})

			go StreamingKGSim(nil, esn, "what is the muppet show", true)

			select {
			case <-robot.released:
			case <-time.After(5 * time.Second):
				t.Fatal("behavior control was never released, the robot stays frozen")
			}
		})
	}
}
