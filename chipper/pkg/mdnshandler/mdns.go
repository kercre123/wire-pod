package mdnshandler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
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
	for {
		resolver, _ := zeroconf.NewResolver(nil)
		entries := make(chan *zeroconf.ServiceEntry)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*130)
		err := resolver.Browse(ctx, "_ankivector._tcp", "local.", entries)
		if err != nil {
			fmt.Println(err)
			cancel()
			return
		}
		for entry := range entries {
			if strings.Contains(entry.Service, "ankivector") {
				PostmDNSNow()
			}
		}
		cancel()
	}

}

func PostmDNSNow() {
	select {
	case MDNSNow <- true:
	default:
	}
}

func PostmDNS() {
	if PostingmDNS {
		return
	}
	go PostmDNSWhenNewVector()
	MDNSNow = make(chan bool)
	go func() {
		for range MDNSNow {
			MDNSTimeBeforeNextRegister = 60
		}
	}()
	PostingmDNS = true
	logger.Println("Registering escapepod.local on network (loop)")
	for {
		ipAddr := botsetup.GetOutboundIP().String()
		server, _ := zeroconf.RegisterProxy("escapepod", "_escapepod._tcp", "local.", 8084, "escapepod", []string{ipAddr}, []string{"txtv=0", "lo=1", "la=2"}, nil)
		for {
			if MDNSTimeBeforeNextRegister >= 60 {
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
