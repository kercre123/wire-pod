package ipc_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverFn func() (ipc.Server, error)
type clientFn func() (ipc.Conn, error)

func testProtocol(t *testing.T, sfn serverFn, cfn clientFn) {
	assert := assert.New(t)
	require := require.New(t)
	serv, err := sfn()
	require.Nil(err, "server create err:", err)
	defer serv.Close()
	time.Sleep(5 * time.Millisecond)

	c1, err := cfn()
	require.Nil(err, "client create err:", err)
	defer c1.Close()
	time.Sleep(5 * time.Millisecond)
	s1 := <-serv.NewConns()
	defer s1.Close()
	c2, err := cfn()
	require.Nil(err, "client create err:", err)
	defer c2.Close()
	s2 := <-serv.NewConns()
	defer s2.Close()

	bufsizes := []int{1, 2, 4, 7, 16, 31, 64, 127, 256, 511, 1024, 1535}

	rand.Seed(1328461)

	for _, val := range bufsizes {
		buf := make([]byte, val)
		rand.Read(buf)
		c1.Write(buf)
		s2.Write(buf)
		assert.Empty(s2.Read())
		assert.Empty(c1.Read())
		time.Sleep(5 * time.Millisecond)
		b1 := s1.Read()
		b2 := c2.Read()
		assert.ElementsMatch(b1, b2)
		assert.ElementsMatch(b1, buf)

		rand.Read(buf)
		s1.Write(buf)
		s1.Write(buf)
		c2.Write(buf)

		b1 = c1.ReadBlock()
		b3 := c1.ReadBlock()
		b2 = s2.ReadBlock()
		assert.ElementsMatch(b1, b2)
		assert.ElementsMatch(b1, b3)
		assert.ElementsMatch(b1, buf)
	}
}

func TestUDP(t *testing.T) {
	// get open port
	s, err := net.ListenPacket("udp", "localhost:0")
	require.Nil(t, err)
	port := (s.LocalAddr().(*net.UDPAddr)).Port
	s.Close()
	sfn := func() (ipc.Server, error) {
		return ipc.NewUDPServer(port)
	}
	cfn := func() (ipc.Conn, error) {
		return ipc.NewUDPClient("127.0.0.1", port)
	}
	testProtocol(t, sfn, cfn)
}

func TestUnixgram(t *testing.T) {
	sfn := func() (ipc.Server, error) {
		return ipc.NewUnixgramServer("unixgramblah")
	}
	i := 0
	cfn := func() (ipc.Conn, error) {
		i++
		return ipc.NewUnixgramClient("unixgramblah", fmt.Sprint("client", i))
	}
	testProtocol(t, sfn, cfn)
}

func TestUnix(t *testing.T) {
	sfn := func() (ipc.Server, error) {
		return ipc.NewUnixServer("unixblah")
	}
	cfn := func() (ipc.Conn, error) {
		return ipc.NewUnixClient("unixblah")
	}
	testProtocol(t, sfn, cfn)
}
