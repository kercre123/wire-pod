package clad

import "bytes"

type Struct interface {
	Size() uint32
	Pack(*bytes.Buffer) error
	Unpack(*bytes.Buffer) error
}
