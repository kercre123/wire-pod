//+build !darwin

package ipc

import (
	"fmt"
	"net"
	"syscall"
)

// NewUnixPacketServer returns a new connection to the server at the specified path, assuming
// there are no errors when connecting
func NewUnixPacketServer(path string) (Server, error) {
	if []byte(path)[0] != '\x00' {
		syscall.Unlink(path)
	}
	listen, err := net.Listen("unixpacket", path)
	if err != nil {
		return nil, err
	}
	return newBaseServer(&listenerWrapper{listen})
}

// NewUnixPacketClient returns a new server object listening for clients on the specified path,
// if no errors are encountered
func NewUnixPacketClient(path string) (Conn, error) {
	conn, err := net.Dial("unixpacket", path)
	if err != nil {
		fmt.Println("Dial error:", err)
		return nil, err
	}
	client := newBaseConn(conn)
	return client, nil
}
