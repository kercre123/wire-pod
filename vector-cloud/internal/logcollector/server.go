package logcollector

import (
	"bytes"
	"context"
	"fmt"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
)

func runServer(ctx context.Context, opts *options) {
	socketName := "logcollector_server"
	if opts.socketNameSuffix != "" {
		socketName = fmt.Sprintf("%s_%s", socketName, opts.socketNameSuffix)
	}

	serv, err := ipc.NewUnixgramServer(ipc.GetSocketPath(socketName))
	if err != nil {
		log.Println("Error creating log collector server:", err)
		return
	}

	// close on context?
	for c := range serv.NewConns() {
		cl := client{c, opts}
		go cl.handleConn(ctx)
	}
}

type client struct {
	ipc.Conn
	opts *options
}

func (c *client) handleConn(ctx context.Context) {
	for {
		buf := c.ReadBlock()
		// TODO: will this ever close?
		if buf == nil || len(buf) == 0 {
			return
		}
		var msg cloud.LogCollectorRequest
		if err := msg.Unpack(bytes.NewBuffer(buf)); err != nil {
			log.Println("Could not unpack log collector request:", err)
			continue
		}

		resp, err := c.handleRequest(ctx, &msg)
		if err != nil {
			log.Println("Error handling log collector request:", err)
		}
		if resp != nil {
			var buf bytes.Buffer
			if err := resp.Pack(&buf); err != nil {
				log.Println("Error packing log collector response:", err)
			} else if n, err := c.Write(buf.Bytes()); n != buf.Len() || err != nil {
				log.Println("Error sending log collector response:", fmt.Sprintf("%d/%d,", n, buf.Len()), err)
			}
		}
	}
}

func (c *client) handleRequest(ctx context.Context, msg *cloud.LogCollectorRequest) (*cloud.LogCollectorResponse, error) {
	cladHandler, err := newCladHandler(c.opts)
	if err != nil {
		return connectErrorResponse, err
		if c.opts.errListener != nil {
			c.opts.errListener.OnError(err)
		}
	}
	return cladHandler.handleRequest(ctx, msg)
}

// Run starts the log collector service
func Run(ctx context.Context, optionValues ...Option) {
	var opts options
	for _, o := range optionValues {
		o(&opts)
	}

	if opts.server {
		runServer(ctx, &opts)
	}
}
