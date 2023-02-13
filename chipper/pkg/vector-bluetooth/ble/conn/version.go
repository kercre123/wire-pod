package conn

// Version replies with the RTS version
func (c *Connection) Version() int {
	return c.version
}
