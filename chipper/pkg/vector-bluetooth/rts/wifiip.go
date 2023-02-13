package rts

import (
	"errors"
)

// BuildWifiScanMessage builds the wifi scan message
func BuildWifiIPMessage(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsWifiIpRequest(
				&RtsWifiIpRequest{},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsWifiIpRequest(
				&RtsWifiIpRequest{},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsWifiIpRequest(
				&RtsWifiIpRequest{},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsWifiIpRequest(
				&RtsWifiIpRequest{},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
