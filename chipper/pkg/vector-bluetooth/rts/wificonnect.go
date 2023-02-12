package rts

import (
	"encoding/hex"
	"errors"
)

// BuildWifiConnectMessage builds the wifi connect message
func BuildWifiConnectMessage(version int, ssid, password string, timeout, authtype int) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsWifiConnectRequest(
				&RtsWifiConnectRequest{
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
					Password:    password,
					Timeout:     uint8(timeout),
					AuthType:    uint8(authtype),
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsWifiConnectRequest(
				&RtsWifiConnectRequest{
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
					Password:    password,
					Timeout:     uint8(timeout),
					AuthType:    uint8(authtype),
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsWifiConnectRequest(
				&RtsWifiConnectRequest{
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
					Password:    password,
					Timeout:     uint8(timeout),
					AuthType:    uint8(authtype),
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsWifiConnectRequest(
				&RtsWifiConnectRequest{
					WifiSsidHex: hex.EncodeToString([]byte(ssid)),
					Password:    password,
					Timeout:     uint8(timeout),
					AuthType:    uint8(authtype),
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
