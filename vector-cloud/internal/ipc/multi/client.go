package multi

import (
	"errors"
	"strings"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
)

type Client interface {
	Receive() (from string, buf []byte, err error)
	ReceiveBlock() (from string, buf []byte, err error)
	Send(dest string, buf []byte) (int, error)
	Close() error
}

type clientImpl struct {
	conn ipc.Conn
	name string
}

func receiveInternal(receiveFunc func() []byte) (from string, buf []byte, err error) {
	buf = receiveFunc()
	if buf == nil || len(buf) == 0 {
		return "", nil, nil
	}
	nullIdx := strings.Index(string(buf), "\x00")
	if nullIdx <= 0 {
		return "", nil, errors.New("no null identifier in buf")
	}
	from = string(buf[:nullIdx])
	buf = buf[nullIdx+1:]
	return
}

func (c *clientImpl) Receive() (from string, buf []byte, err error) {
	return receiveInternal(func() []byte {
		return c.conn.Read()
	})
}

func (c *clientImpl) ReceiveBlock() (from string, buf []byte, err error) {
	return receiveInternal(func() []byte {
		return c.conn.ReadBlock()
	})
}

func (c *clientImpl) Send(dest string, buf []byte) (int, error) {
	sendbuf := getBufferForMessage(dest, buf)
	sizediff := len(sendbuf) - len(buf)
	n, err := c.conn.Write(sendbuf)
	if n >= sizediff {
		n -= sizediff
	}
	return n, err
}

func (c *clientImpl) Close() error {
	return c.conn.Close()
}

func NewClient(conn ipc.Conn, clientName string) (Client, error) {
	client := &clientImpl{conn, clientName}

	client.conn.Write([]byte(clientName))

	return client, nil
}
