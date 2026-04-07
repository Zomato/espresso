package controller

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Zomato/espresso/service/dto"
	"github.com/Zomato/espresso/service/pkg/response"
	"github.com/Zomato/espresso/service/service"
	svcUtils "github.com/Zomato/espresso/service/utils"
)

type PDFController struct {
	service *service.PDFService
}

func NewPDFController(service *service.PDFService) *PDFController {
	return &PDFController{service: service}
}

func (s *PDFController) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &dto.GeneratePDFRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body", err, nil)
		response.RespondWithError(w, "Failed to parse JSON request", http.StatusBadRequest)
		return
	}

	resp, err := s.service.GeneratePDF(ctx, req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf", err, nil)
		response.RespondWithError(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *PDFController) GeneratePDFStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.PDFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body", err, nil)
		response.RespondWithError(w, "Failed to parse JSON request", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		response.RespondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.service.GeneratePDFStream(ctx, &req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf stream", err, nil)
		response.RespondWithError(w, "Failed to generate PDF stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := "generated.pdf"
	if req.Filename != "" {
		fileName = req.Filename
		if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
			fileName += ".pdf"
		}
	}
	fileName = filepath.Base(fileName)

	if len(resp.OutputFileBytes) == 0 {
		response.RespondWithError(w, "No PDF data available", http.StatusInternalServerError)
		return
	}

	if err = response.RespondWithFile(w, fileName, "application/pdf", resp.OutputFileBytes); err != nil {
		svcUtils.Logger.Error(ctx, "error writing pdf stream", err, nil)
		response.RespondWithError(w, "Failed to write PDF stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *PDFController) SignPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &dto.SignPDFRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body", err, nil)
		response.RespondWithError(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.service.SignPDF(ctx, req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in signing pdf", err, nil)
		response.RespondWithError(w, "Failed to sign PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "PDF signed successfully",
		},
		"output_file_path":  resp.OutputFilePath,
		"output_file_bytes": resp.OutputFileBytes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}
