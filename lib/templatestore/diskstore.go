package templatestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

// DiskTemplateStorage is a concrete implementation of TemplateStorageAdapter for disk storage.
type DiskTemplateStorage struct {
	// disk storage implementation
}

func (d *DiskTemplateStorage) GetTemplate(ctx context.Context, req *GetTemplateRequest) (*template.Template, error) {
	if req.TemplatePath == "" {
		return nil, fmt.Errorf("template path is required for disk storage")
	}
	// get template from filepath
	templatePath := req.TemplatePath
	templateFile, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse template file: %v", err)
	}

	return templateFile, nil
	// implement disk storage retrieval
}

func (d *DiskTemplateStorage) PutDocument(ctx context.Context, req *PostDocumentRequest, reader *io.Reader) (string, error) {
	if req.FilePath == "" {
		return "", fmt.Errorf("file path is required for disk storage")
	}
	// Create directories if they don't exist
	dir := filepath.Dir(req.FilePath)
	// make directory from req.filepath, dont append output
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	file, err := os.Create(req.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy the stream to the file
	if _, err := io.Copy(file, *reader); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return req.FilePath, nil
}
func (d *DiskTemplateStorage) GetDocument(ctx context.Context, req *GetDocumentRequest) (io.Reader, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("file path is required for disk storage")
	}
	// Open the file for reading
	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	return file, nil
}

// ListTemplates lists all templates from disk storage.
func (d *DiskTemplateStorage) ListTemplates(ctx context.Context) ([]*TemplateInfo, error) {
	return nil, fmt.Errorf("listing templates is not supported for disk storage")
}
func (m *DiskTemplateStorage) GetTemplateContent(ctx context.Context, req *GetTemplateContentRequest) (*GetTemplateContentResponse, error) {
	return nil, fmt.Errorf("get template content not implemented for disk storage")
}
func (m *DiskTemplateStorage) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (string, error) {
	return "", fmt.Errorf("create template not implemented for disk storage")
}
