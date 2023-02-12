package rts

import "errors"

// BuildAuthMessage builds the auth request
func BuildAuthMessage(version int, key string) ([]byte, error) {
	switch version {
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsCloudSessionRequest5(
				&RtsCloudSessionRequest_5{
					SessionToken: key,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
