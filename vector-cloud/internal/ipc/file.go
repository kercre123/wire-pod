package ipc

import (
	"errors"
	"net"
	"os"
)

// NewFileConn returns a connection that will operate on the
// given file descriptor
func NewFileConn(fd uintptr) (Conn, error) {
	file := os.NewFile(fd, "file client")
	if file == nil {
		return nil, errors.New("Could not create file from descriptor")
	}

	conn, err := net.FileConn(file)
	if err != nil {
		return nil, err
	}
	return newBaseConn(conn), nil
}
