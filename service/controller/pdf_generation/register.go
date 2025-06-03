package pdf_generation

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Zomato/espresso/lib/certmanager"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/service/internal/pkg/config"
)

type EspressoService struct {
	TemplateStorageAdapter *templatestore.StorageAdapter
	FileStorageAdapter     *templatestore.StorageAdapter
	CredentialStore        *certmanager.CredentialStore
}

func NewEspressoService(config config.Config) (*EspressoService, error) {
	templateStorageType := config.TemplateStorageConfig.StorageType

	if config.AppConfig.EnableUI && templateStorageType != templatestore.StorageAdapterTypeMySQL {
		return nil, fmt.Errorf("UI requires MySQL as template storage adapter, got: %s", templateStorageType)
	}

	mySqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.DBConfig.Username,
		config.DBConfig.Password,
		config.DBConfig.Host,
		config.DBConfig.Port,
		config.DBConfig.Database,
	)

	templateStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType:   templateStorageType,
		S3Config:      &config.S3Config,  // for s3 storage only
		AwsCredConfig: &config.AWSConfig, // for s3 storage only
		MysqlDSN:      mySqlDsn,          // for mysql adapter
	})
	if err != nil {
		return nil, err
	}

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType: config.FileStorageConfig.StorageType,
	})
	if err != nil {
		return nil, err
	}

	credentialStore, err := certmanager.NewCredentialStore(config.CertConfig)
	if err != nil {
		return nil, err
	}

	return &EspressoService{
		TemplateStorageAdapter: &templateStorageAdapter,
		FileStorageAdapter:     &fileStorageAdapter,
		CredentialStore:        credentialStore,
	}, nil
}
func Register(mux *http.ServeMux, config config.Config) {
	espressoService, err := NewEspressoService(config)
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
