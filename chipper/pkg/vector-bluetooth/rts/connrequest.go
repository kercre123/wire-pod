package rts

import (
	"errors"
)

// GetConnResponse builds the RTS5 connection response
func GetConnResponse(version int, pubkey [32]uint8) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsConnResponse(
				&RtsConnResponse{
					ConnectionType: RtsConnType_FirstTimePair,
					PublicKey:      pubkey,
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsConnResponse(
				&RtsConnResponse{
					ConnectionType: RtsConnType_FirstTimePair,
					PublicKey:      pubkey,
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsConnResponse(
				&RtsConnResponse{
					ConnectionType: RtsConnType_FirstTimePair,
					PublicKey:      pubkey,
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsConnResponse(
				&RtsConnResponse{
					ConnectionType: RtsConnType_FirstTimePair,
					PublicKey:      pubkey,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
