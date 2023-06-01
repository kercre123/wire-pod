package ipc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

// baseConn wraps a net.Conn to manage the resources we associate with one (reader routines, channels)
type baseConn struct {
	conn       io.ReadWriteCloser
	kill       chan struct{}
	read       chan []byte
	killCloser util.DoOnce
	wg         sync.WaitGroup
}

func (c *baseConn) Close() error {
	c.killCloser.Do()
	ret := c.conn.Close()
	c.wg.Wait() // wait for reader routine to close
	close(c.read)
	return ret
}

func (c *baseConn) Read() (b []byte) {
	select {
	case b = <-c.read:
	case <-c.kill:
	default:
		b = []byte{}
	}
	return
}

func (c *baseConn) ReadBlock() (b []byte) {
	select {
	case b = <-c.read:
	case <-c.kill:
	}
	return
}

func (c *baseConn) Write(b []byte) (int, error) {
	if util.CanSelect(c.kill) {
		return 0, errors.New("connection has been closed")
	}
	return c.conn.Write(b)
}

func newBaseConn(conn io.ReadWriteCloser) Conn {
	kill := make(chan struct{})
	ret := &baseConn{
		conn:       conn,
		kill:       kill,
		read:       make(chan []byte),
		killCloser: util.NewDoOnce(func() { close(kill) })}

	// reader thread: puts incoming messages into read channel, waiting to be consumed
	ret.wg.Add(1)
	go func() {
		defer ret.wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if !util.CanSelect(ret.kill) {
					isTimeout := false
					if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
						isTimeout = true
					}
					if err == io.EOF || isTimeout {
						// unexpected EOF = server probably died
					} else {
						fmt.Println("Socket couldn't read length:", err)
					}
					ret.killCloser.Do()
				}
				break
			}
			if n == 0 {
				continue
			}
			sendbuf := make([]byte, n)
			copy(sendbuf, buf[:n])
			select {
			case ret.read <- sendbuf:
			case <-ret.kill:
			}
		}
	}()

	return ret
}
