package config

type Config struct {
	MCP                 MCPConfig                         `mapstructure:"mcp" yaml:"mcp"`
	TemplateStorage     StorageConfig                     `mapstructure:"template_storage" yaml:"template_storage"`
	FileStorage         StorageConfig                     `mapstructure:"file_storage" yaml:"file_storage"`
	Browser             BrowserConfig                     `mapstructure:"browser" yaml:"browser"`
	WorkerPool          WorkerPoolConfig                  `mapstructure:"workerpool" yaml:"workerpool"`
	S3                  S3Config                          `mapstructure:"s3" yaml:"s3"`
	AWS                 AWSConfig                         `mapstructure:"aws" yaml:"aws"`
	DigitalCertificates map[string]DigitalCertificateSpec `mapstructure:"digital_certificates" yaml:"digital_certificates"`
	MySQL               MySQLConfig                       `mapstructure:"mysql" yaml:"mysql"`
}

type MCPConfig struct {
	Enabled          bool   `mapstructure:"enabled" yaml:"enabled"`
	PDFOutputDir     string `mapstructure:"pdf_output_dir" yaml:"pdf_output_dir"`
	PDFOutputURLPref string `mapstructure:"pdf_output_url_prefix" yaml:"pdf_output_url_prefix"`
	PDFOutputPath    string `mapstructure:"pdf_output_path" yaml:"pdf_output_path"`
}

type StorageConfig struct {
	StorageType string `mapstructure:"storage_type" yaml:"storage_type"`
}

type BrowserConfig struct {
	TabPool int `mapstructure:"tab_pool" yaml:"tab_pool"`
}

type WorkerPoolConfig struct {
	WorkerCount   int `mapstructure:"worker_count" yaml:"worker_count"`
	WorkerTimeout int `mapstructure:"worker_timeout" yaml:"worker_timeout"`
}

type S3Config struct {
	Endpoint              string `mapstructure:"endpoint" yaml:"endpoint"`
	Debug                 bool   `mapstructure:"debug" yaml:"debug"`
	Region                string `mapstructure:"region" yaml:"region"`
	ForcePathStyle        bool   `mapstructure:"forcePathStyle" yaml:"forcePathStyle"`
	UploaderConcurrency   int    `mapstructure:"uploaderConcurrency" yaml:"uploaderConcurrency"`
	UploaderPartSize      int64  `mapstructure:"uploaderPartSize" yaml:"uploaderPartSize"`
	DownloaderConcurrency int    `mapstructure:"downloaderConcurrency" yaml:"downloaderConcurrency"`
	DownloaderPartSize    int64  `mapstructure:"downloaderPartSize" yaml:"downloaderPartSize"`
	RetryMaxAttempts      int    `mapstructure:"retryMaxAttempts" yaml:"retryMaxAttempts"`
	Bucket                string `mapstructure:"bucket" yaml:"bucket"`
	UseCustomTransport    bool   `mapstructure:"useCustomTransport" yaml:"useCustomTransport"`
}

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyID" yaml:"accessKeyID"`
	SecretAccessKey string `mapstructure:"secretAccessKey" yaml:"secretAccessKey"`
	SessionToken    string `mapstructure:"sessionToken" yaml:"sessionToken"`
}

type DigitalCertificateSpec struct {
	CertFilePath string `mapstructure:"cert_filepath" yaml:"cert_filepath"`
	KeyFilePath  string `mapstructure:"key_filepath" yaml:"key_filepath"`
	KeyPassword  string `mapstructure:"key_password" yaml:"key_password"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}
