package pdf_generation

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zomato/espresso/lib/s3"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/spf13/viper"
)

type EspressoService struct {
	TemplateStorageAdapter *templatestore.StorageAdapter
	FileStorageAdapter     *templatestore.StorageAdapter
}

func NewEspressoService() (*EspressoService, error) {
	templateStorageType := viper.GetString("template_storage.storage_type")
	if os.Getenv("ENABLE_UI") == "true" && templateStorageType != templatestore.StorageAdapterTypeMySQL {
		return nil, fmt.Errorf("UI requires MySQL as template storage adapter, got: %s", templateStorageType)
	}
	templateStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType: templateStorageType,
		// for s3 storage only
		S3Config: &s3.Config{
			Endpoint:              viper.GetString("s3.endpoint"),
			Region:                viper.GetString("s3.region"),
			Bucket:                viper.GetString("s3.bucket"),
			Debug:                 viper.GetBool("s3.debug"),
			ForcePathStyle:        viper.GetBool("s3.forcePathStyle"),
			UploaderConcurrency:   viper.GetInt("s3.uploaderConcurrency"),
			UploaderPartSize:      viper.GetInt64("s3.uploaderPartSize"),
			DownloaderConcurrency: viper.GetInt("s3.downloaderConcurrency"),
			DownloaderPartSize:    viper.GetInt64("s3.downloaderPartSize"),
			RetryMaxAttempts:      viper.GetInt("s3.retryMaxAttempts"),
			UseCustomTransport:    viper.GetBool("s3.useCustomTransport"),
		},
		// for s3 storage only
		AwsCredConfig: &s3.AwsCredConfig{
			AccessKeyID:     viper.GetString("aws.accessKeyID"),
			SecretAccessKey: viper.GetString("aws.secretAccessKey"),
			SessionToken:    viper.GetString("aws.sessionToken"),
		},
		MysqlDSN: viper.GetString("mysql.dsn"), // for mysql adapter
	})
	if err != nil {
		return nil, err
	}

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType: viper.GetString("file_storage.storage_type"),
	})
	if err != nil {
		return nil, err
	}

	return &EspressoService{TemplateStorageAdapter: &templateStorageAdapter, FileStorageAdapter: &fileStorageAdapter}, nil
}
func Register(mux *http.ServeMux) {
	espressoService, err := NewEspressoService()
	if err != nil {
		log.Fatalf("Failed to initialize PDF service: %v", err)
	}

	// Register HTTP routes
	// Register handlers with the mux
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/generate-pdf-stream", espressoService.GeneratePDFStream)
	mux.HandleFunc("/create-template", espressoService.CreateTemplate)
	mux.HandleFunc("/list-templates", espressoService.GetAllTemplates)
	mux.HandleFunc("/get-template", espressoService.GetTemplateById)
	mux.HandleFunc("/generate-pdf", espressoService.GeneratePDF)

}
