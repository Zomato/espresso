package templatestore

import (
	"context"
	"errors"
	"io"
	"text/template"

	"github.com/Zomato/espresso/lib/s3"
)

const (
	StorageAdapterTypeDisk   = "disk"
	StorageAdapterTypeS3     = "s3"
	StorageAdapterTypeStream = "stream"
	StorageAdapterTypeMySQL  = "mysql"
)

type StorageAdapter interface {
	// GetTemplate retrieves a template from storage.
	GetTemplate(ctx context.Context, req *GetTemplateRequest) (*template.Template, error)

	// PutDocument stores a file in storage.
	PutDocument(ctx context.Context, req *PostDocumentRequest, reader *io.Reader) (string, error)

	GetDocument(ctx context.Context, req *GetDocumentRequest) (io.Reader, error)

	ListTemplates(ctx context.Context) ([]*TemplateInfo, error)

	GetTemplateContent(ctx context.Context, req *GetTemplateContentRequest) (*GetTemplateContentResponse, error)

	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (string, error)
}

// TemplateStorageAdapterFactory is a factory function for creating template storage adapters.
func TemplateStorageAdapterFactory(conf *StorageConfig) (StorageAdapter, error) {
	switch conf.StorageType {
	case StorageAdapterTypeDisk:
		return &DiskTemplateStorage{}, nil
	case StorageAdapterTypeS3:
		if conf == nil {
			return nil, errors.New("templateStorageConfig is required")
		}
		if conf.S3Config == nil {
			return nil, errors.New("S3 configuration is required")
		}
		if conf.AwsCredConfig == nil {
			return nil, errors.New("AWS credentials are required")
		}
		s3Adapter, err := NewS3StorageAdapter(context.Background(), s3.WithEndpoint(conf.S3Config.Endpoint),
			s3.WithDebug(conf.S3Config.Debug),
			s3.WithRegion(conf.S3Config.Region),
			s3.WithForcePathStyle(conf.S3Config.ForcePathStyle),
			s3.WithUploaderConcurrency(conf.S3Config.UploaderConcurrency),
			s3.WithUploaderPartSize(conf.S3Config.UploaderPartSize),
			s3.WithDownloaderConcurrency(conf.S3Config.DownloaderConcurrency),
			s3.WithDownloaderPartSize(conf.S3Config.DownloaderPartSize),
			s3.WithRetryMaxAttempts(conf.S3Config.RetryMaxAttempts),
			s3.WithBucket(conf.S3Config.Bucket),
			s3.WithCustomTransport(conf.S3Config.UseCustomTransport),
			s3.WithCredentials(conf.AwsCredConfig.AccessKeyID,
				conf.AwsCredConfig.SecretAccessKey, conf.AwsCredConfig.SessionToken))
		if err != nil {
			return nil, err
		}
		return s3Adapter, nil
	case StorageAdapterTypeStream:
		return &StreamStorage{}, nil
	case StorageAdapterTypeMySQL:
		mysqlDSN := conf.MysqlDSN
		if mysqlDSN == "" {
			return nil, errors.New("mysql DSN not configured")
		}
		mysqlAdapter, err := NewMySQLStorageAdapter(mysqlDSN)
		if err != nil {
			return nil, err
		}
		return mysqlAdapter, nil
	default:
		return nil, errors.New("unsupported storage type")
	}
}
