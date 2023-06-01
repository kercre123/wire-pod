package stream

import (
	"errors"

	"github.com/digital-dream-labs/vector-cloud/internal/util"

	"github.com/digital-dream-labs/api-clients/chipper"
)

type Conn interface {
	Close() error
	CloseSend() error
	SendAudio([]byte) error
	WaitForResponse() (interface{}, error)
}

type chipperConn struct {
	conn   *chipper.Conn
	stream chipper.Stream
}

func (c *chipperConn) Close() error {
	var err util.Errors
	if c.stream != nil {
		err.Append(c.stream.Close())
	}
	if c.conn != nil {
		err.Append(c.conn.Close())
	}
	return err.Error()
}

func (c *chipperConn) CloseSend() error {
	if c.stream != nil {
		return c.stream.CloseSend()
	}
	return errors.New("cannot CloseSend on nil stream")
}

func (c *chipperConn) SendAudio(samples []byte) error {
	return c.stream.SendAudio(samples)
}

func (c *chipperConn) WaitForResponse() (interface{}, error) {
	return c.stream.WaitForResponse()
}
