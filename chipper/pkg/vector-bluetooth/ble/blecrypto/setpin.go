package blecrypto

// #cgo pkg-config: libsodium
// #include <stdlib.h>
// #include <sodium.h>
import (
	"C"
)

import (
	"errors"
)

const (
	keysize = 32
)

// SetPin configures the hashes based on the pin entered
func (b *BLECrypto) SetPin(pin string) error {
	if b.remotePublicKey.Bytes == nil {
		return errors.New("remote public key is not set")
	}

	k, err := b.keys.ClientSessionKeys(b.remotePublicKey)
	if err != nil {
		return err
	}
	b.decrypt = genHash(k.Rx.Bytes, []byte(pin))
	b.encrypt = genHash(k.Tx.Bytes, []byte(pin))
	return nil
}

func genHash(key, pin []byte) [32]byte {
	h, _ := sodiumGenerichash(keysize, key, pin)
	rv := [32]byte{}
	for k, v := range h {
		rv[k] = v
	}
	return rv
}

func sodiumGenerichash(outlen int, in, key []byte) (hash []byte, bytes int) {
	hash = make([]byte, outlen)
	exit := int(C.crypto_generichash(
		(*C.uchar)(&hash[0]),
		C.size_t(outlen),
		(*C.uchar)(bytePointer(in)),
		C.ulonglong(len(in)),
		(*C.uchar)(bytePointer(key)),
		C.size_t(len(key))))

	return hash, exit
}

func bytePointer(b []byte) *uint8 {
	if len(b) > 0 {
		return &b[0]
	}
	return nil
}
