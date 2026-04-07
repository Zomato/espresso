package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Zomato/espresso/service/dto"
	"github.com/Zomato/espresso/service/pkg/response"
	"github.com/Zomato/espresso/service/service"
	svcUtils "github.com/Zomato/espresso/service/utils"
)

type TemplateController struct {
	service *service.TemplateService
}

func NewTemplateController(service *service.TemplateService) *TemplateController {
	return &TemplateController{service: service}
}

func (s *TemplateController) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := s.service.GetAllTemplates(ctx)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error listing templates", err, nil)
		response.RespondWithError(w, "Failed to list templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *TemplateController) GetTemplateById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	templateID := r.URL.Query().Get("template_id")
	if templateID == "" {
		svcUtils.Logger.Error(ctx, "template id is required", nil, nil)
		response.RespondWithError(w, "template id is required", http.StatusBadRequest)
		return
	}

	resp, err := s.service.GetTemplateById(ctx, &dto.GetTemplateByIdRequest{TemplateId: templateID})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error getting template content", err, nil)
		response.RespondWithError(w, "Failed to get template content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *TemplateController) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &dto.CreateTemplateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		svcUtils.Logger.Error(ctx, "error decoding request body", err, nil)
		response.RespondWithError(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.service.CreateTemplate(ctx, req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error creating template", err, nil)
		response.RespondWithError(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"status": map[string]string{
			"status":  "success",
			"message": "Template created successfully",
		},
		"template_id": resp.TemplateId,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}
