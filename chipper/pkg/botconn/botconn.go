package botconn

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/digital-dream-labs/hugh/grpc/client"
	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"google.golang.org/grpc"
)

type Info struct {
	ESN    string
	Target string
	GUID   string
}

type conn struct {
	info   Info
	vec    *vector.Vector
	client *client.Client
}

var (
	mu    sync.Mutex
	conns = map[string]*conn{}
)

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}

func Lookup(esn string) (Info, error) {
	esn = strings.TrimSpace(esn)
	botInfo := vars.GetBotInfo()
	for _, robot := range botInfo.Robots {
		if strings.EqualFold(esn, robot.Esn) {
			guid := robot.GUID
			if guid == "" {
				guid = botInfo.GlobalGUID
			}
			return Info{
				ESN:    robot.Esn,
				Target: robot.IPAddress + ":443",
				GUID:   guid,
			}, nil
		}
	}
	return Info{}, fmt.Errorf("robot %s not found in bot info", esn)
}

func dial(info Info) (*vector.Vector, *client.Client, error) {
	c, err := client.New(
		client.WithTarget(info.Target),
		client.WithInsecureSkipVerify(),
		client.WithDialopts(grpc.WithPerRPCCredentials(tokenAuth{token: info.GUID})),
	)
	if err != nil {
		return nil, nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, nil, err
	}
	vec := &vector.Vector{Conn: vectorpb.NewExternalInterfaceClient(c.Conn())}
	for _, opt := range []vector.Option{
		vector.WithSerialNo(info.ESN),
		vector.WithTarget(info.Target),
		vector.WithToken(info.GUID),
	} {
		opt(&vec.Cfg)
	}
	return vec, c, nil
}

// shared, don't close it
func Get(esn string) (*vector.Vector, error) {
	info, err := Lookup(esn)
	if err != nil {
		return nil, err
	}
	key := strings.ToLower(info.ESN)
	mu.Lock()
	defer mu.Unlock()
	if c, ok := conns[key]; ok {
		if c.info == info {
			return c.vec, nil
		}
		c.client.Close()
		delete(conns, key)
	}
	vec, cl, err := dial(info)
	if err != nil {
		return nil, err
	}
	conns[key] = &conn{info: info, vec: vec, client: cl}
	return vec, nil
}

func DialWithGUID(esn, guid string) (*vector.Vector, func(), error) {
	info, err := Lookup(esn)
	if err != nil {
		return nil, nil, err
	}
	info.GUID = guid
	vec, cl, err := dial(info)
	if err != nil {
		return nil, nil, err
	}
	return vec, func() { cl.Close() }, nil
}
