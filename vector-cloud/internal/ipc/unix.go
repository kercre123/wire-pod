package ipc

import (
	"fmt"
	"net"
	"syscall"
)

// NewUnixServer returns a new server object listening for clients on the specified path,
// if no errors are encountered
func NewUnixServer(path string) (Server, error) {
	if []byte(path)[0] != '\x00' {
		syscall.Unlink(path)
	}
	listen, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	return newBaseServer(&listenerWrapper{&streamListener{listen}})
}

// NewUnixClient returns a new connection to the server at the specified path, assuming
// there are no errors when connecting
func NewUnixClient(path string) (Conn, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		fmt.Println("Dial error:", err)
		return nil, err
	}
	client := newBaseConn(newStreamConnection(conn))
	return client, nil
}
