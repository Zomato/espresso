package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Zomato/espresso/lib/templatestore"
	libutils "github.com/Zomato/espresso/lib/utils"
	"github.com/Zomato/espresso/service/dto"
	"github.com/Zomato/espresso/service/internal/service/generateDoc"
	svcUtils "github.com/Zomato/espresso/service/utils"
)

type TemplateService struct {
	TemplateStorageAdapter *templatestore.StorageAdapter
}

func NewTemplateService(templateAdapter templatestore.StorageAdapter) *TemplateService {
	return &TemplateService{TemplateStorageAdapter: &templateAdapter}
}

func (s *TemplateService) GetAllTemplates(ctx context.Context) (*dto.GetAllTemplatesResponse, error) {
	startTime := time.Now()
	reqID := libutils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GetAllTemplates called :: ", map[string]any{"req_id": reqID})

	templates, err := (*s.TemplateStorageAdapter).ListTemplates(ctx)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error listing templates", err, nil)
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	templateDataList := make([]*generateDoc.TemplateListData, 0, len(templates))
	for _, tmpl := range templates {
		createdAt := ""
		if !tmpl.CreatedAt.IsZero() {
			createdAt = tmpl.CreatedAt.Format(time.RFC3339)
		}

		updatedAt := ""
		if !tmpl.UpdatedAt.IsZero() {
			updatedAt = tmpl.UpdatedAt.Format(time.RFC3339)
		}

		templateDataList = append(templateDataList, &generateDoc.TemplateListData{
			TemplateId:   tmpl.TemplateID,
			TemplateName: tmpl.TemplateName,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		})
	}

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "listed templates :: ", map[string]any{"length": len(templateDataList), "duration": duration})

	return &dto.GetAllTemplatesResponse{
		Status: dto.StatusResponse{
			Status:  "success",
			Message: "Templates retrieved successfully",
		},
		TotalRecords: int32(len(templateDataList)),
		Data:         templateDataList,
	}, nil
}

func (s *TemplateService) GetTemplateById(ctx context.Context, req *dto.GetTemplateByIdRequest) (*dto.GetTemplateByIdResponse, error) {
	templateData, err := (*s.TemplateStorageAdapter).GetTemplateContent(ctx, &templatestore.GetTemplateContentRequest{
		TemplateUUID: req.TemplateId,
	})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error getting template content", err, nil)
		return nil, fmt.Errorf("failed to get template content: %w", err)
	}

	return &dto.GetTemplateByIdResponse{
		Status: dto.StatusResponse{
			Status:  "success",
			Message: "Template content retrieved successfully",
		},
		TemplateHtml: templateData.TemplateContent,
		TemplateName: templateData.TemplateName,
		Json:         templateData.TemplateJsonSchema,
	}, nil
}

func (s *TemplateService) CreateTemplate(ctx context.Context, req *dto.CreateTemplateRequest) (*dto.CreateTemplateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}
	if req.TemplateName == "" {
		return nil, fmt.Errorf("template name is required")
	}
	if req.TemplateHtml == "" {
		return nil, fmt.Errorf("template html is required")
	}

	jsonSchema := req.Json
	if jsonSchema == "" {
		jsonSchema = "{}"
	}

	templateID, err := (*s.TemplateStorageAdapter).CreateTemplate(ctx, &templatestore.CreateTemplateRequest{
		TemplateName: req.TemplateName,
		TemplateHTML: req.TemplateHtml,
		TemplateJSON: jsonSchema,
	})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error creating template", err, nil)
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &dto.CreateTemplateResponse{
		Status: dto.StatusResponse{
			Status:  "success",
			Message: "Template created successfully",
		},
		TemplateId: templateID,
	}, nil
}
