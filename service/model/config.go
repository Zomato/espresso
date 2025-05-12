package model

type Config struct {
	AppConfig             AppConfig             `mapstructure:"app"`
	TemplateStorageConfig StorageConfig         `mapstructure:"template_storage"`
	FileStorageConfig     StorageConfig         `mapstructure:"file_storage"`
	BrowserConfig         BrowserConfig         `mapstructure:"browser"`
	WorkerPoolConfig      WorkerPoolConfig      `mapstructure:"workerpool"`
	S3Config              S3Config              `mapstructure:"s3"`
	AWSConfig             AWSConfig             `mapstructure:"aws"`
	CertConfig            map[string]CertConfig `mapstructure:"certificates"`
	DBConfig              DBConfig              `mapstructure:"db"`
}

type AppConfig struct {
	LogLevel      string `mapstructure:"log_level"`
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

type S3Config struct {
	Endpoint              string `mapstructure:"endpoint"`
	Debug                 bool   `mapstructure:"debug"`
	Region                string `mapstructure:"region"`
	ForcePathStyle        bool   `mapstructure:"force_path_style"`
	UploaderConcurrency   int    `mapstructure:"uploader_concurrency"`
	UploaderPartSizeMB    int64  `mapstructure:"uploader_part_size_mb"`
	DownloaderConcurrency int    `mapstructure:"downloader_concurrency"`
	DownloaderPartSizeMB  int64  `mapstructure:"downloader_part_size_mb"`
	RetryMaxAttempts      int    `mapstructure:"retry_max_attempts"`
	Bucket                string `mapstructure:"bucket"`
	UseCustomTransport    bool   `mapstructure:"use_custom_transport"`
}

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	SessionToken    string `mapstructure:"session_token"`
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
