package token

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	testtime "github.com/digital-dream-labs/vector-cloud/internal/testing/time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StsSuite struct {
	suite.Suite

	stsTestCache *stsCredentialsCache
}

var testCredentials = credentials.NewStaticCredentialsFromCreds(credentials.Value{
	AccessKeyID:     "TestAccessKeyId",
	SecretAccessKey: "TestSecretAccessKey",
	SessionToken:    "TestSessionToken",
})

func init() {
	// replace time.Now with something test-friendly.
	testableTime = testtime.NewTestable()
}

func (s *StsSuite) SetupTest() {
	s.stsTestCache = new(stsCredentialsCache)
}

func (s *StsSuite) TestStsCaching() {
	testExpirationTime := time.Now().Add(3 * time.Hour).UTC()

	cacheTests := []struct {
		name                     string
		expirationString         string
		expectedExpirationString string
		expectedCredentials      *credentials.Credentials
		isExpired                bool
	}{
		{
			name:                     "no-add-empty-cache",
			expirationString:         "",
			expectedExpirationString: "0001-01-01T00:00:00Z",
			expectedCredentials:      nil,
			isExpired:                true,
		}, {
			name:                     "token-within-refresh-window",
			expirationString:         "incorrect-time-string",
			expectedExpirationString: "0001-01-01T00:00:00Z",
			expectedCredentials:      nil,
			isExpired:                true,
		}, {
			name:                     "successful-add",
			expirationString:         testExpirationTime.Format(time.RFC3339),
			expectedExpirationString: testExpirationTime.Format(time.RFC3339),
			expectedCredentials:      testCredentials,
			isExpired:                false,
		},
	}

	for _, test := range cacheTests {
		if test.expirationString != "" {
			s.stsTestCache.add(test.expirationString, testCredentials)
		}

		s.Equal(test.isExpired, s.stsTestCache.expired(), test.name)
		s.Equal(test.expectedExpirationString, s.stsTestCache.expiration.Format(time.RFC3339), test.name)
		s.Equal(test.expectedCredentials, s.stsTestCache.credentials, test.name)
	}
}

func (s *StsSuite) TestStsTokenExpiration() {
	t := s.T()

	testableTime := testableTime.(testtime.TestableTime)
	s.stsTestCache.add(time.Now().UTC().Format(time.RFC3339), testCredentials)

	expirationTests := []struct {
		name      string
		nowOffset time.Duration
		isExpired bool
	}{
		{
			name:      "token-not-expired",
			nowOffset: -3 * time.Hour,
			isExpired: false,
		}, {
			name:      "token-within-refresh-window",
			nowOffset: -30 * time.Minute,
			isExpired: true,
		}, {
			name:      "token-expired",
			nowOffset: 3 * time.Hour,
			isExpired: true,
		},
	}

	for _, test := range expirationTests {
		testableTime.WithNowDelta(test.nowOffset, func() {
			assert.Equal(t, test.isExpired, s.stsTestCache.expired(), test.name)
		})
	}
}

func TestStsSuite(t *testing.T) {
	suite.Run(t, new(StsSuite))
}
