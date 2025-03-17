package templatestore

import "github.com/Zomato/espresso/lib/s3"

type StorageConfig struct {
	StorageType   string
	S3Config      *s3.Config
	AwsCredConfig *s3.AwsCredConfig
	MysqlDSN      string
}
