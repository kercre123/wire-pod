/*
Package ipc provides a basic, dumb implementation of discrete messaging
over various underlying transport protocols. Supported protocols are
three kinds of Unix domain sockets (streaming, sequential packet,
datagram) and local UDP sockets. For stream sockets, messages are
prefixed with a 4 byte size so that the connection knows how many bytes
to pull off for each message, preserving message boundaries.

The Conn interface represents a connection between two endpoints, and
is returned by all functions that create clients; the Server interface
represents a running server that spawns new Conns as clients connect to
it. As a Server receives new connections, it will push new clients onto
the channel accessible by Server.NewConns(). Allowing message routing
between named clients is left to the higher-level multi package.
*/
package ipc

// Conn represents a connection between two specific network endpoints - a client
// will hold one Conn to the server it's connected to, a server will hold individual
// Conns for each client that's connected to it.
type Conn interface {
	// Read a message - returns an empty slice if nothing is available, nil if
	// the connection has been shut down
	Read() []byte

	// Read a message and block until one becomes available (returns the buffer)
	// or the connection is closed (returns nil)
	ReadBlock() []byte

	// Write the given buffer to the other end
	Write([]byte) (int, error)

	// Close the connection and associated resources
	Close() error
}

// Server represents a running server waiting for connections. Client connections
// can be accessed by pulling new clients off the channel returned by NewConns().
type Server interface {
	NewConns() <-chan Conn
	Close() error
}
