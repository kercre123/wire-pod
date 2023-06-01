package voice

import (
	"bytes"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
)

type MsgSender interface {
	Send(*cloud.Message) error
}

type MsgReader interface {
	Read() (*cloud.Message, error)
}

// MsgIO defines the API for sending audio data from an external source into
// the cloud process (and potentially receiving info back)
type MsgIO interface {
	MsgSender
	MsgReader
}

type IPCMsgSender struct {
	Conn ipc.Conn
}

type ipcIO struct {
	*IPCMsgSender
}

func (s *IPCMsgSender) Send(msg *cloud.Message) error {
	var buf bytes.Buffer
	if err := msg.Pack(&buf); err != nil {
		return err
	}
	_, err := s.Conn.Write(buf.Bytes())
	return err
}

func (s *ipcIO) Read() (*cloud.Message, error) {
	buf := s.Conn.ReadBlock()
	var msg cloud.Message
	if err := msg.Unpack(bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewIpcIO returns a MsgIO that uses the given IPC connection to
// send audio data to the cloud process
func NewIpcIO(conn ipc.Conn) MsgIO {
	return &ipcIO{&IPCMsgSender{conn}}
}

type memIO struct {
	recv chan *cloud.Message
	dest *Receiver
}

func (s *memIO) Send(msg *cloud.Message) error {
	s.dest.msg <- msg
	return nil
}

func (s *memIO) Read() (*cloud.Message, error) {
	return <-s.recv, nil
}

type ChanMsgSender struct {
	Ch chan *cloud.Message
}

func (c *ChanMsgSender) Send(msg *cloud.Message) error {
	go func() {
		c.Ch <- msg
	}()
	return nil
}

// NewMemPipe returns a connected pair of a MsgIO and a Receiver that directly
// transmit data over channels; the Receiver should be passed in to the cloud
// process to get data from the MsgIO
func NewMemPipe() (MsgIO, *Receiver) {
	io := &memIO{make(chan *cloud.Message), nil}
	receiver := &Receiver{msg: make(chan *cloud.Message),
		writer: &ChanMsgSender{io.recv}}
	io.dest = receiver
	return io, receiver
}
