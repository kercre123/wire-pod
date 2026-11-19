package mdnshandler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/zeroconf"
)

var (
	mdnsMu     sync.Mutex
	mdnsCancel context.CancelFunc
	mdnsNow    = make(chan struct{}, 1)
)

func postmDNSWhenNewVector(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	for ctx.Err() == nil {
		resolver, err := zeroconf.NewResolver(nil)
		if err != nil {
			logger.Println("mDNS resolver error: ", err)
			return
		}
		entries := make(chan *zeroconf.ServiceEntry)
		browseCtx, cancel := context.WithTimeout(ctx, time.Second*80)
		if err := resolver.Browse(browseCtx, "_ankivector._tcp", "local.", entries); err != nil {
			fmt.Println(err)
			cancel()
			return
		}
		for entry := range entries {
			if strings.Contains(entry.Service, "ankivector") {
				logger.Println("Vector discovered on network, broadcasting mDNS")
				cancel()
				time.Sleep(time.Second)
				PostmDNSNow()
				return
			}
		}
		cancel()
	}
}

func PostmDNSNow() {
	logger.Println("Broadcasting mDNS now (outside of timer loop)")
	select {
	case mdnsNow <- struct{}{}:
	default:
	}
}

func StopmDNS() {
	mdnsMu.Lock()
	defer mdnsMu.Unlock()
	if mdnsCancel != nil {
		mdnsCancel()
		mdnsCancel = nil
	}
}

func PostmDNS() {
	if os.Getenv("DISABLE_MDNS") == "true" {
		fmt.Println("mDNS is disabled")
		return
	}
	mdnsMu.Lock()
	if mdnsCancel != nil {
		mdnsMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	mdnsCancel = cancel
	mdnsMu.Unlock()

	go postmDNSWhenNewVector(ctx)
	logger.Println("Registering escapepod.local on network (loop)")
	for {
		ipAddr := vars.GetOutboundIP().String()
		server, err := zeroconf.RegisterProxy("escapepod", "_app-proto._tcp", "local.", 8084, "escapepod", []string{ipAddr}, []string{"txtv=0", "lo=1", "la=2"}, nil)
		if err != nil {
			logger.Println("mDNS register error: ", err)
		} else if os.Getenv("PRINT_MDNS") == "true" {
			logger.Println("mDNS broadcasted")
		}
		stopped := false
		select {
		case <-ctx.Done():
			stopped = true
		case <-mdnsNow:
		case <-time.After(30 * time.Second):
		}
		if server != nil {
			server.Shutdown()
		}
		if stopped {
			logger.Println("Stopped broadcasting escapepod.local")
			return
		}
		time.Sleep(time.Second / 3)
	}
}
