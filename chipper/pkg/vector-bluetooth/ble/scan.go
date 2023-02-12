package ble

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

// Scan shows a list of connectable vectors
func (v *VectorBLE) Scan() (*ScanResponse, error) {
	r, err := v.ble.Scan()
	if err != nil {
		return nil, err
	}

	d := []*Device{}
	for _, v := range r.Devices {
		d = append(
			d,
			&Device{
				ID:      v.ID,
				Name:    v.Name,
				Address: v.Address,
			},
		)
	}

	resp := ScanResponse{
		Devices: d,
	}

	return &resp, nil
}
