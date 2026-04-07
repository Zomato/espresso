package dto

import "github.com/Zomato/espresso/service/internal/service/generateDoc"

type GetAllTemplatesResponse struct {
	Status       StatusResponse                  `json:"status" jsonschema:"operation status details"`
	TotalRecords int32                           `json:"total_records,omitempty" jsonschema:"number of templates returned"`
	Data         []*generateDoc.TemplateListData `json:"data,omitempty" jsonschema:"list of available templates"`
	Error        string                          `json:"error,omitempty" jsonschema:"error message when listing fails"`
}

type GetTemplateByIdRequest struct {
	TemplateId string `json:"template_id" jsonschema:"template UUID to fetch"`
}

type GetTemplateByIdResponse struct {
	Status       StatusResponse `json:"status" jsonschema:"operation status details"`
	TemplateHtml string         `json:"template_html" jsonschema:"template HTML content"`
	Json         string         `json:"json" jsonschema:"template JSON schema/content"`
	TemplateName string         `json:"template_name" jsonschema:"template display name"`
	Error        string         `json:"error,omitempty" jsonschema:"error message when fetch fails"`
}

type CreateTemplateRequest struct {
	TemplateName string `json:"template_name" jsonschema:"template display name"`
	TemplateHtml string `json:"template_html" jsonschema:"template HTML content"`
	Json         string `json:"json" jsonschema:"template JSON schema/content; defaults to {}"`
}

type CreateTemplateResponse struct {
	Status     StatusResponse `json:"status" jsonschema:"operation status details"`
	TemplateId string         `json:"template_id" jsonschema:"created template UUID"`
	Error      string         `json:"error,omitempty" jsonschema:"error message when creation fails"`
}
