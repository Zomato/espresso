package templatestore

import "time"

type GetTemplateRequest struct {
	TemplateUUID   string
	TemplatePath   string
	TemplateS3Path string
	TemplateBytes  []byte
}
type PostDocumentRequest struct {
	FilePath        string
	FileS3Path      string
	OutputFileBytes []byte
}
type GetDocumentRequest struct {
	FilePath        string
	FileS3Path      string
	InputFileBytes  []byte
	OutputFileBytes []byte
}

// TemplateInfo contains metadata about a template
type TemplateInfo struct {
	TemplateID   string    `json:"template_id"`
	TemplateName string    `json:"template_name,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
type GetTemplateContentRequest struct {
	TemplateUUID string
}
type GetTemplateContentResponse struct {
	TemplateContent    string `json:"template_content"`
	TemplateName       string `json:"template_name,omitempty"`
	TemplateJsonSchema string `json:"template_json_schema,omitempty"`
}
type CreateTemplateRequest struct {
	TemplateName string
	TemplateHTML string
	TemplateJSON string
}
