package jdocs

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
)

func runServer(ctx context.Context, opts *options) {
	socketName := "jdocs_server"
	if opts.socketNameSuffix != "" {
		socketName = fmt.Sprintf("%s_%s", socketName, opts.socketNameSuffix)
	}

	serv, err := ipc.NewUnixgramServer(ipc.GetSocketPath(socketName))
	if err != nil {
		log.Println("Error creating jdocs server:", err)
		return
	}

	// close on context?
	for c := range serv.NewConns() {
		cl := client{Conn: c, opts: opts}
		go cl.handleConn(ctx)
	}
}

type client struct {
	ipc.Conn
	opts     *options
	reqMutex sync.Mutex
}

func (c *client) handleConn(ctx context.Context) {
	for {
		buf := c.ReadBlock()
		// TODO: will this ever close?
		if buf == nil || len(buf) == 0 {
			return
		}
		var msg cloud.DocRequest
		if err := msg.Unpack(bytes.NewBuffer(buf)); err != nil {
			log.Println("Could not unpack jdocs request:", err)
			continue
		}

		resp, err := c.handleRequest(ctx, &msg)
		if err != nil {
			log.Println("Error handling jdocs request:", err)
			if c.opts.errListener != nil {
				c.opts.errListener.OnError(err)
			}
		}
		if resp != nil {
			var buf bytes.Buffer
			if err := resp.Pack(&buf); err != nil {
				log.Println("Error packing jdocs response:", err)
			} else if n, err := c.Write(buf.Bytes()); n != buf.Len() || err != nil {
				log.Println("Error sending jdocs response:", fmt.Sprintf("%d/%d,", n, buf.Len()), err)
			}
		}
	}
}

func (c *client) handleRequest(ctx context.Context, msg *cloud.DocRequest) (*cloud.DocResponse, error) {
	c.reqMutex.Lock()
	defer c.reqMutex.Unlock()
	if ok, resp, err := c.handleConnectionless(msg); ok {
		return resp, err
	}
	conn, err := newConn(ctx, c.opts)
	if err != nil {
		return connectErrorResponse, err
	}
	defer conn.close()
	return conn.handleRequest(ctx, msg)
}
