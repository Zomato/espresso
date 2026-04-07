package server

import (
	"github.com/Zomato/espresso/lib/s3"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/service/pkg/config"
)

// initStorageAdapters initialises the template and file storage adapters from
// config and returns them. Both http and mcp server registrations call this so
// the initialisation logic lives in one place.
func initStorageAdapters() (templatestore.StorageAdapter, templatestore.StorageAdapter, error) {
	cfg := config.GetConfig()

	s3Cfg := &s3.Config{
		Endpoint:              cfg.S3.Endpoint,
		Region:                cfg.S3.Region,
		Bucket:                cfg.S3.Bucket,
		Debug:                 cfg.S3.Debug,
		ForcePathStyle:        cfg.S3.ForcePathStyle,
		UploaderConcurrency:   cfg.S3.UploaderConcurrency,
		UploaderPartSize:      cfg.S3.UploaderPartSize,
		DownloaderConcurrency: cfg.S3.DownloaderConcurrency,
		DownloaderPartSize:    cfg.S3.DownloaderPartSize,
		RetryMaxAttempts:      cfg.S3.RetryMaxAttempts,
		UseCustomTransport:    cfg.S3.UseCustomTransport,
	}

	awsCfg := &s3.AwsCredConfig{
		AccessKeyID:     cfg.AWS.AccessKeyID,
		SecretAccessKey: cfg.AWS.SecretAccessKey,
		SessionToken:    cfg.AWS.SessionToken,
	}

	templateStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType:   cfg.TemplateStorage.StorageType,
		S3Config:      s3Cfg,
		AwsCredConfig: awsCfg,
		MysqlDSN:      cfg.MySQL.DSN,
	})
	if err != nil {
		return nil, nil, err
	}

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType:   cfg.FileStorage.StorageType,
		S3Config:      s3Cfg,
		AwsCredConfig: awsCfg,
	})
	if err != nil {
		return nil, nil, err
	}

	return templateStorageAdapter, fileStorageAdapter, nil
}
