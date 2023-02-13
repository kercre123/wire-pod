package conn

const (
	maxSize = 20
)

// Send sends a message via BLE
func (c *Connection) Send(buf []byte) error {
	if c.encrypted.Enabled() {
		var err error
		buf, err = c.crypto.Encrypt(buf)
		if err != nil {
			return err
		}
	}

	size := len(buf)
	if size < maxSize {
		return c.messageWithHeader(msgSolo, buf, size)
	}

	remaining := len(buf)

	for remaining > 0 {
		offset := size - remaining

		switch {
		case remaining == size:
			msgSize := maxSize - 1
			if err := c.messageWithHeader(msgStart, buf[offset:offset+msgSize], msgSize); err != nil {
				return err
			}
			remaining -= msgSize

		case remaining < maxSize:
			if err := c.messageWithHeader(msgEnd, buf[offset:], remaining); err != nil {
				return err
			}
			remaining = 0

		default:
			msgSize := maxSize - 1
			if err := c.messageWithHeader(msgContinue, buf[offset:offset+msgSize], msgSize); err != nil {
				return err
			}
			remaining -= msgSize
		}
	}
	return nil
}

func (c *Connection) messageWithHeader(multipart byte, buffer []byte, size int) error {
	var msg []byte
	msg = append(msg, getHeaderByte(multipart, size))
	msg = append(msg, buffer...)

	return c.rawMessage(msg)
}

func (c *Connection) rawMessage(buffer []byte) error {
	return c.connection.WriteCharacteristic(
		c.reader,
		buffer,
		true,
	)
}

func getHeaderByte(multipart byte, size int) byte {
	//nolint
	return byte(((int(multipart) << 6) | (size & -193)))
}
