package rts

import "errors"

// BuildStatusMessage builds the status request message
func BuildStatusMessage(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsStatusRequest(
				&RtsStatusRequest{},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsStatusRequest(
				&RtsStatusRequest{},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsStatusRequest(
				&RtsStatusRequest{},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsStatusRequest(
				&RtsStatusRequest{},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
