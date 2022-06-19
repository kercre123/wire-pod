package logcollector

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/token"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// logCollector implements functionality for uploading log files to the cloud
type logCollector struct {
	tokener token.Accessor

	certCommonName string

	bucketName       string
	s3BasePrefix     string
	awsRegion        string
	endpoint         string
	s3ForcePathStyle bool
	disableSSL       bool

	httpClient *http.Client

	uploader *s3manager.Uploader
}

func newLogCollector(opts *options) (*logCollector, error) {
	c := &logCollector{
		tokener: opts.tokener,

		bucketName:       opts.bucketName,
		s3BasePrefix:     opts.s3BasePrefix,
		awsRegion:        opts.awsRegion,
		endpoint:         opts.endpoint,
		s3ForcePathStyle: opts.s3ForcePathStyle,
		disableSSL:       opts.disableSSL,

		httpClient: opts.httpClient,
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	c.certCommonName = opts.tokener.IdentityProvider().CertCommonName()

	awsCredentials, err := c.tokener.GetStsCredentials()
	if err != nil {
		return nil, err
	}

	awsSession, err := session.NewSession(&aws.Config{
		HTTPClient:  c.httpClient,
		Credentials: awsCredentials,
		Region:      aws.String(c.awsRegion),

		// Required for testing purposes
		Endpoint:         aws.String(c.endpoint),
		S3ForcePathStyle: aws.Bool(c.s3ForcePathStyle),
		DisableSSL:       aws.Bool(c.disableSSL),
	})
	if err != nil {
		return nil, err
	}

	c.uploader = s3manager.NewUploader(awsSession)

	return c, nil
}

// Upload uploads file to cloud
func (c *logCollector) Upload(ctx context.Context, logFilePath string) (string, error) {
	const defaultUserID = "unknown-user-id"

	// As the user ID may (theoretically) change we retrieve it here for every upload
	userID := defaultUserID
	if c.tokener != nil {
		userID = c.tokener.UserID()
		if userID == "" {
			// Create a sensible fallback user ID for cloud uploads (in case no token is stored in file system)
			userID = defaultUserID
		}
	}

	timestamp := time.Now().UTC()

	encodedCertCommonName := base64.StdEncoding.EncodeToString([]byte(c.certCommonName))
	s3Prefix := path.Join(c.s3BasePrefix, userID, encodedCertCommonName)
	s3FileName := fmt.Sprintf("%s-%s", timestamp.Format("2006-01-02-15-04-05"), path.Base(logFilePath))
	s3Key := path.Join(s3Prefix, s3FileName)

	logFile, err := os.Open(logFilePath)
	if err != nil {
		return "", err
	}

	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(s3Key),
		Body:   logFile,
	}

	result, err := c.uploader.UploadWithContext(ctx, uploadInput)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			dasFields := (&log.DasFields{}).SetStrings(awsErr.Code(), awsErr.Message(), c.bucketName)
			log.Das("logcollector.upload.error", dasFields)
			return "", fmt.Errorf("%s: %s (Bucket=%q, Key=%q)", awsErr.Code(), awsErr.Message(), c.bucketName, s3Key)
		}
		dasFields := (&log.DasFields{}).SetStrings(err.Error(), "", c.bucketName)
		log.Das("logcollector.upload.error", dasFields)
		return "", fmt.Errorf("%v (Bucket=%q, Key=%q)", err, c.bucketName, s3Key)
	}

	dasFields := (&log.DasFields{}).SetStrings(logFilePath, result.Location, c.bucketName)
	log.Das("logcollector.upload.success", dasFields)

	log.Printf("File %q uploaded to %q\n", logFilePath, result.Location)

	return result.Location, nil
}
