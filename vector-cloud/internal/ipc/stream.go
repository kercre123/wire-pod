package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type streamConn struct {
	net.Conn
}

func (c *streamConn) Read(b []byte) (int, error) {
	var buflen uint32
	err := binary.Read(c.Conn, binary.BigEndian, &buflen)
	if err != nil {
		return 0, err
	}
	if buflen > uint32(len(b)) {
		fmt.Println("buffer overflow, have size", len(b), "but need", buflen)
		buflen = uint32(len(b))
	}
	return c.Conn.Read(b[:buflen])
}

func (c *streamConn) Write(b []byte) (int, error) {
	lenbuf := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(lenbuf, binary.BigEndian, uint32(len(b)))
	_, err := c.Conn.Write(lenbuf.Bytes())
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

type streamListener struct {
	net.Listener
}

func (s *streamListener) Accept() (net.Conn, error) {
	c, err := s.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return newStreamConnection(c), nil
}

func newStreamConnection(conn net.Conn) net.Conn {
	return &streamConn{conn}
}
