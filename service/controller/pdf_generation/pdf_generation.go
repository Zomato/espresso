package pdf_generation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/lib/utils"
	"github.com/Zomato/espresso/service/internal/pkg/httppkg"
	"github.com/Zomato/espresso/service/internal/service/generateDoc"
	svcUtils "github.com/Zomato/espresso/service/utils"
)

func (s *EspressoService) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()
	req := &GeneratePDFRequest{}

	// Read and parse the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		svcUtils.Logger.Error(ctx, "Error reading request body: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the request JSON
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		svcUtils.Logger.Error(ctx, "Error parsing JSON: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to parse JSON request", http.StatusBadRequest)
		return
	}

	reqId := utils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GeneratePDF called :: ", map[string]any{"req_id": reqId})

	generatePdfReq := &generateDoc.PDFDto{
		ReqId:              reqId,
		InputTemplatePath:  req.InputFilePath,
		InputFileBytes:     req.InputFileBytes,
		InputTemplateUUID:  req.InputTemplateUuid,
		OutputTemplatePath: req.OutputFilePath,
		Content:            req.Content,
		ViewPort:           req.Viewport,
		PdfParams:          req.PdfParams,
	}

	if req.SignParams != nil && req.SignParams.SignPdf {
		generatePdfReq.SignParams = req.SignParams
	}

	err = generateDoc.GeneratePDF(ctx, generatePdfReq, s.TemplateStorageAdapter, s.FileStorageAdapter, s.CredentialStore)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf :: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "PDF generated successfully",
		},
		"output_file_path":  req.OutputFilePath,
		"output_file_bytes": generatePdfReq.OutputFileBytes,
	}

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "generated pdf :: ", map[string]any{"req_id": reqId, "duration": duration})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}

func (s *EspressoService) GeneratePDFStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()
	reqId := utils.GenerateUniqueID(ctx)

	svcUtils.Logger.Info(ctx, "GeneratePDFStream called :: ", map[string]any{"req_id": reqId})

	// Read and parse the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		svcUtils.Logger.Error(ctx, "Error reading request body: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the request JSON
	var pdfReq PDFRequest
	if err := json.Unmarshal(bodyBytes, &pdfReq); err != nil {
		svcUtils.Logger.Error(ctx, "Error parsing JSON: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to parse JSON request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if pdfReq.TemplateUUID == "" {
		httppkg.RespondWithError(w, "template_uuid is required", http.StatusBadRequest)
		return
	}

	// If content is empty or invalid, use an empty object as default
	if len(pdfReq.Content) == 0 {
		pdfReq.Content = json.RawMessage(`{}`)
	}

	// Set default margin if not provided
	margin := pdfReq.MarginInch
	if margin == 0 {
		margin = 0.4 // Default margin of 0.4 inches
	}
	// // Set up PDF parameters, add your own parameters from request if needed
	pdfSettings := &generateDoc.PDFParams{
		Landscape:           pdfReq.Landscape,
		DisplayHeaderFooter: false,
		PrintBackground:     true,
		PreferCssPageSize:   false,
		MarginTop:           margin,
		MarginBottom:        margin,
		MarginLeft:          margin,
		MarginRight:         margin,
		IsSinglePage:        pdfReq.SinglePage,
	}

	generatePdfReq := &generateDoc.PDFDto{
		ReqId:             reqId,
		InputTemplateUUID: pdfReq.TemplateUUID,
		Content:           pdfReq.Content,
		SignParams:        &generateDoc.SignParams{SignPdf: pdfReq.SignPdf},
		// ViewPort:          req.Viewport,
		PdfParams: pdfSettings,
	}
	if pdfReq.SignPdf {
		generatePdfReq.SignParams = &generateDoc.SignParams{
			SignPdf:       true,
			CertConfigKey: "cert1", // certificate details are stored in config file
		}
	}

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType: "stream",
	})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in getting file storage adapter :: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to get file storage adapter: "+err.Error(), http.StatusExpectationFailed)
		return
	}
	
	err = generateDoc.GeneratePDF(ctx, generatePdfReq, s.TemplateStorageAdapter, &fileStorageAdapter, s.CredentialStore)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf stream:: %v", err, nil)
		httppkg.RespondWithError(w, "Failed to generate PDF stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Determine filename for the PDF
	fileName := "generated.pdf"

	// Use the filename from the request if provided
	if pdfReq.Filename != "" {
		fileName = pdfReq.Filename
		// Ensure it has .pdf extension
		if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
			fileName += ".pdf"
		}
	}

	// Sanitize filename (remove any path elements for security)
	fileName = filepath.Base(fileName)

	// Check if we have PDF data to return
	if len(generatePdfReq.OutputFileBytes) > 0 {
		// Always return the PDF file directly for download
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(generatePdfReq.OutputFileBytes)))
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)

		// Write the PDF data
		_, err = w.Write(generatePdfReq.OutputFileBytes)
		if err != nil {
			svcUtils.Logger.Error(ctx, "error writing pdf stream :: %v", err, nil)
			httppkg.RespondWithError(w, "Failed to write PDF stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		duration := time.Since(startTime)
		svcUtils.Logger.Info(ctx, "generated pdf stream :: ", map[string]any{"req_id": reqId, "duration": duration})

		return
	} else {
		// If no PDF data, return an error
		httppkg.RespondWithError(w, "No PDF data available", http.StatusInternalServerError)
		return
	}

}

func (s *EspressoService) SignPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &SignPDFRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body :: %v", err, nil)
		httppkg.RespondWithError(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	reqId := utils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GeneratePDF called :: ", map[string]any{"req_id": reqId})

	signPDFDto := &generateDoc.SignPDFDto{
		ReqId:          reqId,
		InputFilePath:  req.InputFilePath,
		InputFileBytes: req.InputFileBytes,
		OutputFilePath: req.OutputFilePath,
	}
	if req.SignParams != nil && req.SignParams.SignPdf {
		signPDFDto.SignParams = req.SignParams
	} else {
		err := fmt.Errorf("signPdf param is not true in the request")
		svcUtils.Logger.Error(ctx, "error in signing pdf :: : %v", err, nil)

		httppkg.RespondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := generateDoc.SignPDF(ctx, signPDFDto, s.FileStorageAdapter, s.CredentialStore)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in signing pdf :: : %v", err, nil)
		httppkg.RespondWithError(w, "Failed to sign PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "PDF signed successfully",
		},
		"output_file_path":  req.OutputFilePath,
		"output_file_bytes": signPDFDto.OutputFileBytes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}

func (s *EspressoService) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	reqId := utils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GetAllTemplates called :: ", map[string]any{"req_id": reqId})

	// Get templates from the storage adapter
	templates, err := (*s.TemplateStorageAdapter).ListTemplates(ctx)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error listing templates :: : %v", err, nil)
		httppkg.RespondWithError(w, "Failed to list templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert internal template info to protobuf format
	var templateDataList []*generateDoc.TemplateListData
	for _, tmpl := range templates {
		createdAt := ""
		if !tmpl.CreatedAt.IsZero() {
			createdAt = tmpl.CreatedAt.Format(time.RFC3339)
		}

		updatedAt := ""
		if !tmpl.UpdatedAt.IsZero() {
			updatedAt = tmpl.UpdatedAt.Format(time.RFC3339)
		}
		templateData := &generateDoc.TemplateListData{
			TemplateId:   tmpl.TemplateID,
			TemplateName: tmpl.TemplateName,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}

		templateDataList = append(templateDataList, templateData)
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "Templates retrieved successfully",
		},
		"total_records": len(templateDataList),
		"data":          templateDataList,
	}

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "listed templates :: ", map[string]any{"length": len(templateDataList), "duration": duration})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}
func (s *EspressoService) GetTemplateById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	templateID := r.URL.Query().Get("template_id")
	if templateID == "" {
		svcUtils.Logger.Error(ctx, "template id is required ", nil, nil)
		httppkg.RespondWithError(w, "template id is required", http.StatusBadRequest)
		return
	}

	templateData, err := (*s.TemplateStorageAdapter).GetTemplateContent(ctx, &templatestore.GetTemplateContentRequest{
		TemplateUUID: templateID,
	})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error getting template content :: : %v", err, nil)
		httppkg.RespondWithError(w, "Failed to get template content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "Template retrieved successfully",
		},
		"template_html": templateData.TemplateContent,
		"template_name": templateData.TemplateName,
		"json":          templateData.TemplateJsonSchema,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}
func (s *EspressoService) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	req := &CreateTemplateRequest{}
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body :: : %v", err, nil)

		httppkg.RespondWithError(w, "Error decoding request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Validate request
	if req.TemplateName == "" {
		svcUtils.Logger.Error(ctx, "template name is required ", nil, nil)
		httppkg.RespondWithError(w, "Template name is required", http.StatusBadRequest)
		return
	}

	if req.TemplateHtml == "" {
		svcUtils.Logger.Error(ctx, "template html is required ", nil, nil)
		httppkg.RespondWithError(w, "Template HTML is required", http.StatusBadRequest)
		return
	}

	// Default JSON schema to empty object if not provided
	jsonSchema := req.Json
	if jsonSchema == "" {
		jsonSchema = "{}"
	}

	// Create template using the storage adapter
	createReq := &templatestore.CreateTemplateRequest{
		TemplateName: req.TemplateName,
		TemplateHTML: req.TemplateHtml,
		TemplateJSON: jsonSchema,
	}

	templateId, err := (*s.TemplateStorageAdapter).CreateTemplate(ctx, createReq)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error creating template :: : %v", err, nil)
		httppkg.RespondWithError(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// response.TemplateId = templateId

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "Template created successfully",
		},
		"template_id": templateId,
	}
	// Return success response
	w.WriteHeader(http.StatusCreated) // 201 Created is more appropriate
	json.NewEncoder(w).Encode(responseData)

}
