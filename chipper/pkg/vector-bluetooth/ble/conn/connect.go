package conn

import (
	"context"
	"math/rand"
	"time"

	"github.com/currantlabs/ble"
	"github.com/pkg/errors"
)

const (
	duration   = 2
	readUUID   = "7d2a4bda-d29b-4152-b725-2491478c5cd7"
	writeUUID  = "30619f2d-0f54-41bd-a65a-7588d8c85b45"
	retryCount = 3
	offset     = 1
)

// Connect connects to a specific device
func (c *Connection) Connect(id int) error {
	if err := c.bleConnect(id); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.discoverProfile); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.findReader); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.findWriter); err != nil {
		return err
	}

	errCh := make(chan error)

	go c.subscribe(errCh)
	err := <-errCh
	if err != nil {
		return err
	}
	c.established.Enable()

	go c.handleIncoming()
	return nil
}

// bleConnect handles establishing the actual connection
func (c *Connection) bleConnect(id int) error {
	ctx := ble.WithSigHandler(
		context.WithTimeout(
			context.Background(),
			scanDuration*duration,
		),
	)

	cln, err := c.device.Dial(
		ctx,
		c.scanresults.getresult(id),
	)
	if err != nil {
		return err
	}

	c.connection = cln

	return nil
}

// discoverPRofile finds the device profile and sets it
func (c *Connection) discoverProfile() error {
	p, err := c.connection.DiscoverProfile(false)
	if err != nil {
		return errors.Wrap(err, "can't discover profile")
	}
	c.profile = p
	return nil
}

// findWriter configures the writer
func (c *Connection) findWriter() error {
	wr := c.profile.Find(
		ble.NewCharacteristic(
			c.writeUUID(),
		),
	)
	if wr == nil {
		return errors.New("cannot find write channel")
	}
	c.writer = wr.(*ble.Characteristic)

	return nil
}

// findReader configures the reader
func (c *Connection) findReader() error {
	wr := c.profile.Find(
		ble.NewCharacteristic(
			c.readUUID(),
		),
	)
	if wr == nil {
		return errors.New("cannot find read channel")
	}
	c.reader = wr.(*ble.Characteristic)

	return nil
}

// subscribe pipes incoming data to a reader chan
func (c *Connection) subscribe(errChan chan error) {
	if err := c.connection.Subscribe(
		c.writer,
		true,
		func(req []byte) {
			c.incoming <- req
		},
	); err != nil {
		errChan <- err
	}
	errChan <- nil
}

func (c *Connection) handleIncoming() {
	blebuf := bleBuffer{}
	for {
		incoming := <-c.incoming
		b := blebuf.receiveRawBuffer(incoming)
		if b == nil {
			continue
		}
		switch {
		case !c.connected.Enabled():
			c.handleConnectionRequest(incoming)
		case !c.encrypted.Enabled() && c.connected.Enabled():
			c.out <- b
		case c.encrypted.Enabled() && c.connected.Enabled():
			buf, _ := c.crypto.DecryptMessage(b)
			// IDEA:  should this reset everything?
			c.out <- buf
		default:
			c.established.Disable()
			c.encrypted.Disable()
		}
	}
}

func (c *Connection) readUUID() ble.UUID {
	return ble.MustParse(readUUID)
}

func (c *Connection) writeUUID() ble.UUID {
	return ble.MustParse(writeUUID)
}

func (c *Connection) handleConnectionRequest(buffer []byte) {
	if err := c.rawMessage(buffer); err != nil {
		return
	}

	c.connected.Enable()
	c.version = int(buffer[2])
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		attempts--
		if attempts > 0 {
			//nolint -- we definitely don't need cryptographically secure numbers for jitter
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep += jitter / offset

			time.Sleep(sleep)
			return retry(attempts, offset*sleep, f)
		}
		return err
	}
	return nil
}
