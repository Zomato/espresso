package tools

import (
	"context"

	"github.com/Zomato/espresso/service/dto"
	pdfservice "github.com/Zomato/espresso/service/service"
	svcUtils "github.com/Zomato/espresso/service/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TemplateTools struct {
	service *pdfservice.TemplateService
}

func NewTemplateTools(service *pdfservice.TemplateService) *TemplateTools {
	return &TemplateTools{service: service}
}

func (s *TemplateTools) GetAllTemplates(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, dto.GetAllTemplatesResponse, error) {
	resp, err := s.service.GetAllTemplates(ctx)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error listing templates", err, nil)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to list templates: " + err.Error()}},
			IsError: true,
		}, dto.GetAllTemplatesResponse{}, nil
	}

	return nil, *resp, nil
}

func (s *TemplateTools) GetTemplateById(ctx context.Context, _ *mcp.CallToolRequest, req dto.GetTemplateByIdRequest) (*mcp.CallToolResult, dto.GetTemplateByIdResponse, error) {
	if req.TemplateId == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "template_id is required"}},
			IsError: true,
		}, dto.GetTemplateByIdResponse{}, nil
	}

	resp, err := s.service.GetTemplateById(ctx, &req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error getting template content", err, nil)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to get template content: " + err.Error()}},
			IsError: true,
		}, dto.GetTemplateByIdResponse{}, nil
	}

	return nil, *resp, nil
}

func (s *TemplateTools) CreateTemplate(ctx context.Context, _ *mcp.CallToolRequest, req dto.CreateTemplateRequest) (*mcp.CallToolResult, dto.CreateTemplateResponse, error) {
	resp, err := s.service.CreateTemplate(ctx, &req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error creating template", err, nil)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to create template: " + err.Error()}},
			IsError: true,
		}, dto.CreateTemplateResponse{}, nil
	}

	return nil, *resp, nil
}
