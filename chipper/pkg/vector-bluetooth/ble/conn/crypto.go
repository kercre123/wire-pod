package conn

import "github.com/kercre123/chipper/pkg/vector-bluetooth/rts"

// SetRemotePublicKey sets the remote public key
func (c *Connection) SetRemotePublicKey(msg *rts.RtsConnRequest) error {
	return c.crypto.SetRemotePublicKey(msg)
}

// GetRemotePublicKey gets the remote public key
func (c *Connection) GetRemotePublicKey() [32]uint8 {
	return c.crypto.GetRemotePublicKey()
}

// SetNonces sets the nonces
func (c *Connection) SetNonces(msg *rts.RtsNonceMessage) error {
	return c.crypto.SetNonces(msg)
}

// SetPin sets the pin
func (c *Connection) SetPin(msg string) error {
	return c.crypto.SetPin(msg)
}
