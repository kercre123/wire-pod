package logcollector

import (
	"net/http"
	"net/url"

	"github.com/digital-dream-labs/vector-cloud/internal/token"
	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

type options struct {
	server           bool
	socketNameSuffix string
	tokener          token.Accessor
	httpClient       *http.Client
	errListener      util.ErrorListener

	bucketName       string
	s3BasePrefix     string
	awsRegion        string
	endpoint         string
	s3ForcePathStyle bool
	disableSSL       bool
}

// Option defines an option that can be set on the server
type Option func(o *options)

// WithServer specifies that an IPC server should be started so other processes
// can use log collection
func WithServer() Option {
	return func(o *options) {
		o.server = true
	}
}

// WithSocketNameSuffix specifies the (optional) suffix of the socket name
func WithSocketNameSuffix(socketNameSuffix string) Option {
	return func(o *options) {
		o.socketNameSuffix = socketNameSuffix
	}
}

// WithHTTPClient specifies the HTTP client to use
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}

// WithTokener specifies that the given token.Accessor should be used to obtain
// authorization credentials (used to retrieve USerID)
func WithTokener(value token.Accessor) Option {
	return func(o *options) {
		o.tokener = value
	}
}

// WithEndpoint specifies the S3 endpoint
func WithEndpoint(endpoint string) Option {
	return func(o *options) {
		o.endpoint = endpoint
	}
}

// WithS3ForcePathStyle sets whether path style needs to be forced
func WithS3ForcePathStyle(s3ForcePathStyle bool) Option {
	return func(o *options) {
		o.s3ForcePathStyle = s3ForcePathStyle
	}
}

// WithDisableSSL sets whether SSL is used to access the cloud
func WithDisableSSL(disableSSL bool) Option {
	return func(o *options) {
		o.disableSSL = disableSSL
	}
}

// WithS3UrlPrefix specifies the S3 bucket and key prefix in the cloud
// E.g. s3://anki-device-logs-dev/victor
func WithS3UrlPrefix(s3UrlPrefix string) Option {
	return func(o *options) {
		parsedURL, err := url.Parse(s3UrlPrefix)
		if err == nil {
			o.bucketName = parsedURL.Host
			o.s3BasePrefix = parsedURL.Path
		}
	}
}

// WithAwsRegion specifies the AWS region
func WithAwsRegion(awsRegion string) Option {
	return func(o *options) {
		o.awsRegion = awsRegion
	}
}

// WithErrorListener reports errors to the specified error listener
func WithErrorListener(errListener util.ErrorListener) Option {
	return func(o *options) {
		o.errListener = errListener
	}
}
