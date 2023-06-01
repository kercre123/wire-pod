package logcollector

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"

	ac "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/digital-dream-labs/hugh/testing/s3"
	gc "google.golang.org/grpc/credentials"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	accessKey      = "xxx"
	secretKey      = "yyy"
	region         = "us-east-1"
	testBucketName = "logbucket"
	testPathPrefix = "pathprefix"
	testUserID     = "test-user-id"
	testDeviceID   = "00000000"
)

// Accessor implementation that allows for testing without real token service
type TestTokener struct{}

func (t TestTokener) Credentials() (gc.PerRPCCredentials, error) {
	return nil, nil
}

func (t TestTokener) UserID() string {
	return testUserID
}

func (t TestTokener) GetStsCredentials() (*ac.Credentials, error) {
	return ac.NewStaticCredentialsFromCreds(ac.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}), nil
}

func (t TestTokener) IdentityProvider() identity.Provider {
	return nil
}

func newTestLogCollector(endpoint, bucketName string) (*logCollector, error) {
	logcollectorOpts := []Option{WithServer()}

	logcollectorOpts = append(logcollectorOpts, WithTokener(TestTokener{}))
	logcollectorOpts = append(logcollectorOpts, WithEndpoint(endpoint))
	logcollectorOpts = append(logcollectorOpts, WithS3ForcePathStyle(true))
	logcollectorOpts = append(logcollectorOpts, WithS3UrlPrefix(fmt.Sprintf("s3://%s/%s", bucketName, testPathPrefix)))
	logcollectorOpts = append(logcollectorOpts, WithDisableSSL(true))
	logcollectorOpts = append(logcollectorOpts, WithAwsRegion(region))

	var opts options
	for _, o := range logcollectorOpts {
		o(&opts)
	}

	return newLogCollector(&opts)
}

type LogCollectorSuite struct {
	s3.S3Suite

	testRunID    int
	testFilePath string
}

func (s *LogCollectorSuite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
	s.testRunID = rand.Intn(100000)

	s.AccessKey = accessKey
	s.SecretKey = secretKey

	s.S3Suite.SetupSuite()

	s.CreateBucket(testBucketName)

	s.createTestLogFile(fmt.Sprintf("test_run_%d\n", s.testRunID))
}

func (s *LogCollectorSuite) TearDownSuite() {
	s.S3Suite.TearDownSuite()
}

func (s *LogCollectorSuite) createTestLogFile(content string) error {
	file, err := ioutil.TempFile(os.TempDir(), "logcollector_test_")
	if err != nil {
		return err
	}
	defer file.Close()

	s.testFilePath = file.Name()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogCollectorSuite) TestUploadNoServer() {
	t := s.T()

	collector, err := newTestLogCollector("http://1.2.3.4:12345", testBucketName)
	require.NoError(t, err)

	s3Url, err := collector.Upload(context.Background(), testBucketName)
	s.Error(err)
	s.Empty(s3Url)
}

func (s *LogCollectorSuite) TestUploadNonExistingLogFile() {
	t := s.T()

	collector, err := newTestLogCollector(s.Container.Addr(), testBucketName)
	require.NoError(t, err)

	s3Url, err := collector.Upload(context.Background(), "/non/existing/file.log")
	s.Error(err)
	s.Equal("open /non/existing/file.log: no such file or directory", err.Error())
	s.Empty(s3Url)
}

func (s *LogCollectorSuite) TestUploadNonExistingBucketName() {
	t := s.T()

	collector, err := newTestLogCollector(s.Container.Addr(), "non-existing-bucket")
	require.NoError(t, err)

	s3Url, err := collector.Upload(context.Background(), s.testFilePath)

	s.Error(err)
	s.True(strings.HasPrefix(err.Error(), "NoSuchBucket: The specified bucket does not exist"))
	s.Empty(s3Url)
}

func (s *LogCollectorSuite) TestUpload() {
	t := s.T()

	collector, err := newTestLogCollector(s.Container.Addr(), testBucketName)
	require.NoError(t, err)

	s3Url, err := collector.Upload(context.Background(), s.testFilePath)
	require.NoError(t, err)

	parsedURL, err := url.Parse(s3Url)
	require.NoError(t, err)
	s.NotEmpty(s3Url)

	s.Equal(s.Container.Addr(), parsedURL.Host)

	segments := strings.Split(parsedURL.Path[1:], "/")
	s.Equal([]string{testBucketName, testPathPrefix, testUserID, testDeviceID}, segments[0:4])

	key := strings.Join(segments[1:], "/")
	content := s.GetObject(testBucketName, key)

	s.Equal(fmt.Sprintf("test_run_%d\n", s.testRunID), string(content))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(LogCollectorSuite))
}
