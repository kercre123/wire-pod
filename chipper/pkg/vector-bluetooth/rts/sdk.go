package rts

import "errors"

// BuildSDKMessage builds an SDK message
func BuildSDKMessage(version int, token, id, urlpath, json string) ([]byte, error) {
	switch version {
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsSdkProxyRequest(
				&RtsSdkProxyRequest{
					ClientGuid: token,
					MessageId:  id,
					UrlPath:    urlpath,
					Json:       json,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)
	}
}
