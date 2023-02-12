package blecrypto

import (
	"github.com/jamesruan/sodium"
	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// SetRemotePublicKey populates the remote public key
func (b *BLECrypto) SetRemotePublicKey(msg *rts.RtsConnRequest) error {
	b.remotePublicKey = sodium.KXPublicKey{
		Bytes: msg.PublicKey[:],
	}
	return nil
}

// GetRemotePublicKey returns the public key
func (b *BLECrypto) GetRemotePublicKey() [32]uint8 {
	pkb := [32]uint8{}
	for k, v := range b.keys.PublicKey.Bytes {
		pkb[k] = v
	}
	return pkb
}
