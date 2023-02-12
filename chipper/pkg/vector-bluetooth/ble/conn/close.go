package conn

// Close shuts down the BLE connection
func (c *Connection) Close() error {
	c.connected.Disable()
	c.encrypted.Disable()
	c.established.Disable()
	return c.device.Stop()
}

// Reset clears all connection information
func (c *Connection) Reset() {
	c.connected.Disable()
	c.encrypted.Disable()
	c.established.Disable()
}
