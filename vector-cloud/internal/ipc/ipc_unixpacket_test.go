//+build !darwin

package ipc_test

import (
	"fmt"
	"testing"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
)

func TestUnixPacket(t *testing.T) {
	sfn := func() (ipc.Server, error) {
		return ipc.NewUnixPacketServer("unixpacketblah")
	}
	i := 0
	cfn := func() (ipc.Conn, error) {
		i++
		return ipc.NewUnixgramClient("unixpacketblah", fmt.Sprint("client", i))
	}
	testProtocol(t, sfn, cfn)
}
