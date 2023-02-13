package blecrypto

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"
import (
	"errors"
)

const (
	aBytes int = C.crypto_aead_xchacha20poly1305_ietf_ABYTES // Size of an authentication tag in bytes
)

// DecryptMessage decrypts an incoming message
func (b *BLECrypto) DecryptMessage(msg []byte) ([]byte, error) {
	m := make([]byte, len(msg)-aBytes)

	exit := C.crypto_aead_xchacha20poly1305_ietf_decrypt(
		(*C.uchar)(bytePointer(m)),
		(*C.ulonglong)(nil),
		(*C.uchar)(nil),
		(*C.uchar)(&msg[0]),
		C.ulonglong(len(msg)),
		(*C.uchar)(nil),
		C.ulonglong(0),
		(*C.uchar)(&b.decryptionNonce[0]),
		(*C.uchar)(&b.decrypt[0]))

	if exit != 0 {
		return nil, errors.New("verification failed")
	}

	b.nextDecryptNonce()
	return m, nil
}
