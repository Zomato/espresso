package config

import (
	"github.com/Zomato/espresso/lib/s3"
)

type Config struct {
	AppConfig             AppConfig             `mapstructure:"app"`
	TemplateStorageConfig StorageConfig         `mapstructure:"template_storage"`
	FileStorageConfig     StorageConfig         `mapstructure:"file_storage"`
	BrowserConfig         BrowserConfig         `mapstructure:"browser"`
	WorkerPoolConfig      WorkerPoolConfig      `mapstructure:"workerpool"`
	S3Config              s3.Config             `mapstructure:"s3"`
	AWSConfig             s3.AwsCredConfig      `mapstructure:"aws"`
	CertConfig            map[string]CertConfig `mapstructure:"certificates"`
	DBConfig              DBConfig              `mapstructure:"db"`
}

type AppConfig struct {
	LogLevel      string `mapstructure:"log_level"`
	ServerPort    int    `mapstructure:"server_port"`
	EnableUI      bool   `mapstructure:"enable_ui"`
	RodBrowserBin string `mapstructure:"rod_browser_bin"`
}

type StorageConfig struct {
	StorageType string `mapstructure:"storage_type"`
}

type BrowserConfig struct {
	TabPool int `mapstructure:"tab_pool"`
}

type WorkerPoolConfig struct {
	WorkerCount     int `mapstructure:"worker_count"`
	WorkerTimeoutMs int `mapstructure:"worker_timeout_ms"`
}

type CertConfig struct {
	CertPath    string `mapstructure:"cert_path"`
	KeyPath     string `mapstructure:"key_path"`
	KeyPassword string `mapstructure:"key_password"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}
