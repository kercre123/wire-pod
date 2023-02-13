package rts

import "errors"

// BuildChallengeResponse builds the challenge response
func BuildChallengeResponse(version int, number uint32) ([]byte, error) {
	switch version {
	case rtsv2:
		return buildMessage2(
			NewRtsConnection_2WithRtsChallengeMessage(
				&RtsChallengeMessage{
					Number: number + 1,
				},
			),
		)
	case rtsv3:
		return buildMessage3(
			NewRtsConnection_3WithRtsChallengeMessage(
				&RtsChallengeMessage{
					Number: number + 1,
				},
			),
		)
	case rtsv4:
		return buildMessage4(
			NewRtsConnection_4WithRtsChallengeMessage(
				&RtsChallengeMessage{
					Number: number + 1,
				},
			),
		)
	case rtsv5:
		return buildMessage5(
			NewRtsConnection_5WithRtsChallengeMessage(
				&RtsChallengeMessage{
					Number: number + 1,
				},
			),
		)
	default:
		return nil, errors.New(errUnsupportedVersion)

	}
}
