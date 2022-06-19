package ipc

import (
	"net"
	"syscall"
)

// NewUnixgramClient returns a new connection to the server at the specified path, assuming
// there are no errors when connecting
func NewUnixgramClient(path string, name string) (Conn, error) {
	servaddr, err := net.ResolveUnixAddr("unixgram", path)
	if err != nil {
		return nil, err
	}
	clientpath := path + "_" + name
	syscall.Unlink(clientpath)
	cliaddr, err := net.ResolveUnixAddr("unixgram", clientpath)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unixgram", cliaddr, servaddr)
	if err != nil {
		return nil, err
	}

	// Set receive buffer size to 64k (defaults to 4k)
	conn.SetReadBuffer(64 * 1024)

	return newDatagramClient(conn)
}

// NewUnixgramServer returns a new server object listening for clients on the specified path,
// if no errors are encountered
func NewUnixgramServer(path string) (Server, error) {
	if []byte(path)[0] != '\x00' {
		syscall.Unlink(path)
	}
	conn, err := net.ListenPacket("unixgram", path)
	if err != nil {
		return nil, err
	}

	return newDatagramServer(conn)
}
