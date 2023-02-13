package rts

import (
	"errors"
)

// BuildNonceResponse builds the acknowledgement message for the nonce
func BuildNonceResponse(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsAck(
				&RtsAck{
					RtsConnectionTag: uint8(RtsConnection_5Tag_RtsNonceMessage),
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsAck(
				&RtsAck{
					RtsConnectionTag: uint8(RtsConnection_5Tag_RtsNonceMessage),
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsAck(
				&RtsAck{
					RtsConnectionTag: uint8(RtsConnection_5Tag_RtsNonceMessage),
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsAck(
				&RtsAck{
					RtsConnectionTag: uint8(RtsConnection_5Tag_RtsNonceMessage),
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
