package rts

import (
	"errors"
)

// BuildWifiAccesspointMessage builds the wifi AP message
func BuildWifiAccesspointMessage(version int, enabled bool) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsWifiAccessPointRequest(
				&RtsWifiAccessPointRequest{
					Enable: enabled,
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsWifiAccessPointRequest(
				&RtsWifiAccessPointRequest{
					Enable: enabled,
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsWifiAccessPointRequest(
				&RtsWifiAccessPointRequest{
					Enable: enabled,
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsWifiAccessPointRequest(
				&RtsWifiAccessPointRequest{
					Enable: enabled,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
