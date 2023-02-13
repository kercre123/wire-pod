package rts

import (
	"errors"
)

// BuildWifiScanMessage builds the wifi scan message
func BuildLogRequestMessage(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsLogRequest(
				&RtsLogRequest{
					Mode:   0,
					Filter: []string{},
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsLogRequest(
				&RtsLogRequest{
					Mode:   0,
					Filter: []string{},
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsLogRequest(
				&RtsLogRequest{
					Mode:   0,
					Filter: []string{},
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsLogRequest(
				&RtsLogRequest{
					Mode:   0,
					Filter: []string{},
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
