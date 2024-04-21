package mdnshandler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/zeroconf"
)

/*
	goal:
		- register a proxy every minute or so
		(it is every 4 seconds right now, which is too much)
		- BroadcastNow function which resets the timer and broadcasts right now
		- detect when a new bot is available on the network and broadcast ASAP
*/

var PostingmDNS bool
var MDNSNow chan bool
var MDNSTimeBeforeNextRegister float32

func PostmDNSWhenNewVector() {
	time.Sleep(time.Second * 5)
	for {
		resolver, _ := zeroconf.NewResolver(nil)
		entries := make(chan *zeroconf.ServiceEntry)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*80)
		err := resolver.Browse(ctx, "_ankivector._tcp", "local.", entries)
		if err != nil {
			fmt.Println(err)
			cancel()
			return
		}
		for entry := range entries {
			if strings.Contains(entry.Service, "ankivector") {
				logger.Println("Vector discovered on network, broadcasting mDNS")
				time.Sleep(time.Second * 2)
				PostmDNSNow()
				cancel()
				return
			}
		}
		cancel()
	}

}

func PostmDNSNow() {
	logger.Println("Broadcasting mDNS now (outside of timer loop)")
	select {
	case MDNSNow <- true:
	default:
	}
}

func PostmDNS() {
	if os.Getenv("DISABLE_MDNS") == "true" {
		fmt.Println("mDNS is disabled")
		return
	}
	if PostingmDNS {
		return
	}
	go PostmDNSWhenNewVector()
	MDNSNow = make(chan bool)
	go func() {
		for range MDNSNow {
			MDNSTimeBeforeNextRegister = 30
		}
	}()
	PostingmDNS = true
	logger.Println("Registering escapepod.local on network (loop)")
	for {
		ipAddr := vars.GetOutboundIP().String()
		server, _ := zeroconf.RegisterProxy("escapepod", "_app-proto._tcp", "local.", 8084, "escapepod", []string{ipAddr}, []string{"txtv=0", "lo=1", "la=2"}, nil)
		for {
			if MDNSTimeBeforeNextRegister >= 30 {
				MDNSTimeBeforeNextRegister = 0
				break
			}
			MDNSTimeBeforeNextRegister = MDNSTimeBeforeNextRegister + (float32(1) / float32(4))
			time.Sleep(time.Second / 4)
		}
		server.Shutdown()
		server = nil
		time.Sleep(time.Second / 3)
	}
}
