package conn

import (
	"math/rand"
	"time"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"github.com/kercre123/chipper/pkg/vector-bluetooth/ble/blecrypto"
	"github.com/pkg/errors"
)

// Connection is the configuration struct for ble connections
type Connection struct {
	device      ble.Device
	scanresults *scan
	connection  ble.Client
	reader      *ble.Characteristic
	writer      *ble.Characteristic
	profile     *ble.Profile
	incoming    chan []byte
	out         chan []byte
	crypto      *blecrypto.BLECrypto
	version     int
	established lockState
	connected   lockState
	encrypted   lockState
}

// New returns a connection, or an error on failure
func New(output chan []byte) (*Connection, error) {
	rand.Seed(time.Now().UnixNano())

	c := Connection{
		scanresults: newScan(),
		incoming:    make(chan []byte),
		out:         output,
		crypto:      blecrypto.New(),
	}

	d, err := linux.NewDevice()
	if err != nil {
		return nil, errors.Wrap(err, "can't add new device")
	}
	c.device = d

	return &c, nil
}

// EnableEncryption sets the encryption bit to automatically de/encrypt
func (c *Connection) EnableEncryption() {
	c.encrypted.Enable()
}

// Connected lets external packages know if the initial connection attempt has happened
func (c *Connection) Connected() bool {
	return c.connected.Enabled()
}
