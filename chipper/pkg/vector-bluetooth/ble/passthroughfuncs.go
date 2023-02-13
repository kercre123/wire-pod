package ble

// Connected returns true if the device is connected
func (v *VectorBLE) Connected() bool {
	return v.ble.Connected()
}

// Close stops the BLE connection
func (v *VectorBLE) Close() error {
	return v.ble.Close()
}
