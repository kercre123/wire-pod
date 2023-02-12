package rts

import "errors"

// BuildOTAStartMessage builds the ota start message
func BuildOTAStartMessage(version int, url string) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsOtaUpdateRequest(
				&RtsOtaUpdateRequest{
					Url: url,
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsOtaUpdateRequest(
				&RtsOtaUpdateRequest{
					Url: url,
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsOtaUpdateRequest(
				&RtsOtaUpdateRequest{
					Url: url,
				},
			),
		)

	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsOtaUpdateRequest(
				&RtsOtaUpdateRequest{
					Url: url,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
