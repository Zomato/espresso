package s3

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint              string
	Region                string
	ForcePathStyle        bool
	UploaderPartSize      int64
	UploaderConcurrency   int
	DownloaderPartSize    int64
	DownloaderConcurrency int
	Debug                 bool
	RetryMaxAttempts      int
	Bucket                string
	AccessKeyID           string
	SecretAccessKey       string
	SessionToken          string
	UseCustomTransport    bool
}
type AwsCredConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type S3Client struct {
	Uploader   *manager.Uploader
	Presigner  *s3.PresignClient
	Downloader *manager.Downloader
	Config     *Config
}

func NewS3Client(ctx context.Context, options ...func(*Config)) (*S3Client, error) {
	s3Client := &S3Client{
		Config: &Config{},
	}

	config := s3Client.Config
	for _, option := range options {
		option(config)
	}

	logMode := aws.LogDeprecatedUsage | aws.LogRetries
	if config.Debug {
		logMode |= aws.LogRequest | aws.LogResponse
	}

	cfgOptions := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRetryMaxAttempts(config.RetryMaxAttempts),
		awsConfig.WithRegion(config.Region),
		awsConfig.WithClientLogMode(logMode),
	}
	if config.AccessKeyID != "" && config.SecretAccessKey != "" {
		cfgOptions = append(cfgOptions, awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, config.SessionToken)))
	}
	// used for local as localstack has endpoint of format localhost:4566
	if os.Getenv("GO_ENV") == "local" {
		cfgOptions = append(cfgOptions, awsConfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               config.Endpoint,
				SigningRegion:     config.Region,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		})))
	}
	if config.UseCustomTransport {
		// Create custom HTTP client with TLS config that skips verification
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		cfgOptions = append(cfgOptions, awsConfig.WithHTTPClient(&http.Client{
			Transport: customTransport,
		}))
	}

	s3AwsConfig, err := awsConfig.LoadDefaultConfig(
		ctx,
		cfgOptions...,
	)

	awsS3Client := s3.NewFromConfig(s3AwsConfig, func(o *s3.Options) {
		o.UsePathStyle = config.ForcePathStyle
	})

	// https://levyeran.medium.com/high-memory-allocations-and-gc-cycles-while-downloading-large-s3-objects-using-the-aws-sdk-for-go-e776a136c5d0
	uploader := manager.NewUploader(awsS3Client, func(u *manager.Uploader) {
		u.PartSize = config.UploaderPartSize * 1024 * 1024
		u.Concurrency = config.UploaderConcurrency
		u.BufferProvider = manager.NewBufferedReadSeekerWriteToPool(int(config.UploaderPartSize) * 1024 * 1024)
		u.LeavePartsOnError = false
	})

	downloader := manager.NewDownloader(awsS3Client, func(d *manager.Downloader) {
		d.PartSize = config.DownloaderPartSize * 1024 * 1024
		d.Concurrency = config.DownloaderConcurrency
		d.BufferProvider = manager.NewPooledBufferedWriterReadFromProvider(int(config.DownloaderPartSize) * 1024 * 1024)
	})

	presignClient := s3.NewPresignClient(awsS3Client)

	s3Client.Uploader = uploader
	s3Client.Downloader = downloader
	s3Client.Presigner = presignClient

	if err != nil {
		return nil, errors.Join(err, errors.New("failed to load aws config"))
	}
	return s3Client, nil
}

func (s3Client *S3Client) UploadFile(ctx context.Context, key string, body io.Reader) (*manager.UploadOutput, error) {
	input := &s3.PutObjectInput{
		Body:   body,
		Bucket: aws.String(s3Client.Config.Bucket),
		Key:    aws.String(key),
	}
	output, err := s3Client.Uploader.Upload(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s3Client *S3Client) DownloadFile(ctx context.Context, key string, writer io.WriterAt) (int64, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Client.Config.Bucket),
		Key:    aws.String(key),
	}
	n, err := s3Client.Downloader.Download(ctx, writer, input)
	if err != nil {
		return 0, err
	}
	return n, nil
}
func (s3Client *S3Client) GetFileReader(ctx context.Context, templatePath string) (io.Reader, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Client.Config.Bucket),
		Key:    aws.String(templatePath),
	}
	resp, err := s3Client.Downloader.S3.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s3Client *S3Client) GetPresignURL(ctx context.Context, key string, presignTime int) (*v4.PresignedHTTPRequest, error) {
	presign, err := s3Client.Presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Client.Config.Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(presignTime) * time.Second
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get presign url"))
	}
	return presign, nil
}

func (s3Client *S3Client) UploadFileToS3AndGetPresignedURL(ctx context.Context, filePath string, presignTime int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to open file"))
	}
	defer file.Close()
	if _, err := s3Client.UploadFile(ctx, filePath, file); err != nil {
		return "", errors.Join(err, errors.New("failed to upload file"))

	}
	presign, err := s3Client.GetPresignURL(ctx, filePath, presignTime)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to get presign url"))
	}
	return presign.URL, nil
}

func WithEndpoint(endpoint string) func(*Config) {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

func WithRegion(region string) func(*Config) {
	return func(c *Config) {
		c.Region = region
	}
}

func WithForcePathStyle(forcePathStyle bool) func(*Config) {
	return func(c *Config) {
		c.ForcePathStyle = forcePathStyle
	}
}

func WithUploaderPartSize(uploaderPartSize int64) func(*Config) {
	return func(c *Config) {
		c.UploaderPartSize = uploaderPartSize
	}
}

func WithUploaderConcurrency(uploaderConcurrency int) func(*Config) {
	return func(c *Config) {
		c.UploaderConcurrency = uploaderConcurrency
	}
}

func WithDebug(debug bool) func(*Config) {
	return func(c *Config) {
		c.Debug = debug
	}
}

func WithDownloaderPartSize(downloaderPartSize int64) func(*Config) {
	return func(c *Config) {
		c.DownloaderPartSize = downloaderPartSize
	}
}

func WithDownloaderConcurrency(downloaderConcurrency int) func(*Config) {
	return func(c *Config) {
		c.DownloaderConcurrency = downloaderConcurrency
	}
}

func WithRetryMaxAttempts(retryMaxAttempts int) func(*Config) {
	return func(c *Config) {
		c.RetryMaxAttempts = retryMaxAttempts
	}
}

func WithBucket(bucket string) func(*Config) {
	return func(c *Config) {
		c.Bucket = bucket
	}
}
func WithCredentials(accessKeyID, accessKey, sessionToken string) func(*Config) {
	return func(c *Config) {
		c.AccessKeyID = accessKeyID
		c.SecretAccessKey = accessKey
		c.SessionToken = sessionToken
	}
}

// Add this new function at the end of the file
func WithCustomTransport(useCustomTransport bool) func(*Config) {
	return func(c *Config) {
		c.UseCustomTransport = useCustomTransport
	}
}
