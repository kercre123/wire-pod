package dev

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/config"
)

type connection struct {
	stamp time.Time
	ms    int
	err   error
}

// an hour's worth of connections, every 5 sec
var connData struct {
	m       sync.Mutex
	storage [720]connection
	head    int
	count   int
}

const connInterval = 5 * time.Second

func initConnect() {
	go connectRoutine()
}

func connectRoutine() {
	i := 0
	circled := false
	tick := time.NewTicker(connInterval)
	for {
		ts := <-tick.C
		dest := &connData.storage[i]
		i = i + 1
		if i >= len(connData.storage) {
			circled = true
			i = 0
		}
		if circled {
			connData.m.Lock()
			connData.head = (i + 1) % len(connData.storage)
			connData.count--
			connData.m.Unlock()
		}
		go func() {
			// try to connect to OTA server, log how long it took
			dest.stamp = ts
			ctx, cancel := context.WithTimeout(context.Background(), connInterval)
			// construct a HTTP URL from OTA address (something like `ota-cdn.anki.com:443`)
			otaURL := "http://" + config.Env.Check
			if req, err := http.NewRequest("HEAD", otaURL, nil); err != nil {
				dest.ms = -1
				dest.err = err
			} else if resp, err := http.DefaultClient.Do(req.WithContext(ctx)); err != nil {
				dest.ms = -1
				dest.err = err
			} else {
				dest.ms = int(time.Now().Sub(ts) / time.Millisecond)
				dest.err = nil
				resp.Body.Close()
			}
			cancel()
			connData.m.Lock()
			connData.count++
			connData.m.Unlock()
		}()
	}
}

func connectHandler(rw http.ResponseWriter, r *http.Request) {
	connData.m.Lock()
	head := connData.head
	count := connData.count
	connData.m.Unlock()

	tf := func(t time.Time) string {
		return t.Format("15:04:05")
	}
	successStringer := func(arr []*connection) string {
		if len(arr) == 0 {
			return ""
		}
		if len(arr) == 1 {
			return fmt.Sprintf("<li>%s: success (%d ms)</li>\n", tf(arr[0].stamp), arr[0].ms)
		}
		min := math.MaxInt32
		max := 0
		var avg int
		for _, c := range arr {
			if c.ms < min {
				min = c.ms
			}
			if c.ms > max {
				max = c.ms
			}
			avg += c.ms
		}
		avg = avg / len(arr)
		return fmt.Sprintf("<li>%s-%s: %d successes (min time %d ms, max %d, avg %d)</li>\n",
			tf(arr[0].stamp), tf(arr[len(arr)-1].stamp), len(arr), min, max, avg)
	}
	errorStringer := func(c *connection) string {
		return fmt.Sprintf("<li>%s: ERROR: %s</li>\n", tf(c.stamp), c.err)
	}

	var b strings.Builder
	var consecutive []*connection
	for i := 0; i < count; i++ {
		src := &connData.storage[(head+i)%len(connData.storage)]
		if src.err == nil {
			consecutive = append(consecutive, src)
		} else {
			b.WriteString(successStringer(consecutive))
			consecutive = consecutive[:0]
			b.WriteString(errorStringer(src))
		}
	}
	b.WriteString(successStringer(consecutive))
	fmt.Fprintf(rw, "<html><h2>Results for last %d connections</h2>", count)
	fmt.Fprintf(rw, "\n<ul>")
	fmt.Fprintf(rw, b.String())
	fmt.Fprintf(rw, "\n<ul></html>")
}
