package conn

const (
	sizeBits = 0b00111111

	msgStart    = 0b10
	msgContinue = 0b00
	msgEnd      = 0b01
	msgSolo     = 0b11
	// msgBits     = 0b11 << 6
)

type bleBuffer struct {
	Buf   []byte `json:"buf,omitempty"`
	State uint   `json:"state,omitempty"`
}

func (b *bleBuffer) receiveRawBuffer(buf []byte) []byte {
	headerByte := buf[0]
	sizeByte := getSize(headerByte)
	multipartState := getMultipartBits(headerByte)

	if int(sizeByte) != len(buf)-1 {
		return nil
	}

	switch multipartState {
	case msgStart:
		b.Buf = []byte{}
		b.append(buf, int(sizeByte))
		b.State = msgContinue

	case msgContinue:
		b.append(buf, int(sizeByte))
		b.State = msgContinue

	case msgEnd:
		b.append(buf, int(sizeByte))
		b.State = msgStart
		t := b.Buf
		b.Buf = []byte{}
		return t

	case msgSolo:
		b.append(buf, int(sizeByte))
		b.State = msgStart
		t := b.Buf
		b.Buf = []byte{}
		return t
	}

	return nil
}

func (b *bleBuffer) append(buf []byte, size int) {
	b.Buf = append(b.Buf, buf[1:]...)
}

func getSize(msgSize byte) byte {
	return msgSize & sizeBits
}

func getMultipartBits(msgSize byte) byte {
	//nolint
	return msgSize >> 6
}
