package blecrypto

import (
	"github.com/jamesruan/sodium"
)

// BLECrypto is a container for all of the BLE crypto
type BLECrypto struct {
	keys            sodium.KXKP
	remotePublicKey sodium.KXPublicKey
	encryptionNonce [24]uint8
	decryptionNonce [24]uint8
	decrypt         [32]byte
	encrypt         [32]byte
}

// New returns a BLECrypto with a populated keyset
func New() *BLECrypto {
	return &BLECrypto{
		keys: sodium.MakeKXKP(),
	}
}
