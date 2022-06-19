package voice

import (
	"bytes"
	"strings"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

// Receiver is an object that should be passed in to the cloud process
// and determines how it will receive audio data from external sources
type Receiver struct {
	msg    chan *cloud.Message
	writer MsgSender
	isTest bool
}

func (r *Receiver) writeBack(msg *cloud.Message) error {
	return r.writer.Send(msg)
}

// bufToGoString converts byte buffers that may be null-terminated if created in C
// to Go strings by trimming off null chars
func bufToGoString(buf []byte) string {
	return strings.Trim(string(buf), "\x00")
}

// reads messages from a socket and places them on the channel
func socketReader(s ipc.Conn, msg chan *cloud.Message, kill <-chan struct{}) {
	for {
		if util.CanSelect(kill) {
			return
		}

		buf := s.ReadBlock()
		if buf != nil && len(buf) > 0 {
			var message cloud.Message
			if err := message.Unpack(bytes.NewBuffer(buf)); err != nil {
				log.Println("Error unpacking cloud message:", err)
				continue
			}
			msg <- &message
		}
	}
}

// NewIpcReceiver constructs a Receiver that receives audio data and hotword signals
// over the given IPC connection
func NewIpcReceiver(conn ipc.Conn, kill <-chan struct{}) *Receiver {
	msg := make(chan *cloud.Message)
	recv := &Receiver{msg: msg,
		writer: &ipcIO{&IPCMsgSender{conn}}}
	go socketReader(conn, msg, kill)

	return recv
}
