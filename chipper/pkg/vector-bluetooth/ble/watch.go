package ble

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

func (v *VectorBLE) watch() ([]byte, error) {
	var (
		resp []byte
		cont bool
		err  error
	)

	// cont tells the loop whether to continue watching or not.
	cont = true

	for {
		if cont {
			incoming := <-v.bleReader
			z := bytes.NewBuffer(incoming)
			comm := rts.ExternalComms{}
			if err := comm.Unpack(z); err != nil {
				return nil, err
			}
			m := comm.GetRtsConnection()

			if m == nil {
				return nil, errors.New("empty rts connection")
			}

			switch m.Tag() {
			case rts.RtsConnectionTag_RtsConnection2:
				f, ok := rtsHandlers[m.GetRtsConnection2().Tag().String()]
				if !ok {
					cont = false
					err = fmt.Errorf("unsupported message: %v", m.GetRtsConnection2().Tag())
					continue
				}
				resp, cont, err = f(v, m.GetRtsConnection2())

			case rts.RtsConnectionTag_RtsConnection3:
				f, ok := rtsHandlers[m.GetRtsConnection3().Tag().String()]
				if !ok {
					cont = false
					err = fmt.Errorf("unsupported message: %v", m.GetRtsConnection3().Tag())
					continue
				}
				resp, cont, err = f(v, m.GetRtsConnection3())

			case rts.RtsConnectionTag_RtsConnection4:
				f, ok := rtsHandlers[m.GetRtsConnection4().Tag().String()]
				if !ok {
					cont = false
					err = fmt.Errorf("unsupported message: %v", m.GetRtsConnection4().Tag())
					continue
				}
				resp, cont, err = f(v, m.GetRtsConnection4())

			case rts.RtsConnectionTag_RtsConnection5:
				f, ok := rtsHandlers[m.GetRtsConnection5().Tag().String()]
				if !ok {
					cont = false
					err = fmt.Errorf("unsupported message: %v", m.GetRtsConnection5().Tag())
					continue
				}
				resp, cont, err = f(v, m.GetRtsConnection5())

			case rts.RtsConnectionTag_INVALID:
				cont = false
				err = errors.New("invalid message")

			case rts.RtsConnectionTag_Error:
				cont = false
				err = errors.New("invalid message")

			default:
				cont = false
				err = errors.New("unsupported message version")
			}
		} else {
			return resp, err
		}
	}
}
