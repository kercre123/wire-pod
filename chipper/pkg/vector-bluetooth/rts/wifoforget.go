package rts

import (
	"encoding/hex"
	"errors"
)

// BuildWifiConnectMessage builds the wifi connect message
func BuildWifiForgetMessage(version int, ssid string, all bool) ([]byte, error) {
	switch version {
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsWifiForgetRequest(
				&RtsWifiForgetRequest{
					DeleteAll:   all,
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsWifiForgetRequest(
				&RtsWifiForgetRequest{
					DeleteAll:   all,
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsWifiForgetRequest(
				&RtsWifiForgetRequest{
					DeleteAll:   all,
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
