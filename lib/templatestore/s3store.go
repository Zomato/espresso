package templatestore

import (
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/Zomato/espresso/lib/s3"
)

type S3TemplateStorage struct {
	client *s3.S3Client
}

func NewS3StorageAdapter(ctx context.Context, options ...func(*s3.Config)) (*S3TemplateStorage, error) {
	s3Client, err := s3.NewS3Client(ctx, options...)
	if err != nil {
		return nil, err
	}
	return &S3TemplateStorage{client: s3Client}, nil
}

func (s *S3TemplateStorage) GetTemplate(ctx context.Context, req *GetTemplateRequest) (*template.Template, error) {
	if req.TemplateS3Path == "" {
		return nil, fmt.Errorf("template path is required for S3 storage")
	}
	reader, err := s.client.GetFileReader(ctx, req.TemplateS3Path)
	if err != nil {
		return nil, err
	}
	templateData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return template.New("template").Parse(string(templateData))
}

func (s *S3TemplateStorage) PutDocument(ctx context.Context, req *PostDocumentRequest, reader *io.Reader) (string, error) {
	if req.FileS3Path == "" {
		return "", fmt.Errorf("file S3 path is required for S3 storage")
	}
	// Upload the file to S3
	_, err := s.client.UploadFile(ctx, req.FileS3Path, *reader)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return req.FileS3Path, nil
}
func (s *S3TemplateStorage) GetDocument(ctx context.Context, req *GetDocumentRequest) (io.Reader, error) {
	if req.FileS3Path == "" {
		return nil, fmt.Errorf("file S3 path is required for S3 storage")
	}
	return s.client.GetFileReader(ctx, req.FileS3Path)
}

// ListTemplates lists all templates from S3 storage.
func (s *S3TemplateStorage) ListTemplates(ctx context.Context) ([]*TemplateInfo, error) {
	return nil, fmt.Errorf("list templates not implemented for S3 storage")
}
func (m *S3TemplateStorage) GetTemplateContent(ctx context.Context, req *GetTemplateContentRequest) (*GetTemplateContentResponse, error) {
	return nil, fmt.Errorf("get template content not implemented for S3 storage")
}
func (m *S3TemplateStorage) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (string, error) {
	return "", fmt.Errorf("create template not implemented for S3 storage")
}
