package conn

import (
	"context"
	"time"

	"github.com/currantlabs/ble"
)

const (
	scanDuration = 3 * time.Second
)

// ScanResponse is a list of devices the BLE adaptor has found
type ScanResponse struct {
	Devices []*Device
}

// Device is a single device entity
type Device struct {
	ID      int
	Name    string
	Address string
}

// Scan looks for BLE devices matching the vector requirements
func (c *Connection) Scan() (*ScanResponse, error) {
	ctx := ble.WithSigHandler(
		context.WithTimeout(
			context.Background(),
			scanDuration,
		),
	)

	// This error is intentionally ignored.  If you were to do something with it,
	// you'd get a deadline exceeded message every time.
	_ = c.device.Scan(
		ctx,
		false,
		c.scan,
	)

	d := []*Device{}

	for k, v := range c.scanresults.getresults() {
		td := Device{
			ID:      k,
			Name:    v.name,
			Address: v.addr.String(),
		}
		d = append(d, &td)
	}

	resp := ScanResponse{
		Devices: d,
	}

	return &resp, nil
}
