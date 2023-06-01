package ipc

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

// for the UDP server, we need to construct a translation from
// the net.PacketConn returned by net.ListenPacket("udp", ...)
// to a net.Listener with semantics to return new connections
type packetListener struct {
	conn       net.PacketConn
	clients    map[string]*packetClient
	newClients chan *packetClient
	kill       chan struct{}
	closer     *util.DoOnce
}

// Common string to indicate connection handshake shared by C++ code
// Defined in LocalUdpServer.cpp/UdpServer.cpp
const connPacket = "ANKICONN"

func (p *packetListener) Accept() (Conn, error) {
	select {
	case conn := <-p.newClients:
		return conn, nil
	case <-p.kill:
		return nil, io.EOF
	}
}

func (p *packetListener) Close() error {
	p.closer.Do()
	return p.conn.Close()
}

type packetClient struct {
	conn net.PacketConn
	addr net.Addr
	read chan []byte
	kill chan struct{}
}

func (c *packetClient) Read() []byte {
	select {
	case <-c.kill:
		return nil
	case recv := <-c.read:
		return recv
	default:
		return []byte{}
	}
}

func (c *packetClient) ReadBlock() []byte {
	select {
	case <-c.kill:
		return nil
	case recv := <-c.read:
		return recv
	}
}

func (c *packetClient) Close() error {
	// packetClient is closed by its server being closed; no-op
	return nil
}

func (c *packetClient) Write(buf []byte) (int, error) {
	return c.conn.WriteTo(buf, c.addr)
}

func newDatagramClient(conn net.Conn) (Conn, error) {
	client := newBaseConn(conn)

	packet := []byte(connPacket)
	n, err := conn.Write(packet)
	if err != nil {
		return nil, err
	} else if n != len(connPacket) {
		return nil, errors.New(fmt.Sprint("unexpected write size: ", n))
	}

	return client, nil
}

func newDatagramServer(conn net.PacketConn) (Server, error) {
	kill := make(chan struct{})
	closer := util.NewDoOnce(func() { close(kill) })
	newClients := make(chan *packetClient)
	listener := &packetListener{conn, make(map[string]*packetClient), newClients, kill, &closer}

	// start reader thread
	go func() {
		defer closer.Do()
		buf := make([]byte, 32768)
		for {
			if util.CanSelect(kill) {
				return
			}
			n, address, err := conn.ReadFrom(buf)
			if err != nil || address == nil {
				return
			}
			if n == len(buf) {
				fmt.Println("Maxed buffer")
			}

			isConnPacket := n == len(connPacket) && string(buf[:n]) == connPacket
			addr := address.String()
			if client, ok := listener.clients[addr]; ok {
				if isConnPacket {
					// ignore; they're trying to reconnect
				} else {
					sendbuf := make([]byte, n)
					copy(sendbuf, buf)
					select {
					case client.read <- sendbuf:
					case <-kill:
						return
					}
				}
			} else if isConnPacket {
				client = &packetClient{conn, address, make(chan []byte), kill}
				listener.clients[addr] = client
				select {
				case newClients <- client:
				case <-kill:
					return
				}
			} else {
				fmt.Println("unexpected buf size from unknown client:", n, addr)
			}
		}
	}()

	return newBaseServer(listener)
}
