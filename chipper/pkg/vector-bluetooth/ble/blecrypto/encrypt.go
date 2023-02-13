package blecrypto

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import "C"

// Encrypt encrypts an incoming message
func (b *BLECrypto) Encrypt(msg []byte) ([]byte, error) {
	c := make([]byte, len(msg)+aBytes)

	C.crypto_aead_xchacha20poly1305_ietf_encrypt(
		(*C.uchar)(bytePointer(c)),
		(*C.ulonglong)(nil),
		(*C.uchar)(bytePointer(msg)),
		C.ulonglong(len(msg)),
		(*C.uchar)(nil),
		C.ulonglong(0),
		(*C.uchar)(nil),
		(*C.uchar)(&b.encryptionNonce[0]),
		(*C.uchar)(&b.encrypt[0]))

	b.nextEncryptNonce()
	return c, nil
}
