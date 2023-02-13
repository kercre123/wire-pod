package rts

import "errors"

// BuildOTACancelMessage builds the ota cancel message
func BuildOTACancelMessage(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsOtaCancelRequest(
				&RtsOtaCancelRequest{},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsOtaCancelRequest(
				&RtsOtaCancelRequest{},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsOtaCancelRequest(
				&RtsOtaCancelRequest{},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsOtaCancelRequest(
				&RtsOtaCancelRequest{},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
