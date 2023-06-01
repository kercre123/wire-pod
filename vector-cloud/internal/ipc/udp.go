package ipc

import (
	"errors"
	"fmt"
	"net"
)

// NewUDPClient returns a new Socket that will attempt to connect to the server
// on the given IP and port
func NewUDPClient(ip string, port int) (Conn, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		return nil, errors.New(fmt.Sprint("Couldn't connect:", err))
	}

	return newDatagramClient(conn)
}

// NewUDPServer returns a new Server that will listen for clients on the given
// local port
func NewUDPServer(port int) (Server, error) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		return nil, err
	}

	return newDatagramServer(conn)
}
