package server

import (
	"log"
	"net/http"

	controller "github.com/Zomato/espresso/service/controller/http"
	"github.com/Zomato/espresso/service/service"
)

func RegisterHTTP(mux *http.ServeMux) {
	templateAdapter, fileAdapter, err := initStorageAdapters()
	if err != nil {
		log.Fatalf("Failed to initialize storage adapters: %v", err)
	}

	pdfController := controller.NewPDFController(service.NewPDFService(templateAdapter, fileAdapter))
	templateController := controller.NewTemplateController(service.NewTemplateService(templateAdapter))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/generate-pdf-stream", pdfController.GeneratePDFStream)
	mux.HandleFunc("/generate-pdf", pdfController.GeneratePDF)
	mux.HandleFunc("/create-template", templateController.CreateTemplate)
	mux.HandleFunc("/list-templates", templateController.GetAllTemplates)
	mux.HandleFunc("/get-template", templateController.GetTemplateById)
}
