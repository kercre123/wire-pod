package dockerutil

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/suite"
)

// Suite provides a base type to use for constructing test suites that
// depend on docker container. Concrete test suite types must set the
// Start function.
type Suite struct {
	suite.Suite

	// Container is a pointer to the Container structure, which will
	// be assigned after SetupSuite() successfully starts the
	// container.
	Container *Container

	// Start is a function which launches a container, returning the
	// container and an optional error. Each concrete docker test
	// suite must set the Start member before propagating the
	// SetupSuite() call, and should call one of the container
	// starting functions provided by dockerutil.
	//
	// See S3Suite for a concrete example.
	Start func() (*Container, error)
}

type DockerContainer = docker.Container
type DockerClient = docker.Client

// Client wraps the go-dockerclient Client type to add some convenience methods.
//
// It embeds that type so all properties and methods are supported.
type Client struct {
	*DockerClient
}

// Container wraps the go-dockerclient Container type to add some convenience methods.
//
// It embeds that type so all properties and methods are supported.
type Container struct {
	*DockerContainer
	client *Client
}

// S3Suite is a base test suite type which launches a local Amazon
// S3-compatible docker container and configures the Amazon AWS SDK's
// S3 client to point to it.
//
// It also provides a number of wrapper functions for S3 SDK calls
// that simplify manipulating and inspecting S3 for test case setup
// and teardown. Rather than returning errors, these functions will
// cause test errors or failures, to keep the test code minimal.
type S3Suite struct {
	Suite

	// S3 is the AWS SDK client, which will be assigned when
	// SetupSuite() completes successfully.
	S3        *s3.S3
	AccessKey string
	SecretKey string
	Config    *aws.Config
}

const (
	S3DefaultAccessKey = "access_key"
	S3DefaultSecretKey = "secret_key"
)

func (s *S3Suite) SetupSuite() {
	if s.AccessKey == "" {
		s.AccessKey = S3DefaultAccessKey
	}
	if s.SecretKey == "" {
		s.SecretKey = S3DefaultSecretKey
	}
	s.Start = func() (*Container, error) {
		return StartS3Container(s.AccessKey, s.SecretKey)
	}

	s.Suite.SetupSuite()

	s.Config = &aws.Config{
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("http://" + s.Container.Addr()),
		DisableSSL:       aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(s.AccessKey, s.SecretKey, ""),
	}

	s.S3 = s3.New(session.New(), s.Config)
}

func (s *S3Suite) CreateBucket(name string) {
	_, err := s.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String("us-east-1"),
		},
	})
	if err != nil {
		s.T().Fatalf("Failed to create S3 bucket %s: %s", name, err)
	}
}

func (s *S3Suite) DeleteBucket(name string) {
	_, err := s.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		s.T().Fatalf("Failed to delete S3 bucket %s: %s", name, err)
	}
}

func (s *S3Suite) PutObject(bucket, path, contentType string, data []byte) {
	_, err := s.S3.PutObject(&s3.PutObjectInput{
		ContentEncoding: aws.String(contentType),
		Bucket:          aws.String(bucket),
		Key:             aws.String(path),
		Body:            aws.ReadSeekCloser(bytes.NewReader(data)),
	})
	if err != nil {
		s.T().Errorf("Error uploading %s to bucket %s: %s", path, bucket, err)
	}
}

func (s *S3Suite) PutObjects(bucket, contentType string, objs map[string][]byte) {
	for path, data := range objs {
		s.PutObject(bucket, path, contentType, data)
	}
}

func (s *S3Suite) GetObject(bucket, path string) []byte {
	resp, err := s.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.T().Errorf("Error getting object %s from bucket %s: %s", path, bucket, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.T().Errorf("Error reading body of object %s from bucket %s: %s", path, bucket, err)
	}
	return data
}

func (s *S3Suite) GetObjectWithResponse(bucket, path string) (resp *s3.GetObjectOutput, respData []byte) {
	resp, err := s.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.T().Errorf("Error getting object %s from bucket %s: %s", path, bucket, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.T().Errorf("Error reading body of object %s from bucket %s: %s", path, bucket, err)
	}
	return resp, data
}

func (s *S3Suite) GetObjects(bucket string, paths []string) map[string][]byte {
	result := make(map[string][]byte)

	for _, p := range paths {
		result[p] = s.GetObject(bucket, p)
	}

	return result
}

type S3Object struct {
	Path string
	Size int64
}

func (s *S3Suite) ListObjects(bucket, prefix string) []S3Object {
	var result []S3Object
	err := s.S3.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}, func(resp *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range resp.Contents {
			result = append(result, S3Object{
				Path: *obj.Key,
				Size: *obj.Size,
			})
		}
		return true
	})
	if err != nil {
		s.T().Errorf("LIST of s3://%s/%s failed: %s", bucket, prefix, err)
	}
	return result
}

func (s *S3Suite) DeleteObject(bucket, path string) {
	_, err := s.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.T().Errorf("Error deleting object %s from bucket %s: %s", path, bucket, err)
	}
}

func (s *S3Suite) DeleteObjects(bucket string, paths []string) {
	for _, p := range paths {
		s.DeleteObject(bucket, p)
	}
}

func (s *S3Suite) DeleteAllObjects(bucket string) {
	objs := s.ListObjects(bucket, "")
	for _, obj := range objs {
		s.DeleteObject(bucket, obj.Path)
	}
}

func (s *S3Suite) SetupSuite() {
	if s.AccessKey == "" {
		s.AccessKey = S3DefaultAccessKey
	}
	if s.SecretKey == "" {
		s.SecretKey = S3DefaultSecretKey
	}
	s.Start = func() (*Container, error) {
		return StartS3Container(s.AccessKey, s.SecretKey)
	}

	s.Suite.SetupSuite()

	s.Config = &aws.Config{
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("http://" + s.Container.Addr()),
		DisableSSL:       aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(s.AccessKey, s.SecretKey, ""),
	}

	s.S3 = s3.New(session.New(), s.Config)
}
