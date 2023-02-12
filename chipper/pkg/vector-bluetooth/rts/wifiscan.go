package rts

import (
	"errors"
)

// BuildWifiScanMessage builds the wifi scan message
func BuildWifiScanMessage(version int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsWifiScanRequest(
				&RtsWifiScanRequest{},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsWifiScanRequest(
				&RtsWifiScanRequest{},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsWifiScanRequest(
				&RtsWifiScanRequest{},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsWifiScanRequest(
				&RtsWifiScanRequest{},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
