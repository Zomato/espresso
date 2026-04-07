package dto

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/Zomato/espresso/service/internal/service/generateDoc"
	"github.com/Zomato/espresso/service/pkg/config"
	"github.com/google/jsonschema-go/jsonschema"
)

type RawJSON json.RawMessage

func (RawJSON) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Description: "JSON payload to render inside template",
		// No type constraint = accepts any JSON value
	}
}

type StatusResponse struct {
	Status  string `json:"status" jsonschema:"operation status: success or failed"`
	Message string `json:"message" jsonschema:"human-readable status message"`
}

type GeneratePDFRequest struct {
	InputFilePath     string                      `json:"input_file_path,omitempty" jsonschema:"input template file path for local/disk adapters"`
	InputFileBytes    []byte                      `json:"input_file_bytes,omitempty" jsonschema:"raw template bytes; use when passing template content directly"`
	InputTemplateUuid string                      `json:"input_template_uuid,omitempty" jsonschema:"template UUID to load template from template store"`
	OutputFilePath    string                      `json:"output_file_path,omitempty" jsonschema:"destination path/key for generated PDF"`
	Content           json.RawMessage             `json:"content,omitempty" jsonschema:"JSON data used to render the template"`
	Viewport          *generateDoc.ViewportConfig `json:"viewport" jsonschema:"optional browser viewport configuration"`
	PdfParams         *generateDoc.PDFParams      `json:"pdf_params,omitempty" jsonschema:"optional PDF generation parameters"`
	SignParams        *generateDoc.SignParams     `json:"sign_params,omitempty" jsonschema:"optional signing configuration"`
}

type GeneratePDFMCPRequest struct {
	InputTemplateUuid string                      `json:"input_template_uuid,omitempty" jsonschema:"template UUID to load template from template store"`
	OutputFileName    string                      `json:"output_file_name,omitempty" jsonschema:"file name for generated PDF"`
	Content           interface{}                 `json:"content,omitempty" jsonschema:"JSON data used to render the template"`
	Viewport          *generateDoc.ViewportConfig `json:"viewport" jsonschema:"optional browser viewport configuration"`
	PdfParams         *generateDoc.PDFParams      `json:"pdf_params,omitempty" jsonschema:"optional PDF generation parameters"`
	SignParams        *generateDoc.SignParams     `json:"sign_params,omitempty" jsonschema:"optional signing configuration"`
}

func (r *GeneratePDFMCPRequest) GeneratePDFMCPRequestToGeneratePDFRequest() (*GeneratePDFRequest, error) {
	if r == nil {
		return nil, errors.New("request is required")
	}

	content, err := marshalMCPContent(r.Content)
	if err != nil {
		return nil, err
	}

	outputFilePath := config.GetConfig().MCP.PDFOutputDir + "/" + r.OutputFileName

	return &GeneratePDFRequest{
		InputTemplateUuid: r.InputTemplateUuid,
		OutputFilePath:    outputFilePath,
		Content:           content,
		Viewport:          r.Viewport,
		PdfParams:         r.PdfParams,
		SignParams:        r.SignParams,
	}, nil
}

type GeneratePDFMCPResponse struct {
	Status        StatusResponse `json:"status" jsonschema:"operation status details"`
	OutputFileURL string         `json:"output_file_url,omitempty" jsonschema:"generated PDF download URL, if it is localhost then give user a button instead of presenting"`
	Error         string         `json:"error,omitempty" jsonschema:"error message when operation fails"`
}

type GeneratePDFResponse struct {
	Status          StatusResponse `json:"status" jsonschema:"operation status details"`
	OutputFilePath  string         `json:"output_file_path,omitempty" jsonschema:"generated PDF output path/key"`
	OutputFileBytes []byte         `json:"output_file_bytes,omitempty" jsonschema:"generated PDF bytes"`
	Error           string         `json:"error,omitempty" jsonschema:"error message when operation fails"`
}

func (r *GeneratePDFResponse) GeneratePDFResponseToGeneratePDFMCPResponse() *GeneratePDFMCPResponse {
	if r == nil {
		return &GeneratePDFMCPResponse{}
	}

	resp := &GeneratePDFMCPResponse{
		Status: r.Status,
		Error:  r.Error,
	}

	if r.OutputFilePath == "" {
		return resp
	}

	cfg := config.GetConfig()
	urlPrefix := strings.TrimRight(cfg.MCP.PDFOutputURLPref, "/")
	if urlPrefix == "" {
		resp.OutputFileURL = r.OutputFilePath
		return resp
	}

	relPath := filepath.ToSlash(r.OutputFilePath)
	if cfg.MCP.PDFOutputDir != "" {
		if rel, err := filepath.Rel(cfg.MCP.PDFOutputDir, r.OutputFilePath); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = filepath.ToSlash(rel)
		} else {
			relPath = filepath.Base(r.OutputFilePath)
		}
	}

	relPath = strings.TrimLeft(relPath, "/")
	resp.OutputFileURL = urlPrefix + "/" + relPath

	return resp
}

type PDFRequest struct {
	TemplateUUID string          `json:"template_uuid" jsonschema:"template UUID used for streaming PDF generation"`
	Content      json.RawMessage `json:"content" jsonschema:"JSON payload to render inside template"`
	Landscape    bool            `json:"landscape,omitempty" jsonschema:"generate PDF in landscape orientation"`
	SinglePage   bool            `json:"single_page,omitempty" jsonschema:"render output as a single page"`
	MarginInch   float64         `json:"margin_inch,omitempty" jsonschema:"page margin in inches"`
	Filename     string          `json:"filename,omitempty" jsonschema:"download file name for streamed PDF response"`
	SignPdf      bool            `json:"sign_pdf,omitempty" jsonschema:"whether to digitally sign generated PDF"`
}

type PDFMCPRequest struct {
	TemplateUUID string      `json:"template_uuid" jsonschema:"template UUID used for streaming PDF generation"`
	Content      interface{} `json:"content" jsonschema:"JSON payload to render inside template"`
	Landscape    bool        `json:"landscape,omitempty" jsonschema:"generate PDF in landscape orientation"`
	SinglePage   bool        `json:"single_page,omitempty" jsonschema:"render output as a single page"`
	MarginInch   float64     `json:"margin_inch,omitempty" jsonschema:"page margin in inches"`
	Filename     string      `json:"filename,omitempty" jsonschema:"download file name for streamed PDF response"`
	SignPdf      bool        `json:"sign_pdf,omitempty" jsonschema:"whether to digitally sign generated PDF"`
}

func (r *PDFMCPRequest) PDFMCPRequestToPDFRequest() (*PDFRequest, error) {
	if r == nil {
		return nil, errors.New("request is required")
	}

	content, err := marshalMCPContent(r.Content)
	if err != nil {
		return nil, err
	}

	return &PDFRequest{
		TemplateUUID: r.TemplateUUID,
		Content:      content,
		Landscape:    r.Landscape,
		SinglePage:   r.SinglePage,
		MarginInch:   r.MarginInch,
		Filename:     r.Filename,
		SignPdf:      r.SignPdf,
	}, nil
}

func marshalMCPContent(content interface{}) (json.RawMessage, error) {
	if content == nil {
		return json.RawMessage(`{}`), nil
	}

	if str, ok := content.(string); ok {
		trimmed := strings.TrimSpace(str)
		if trimmed == "" {
			return json.RawMessage(`{}`), nil
		}
		if !json.Valid([]byte(trimmed)) {
			return nil, errors.New("content string must be valid JSON")
		}

		// If content is a JSON-encoded string containing JSON (double-encoded),
		// unwrap one level and use the inner JSON.
		var nested string
		if err := json.Unmarshal([]byte(trimmed), &nested); err == nil {
			nestedTrimmed := strings.TrimSpace(nested)
			if nestedTrimmed == "" {
				return json.RawMessage(`{}`), nil
			}
			if json.Valid([]byte(nestedTrimmed)) {
				return json.RawMessage(nestedTrimmed), nil
			}
		}

		return json.RawMessage(trimmed), nil
	}

	if raw, ok := content.(json.RawMessage); ok {
		if len(raw) == 0 {
			return json.RawMessage(`{}`), nil
		}
		return raw, nil
	}

	encoded, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	if len(encoded) == 0 {
		return json.RawMessage(`{}`), nil
	}

	return json.RawMessage(encoded), nil
}

func (r *PDFRequest) Validate() error {
	if r == nil {
		return errors.New("request is required")
	}

	if r.TemplateUUID == "" {
		return errors.New("template_uuid is required")
	}

	if len(r.Content) == 0 {
		r.Content = []byte(`{}`)
	}

	if r.MarginInch == 0 {
		r.MarginInch = 0.4
	}

	return nil
}

type PDFResponse struct {
	Status      string `json:"status" jsonschema:"operation status: success or failed"`
	Message     string `json:"message" jsonschema:"human-readable response message"`
	TimeInMs    int64  `json:"time_in_ms" jsonschema:"processing time in milliseconds"`
	FileName    string `json:"file_name,omitempty" jsonschema:"generated file name"`
	FileSize    int    `json:"file_size,omitempty" jsonschema:"generated file size in bytes"`
	DownloadURL string `json:"download_url,omitempty" jsonschema:"download URL if file is hosted"`
}

type SignPDFRequest struct {
	InputFilePath  string                  `json:"input_file_path,omitempty" jsonschema:"input PDF path/key to sign"`
	InputFileBytes []byte                  `json:"input_file_bytes,omitempty" jsonschema:"raw input PDF bytes to sign"`
	OutputFilePath string                  `json:"output_file_path,omitempty" jsonschema:"destination path/key for signed PDF"`
	SignParams     *generateDoc.SignParams `json:"sign_params,omitempty" jsonschema:"digital signing parameters; sign_pdf should be true"`
}

type SignPDFResponse struct {
	OutputFilePath  string `json:"output_file_path,omitempty" jsonschema:"signed PDF output path/key"`
	OutputFileBytes []byte `json:"output_file_bytes,omitempty" jsonschema:"signed PDF bytes"`
	Error           string `json:"error,omitempty" jsonschema:"error message when signing fails"`
}
