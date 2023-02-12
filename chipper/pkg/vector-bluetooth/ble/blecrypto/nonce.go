package blecrypto

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"

import (
	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

var (
	cryptoSecretBoxNonceBytes = int(C.crypto_secretbox_noncebytes())
)

// SetNonces populates the nonces
func (b *BLECrypto) SetNonces(msg *rts.RtsNonceMessage) error {
	b.encryptionNonce = msg.ToRobotNonce
	b.decryptionNonce = msg.ToDeviceNonce
	return nil
}

func (b *BLECrypto) nextEncryptNonce() {
	C.sodium_increment(
		(*C.uchar)(&b.encryptionNonce[0]),
		C.size_t(cryptoSecretBoxNonceBytes),
	)
}

func (b *BLECrypto) nextDecryptNonce() {
	C.sodium_increment(
		(*C.uchar)(&b.decryptionNonce[0]),
		C.size_t(cryptoSecretBoxNonceBytes),
	)
}
