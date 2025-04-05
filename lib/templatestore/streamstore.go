package templatestore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/template"
)

type StreamStorage struct {
}

func (s *StreamStorage) GetTemplate(ctx context.Context, req *GetTemplateRequest) (*template.Template, error) {
	if req.TemplateBytes == nil {
		return nil, fmt.Errorf("input template stream is required for stream storage")
	}

	return template.New("stream").Parse(string(req.TemplateBytes))
}
func (s *StreamStorage) PutDocument(ctx context.Context, req *PostDocumentRequest, reader *io.Reader) (string, error) {
	// Read all bytes from the rod.StreamReader
	pdfBytes, err := io.ReadAll(*reader)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF stream: %v", err)
	}

	// Store bytes in the request for later retrieval
	req.OutputFileBytes = pdfBytes
	return "stream", nil
}
func (s *StreamStorage) GetDocument(ctx context.Context, req *GetDocumentRequest) (io.Reader, error) {
	if req.InputFileBytes == nil {
		return nil, fmt.Errorf("input file bytes are required for stream storage")
	}

	return bytes.NewReader(req.InputFileBytes), nil
}

// ListTemplates returns an error for stream storage since it doesn't support listing templates.
func (s *StreamStorage) ListTemplates(ctx context.Context) ([]*TemplateInfo, error) {
	return nil, fmt.Errorf("listing templates is not supported for stream storage")
}

func (m *StreamStorage) GetTemplateContent(ctx context.Context, req *GetTemplateContentRequest) (*GetTemplateContentResponse, error) {
	return nil, fmt.Errorf("get template content not implemented for stream storage")
}

func (m *StreamStorage) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (string, error) {
	return "", fmt.Errorf("create template not implemented for stream storage")
}
