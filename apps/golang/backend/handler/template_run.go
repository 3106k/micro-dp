package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type TemplateRunHandler struct {
	templateRuns *usecase.TemplateRunService
}

func NewTemplateRunHandler(templateRuns *usecase.TemplateRunService) *TemplateRunHandler {
	return &TemplateRunHandler{templateRuns: templateRuns}
}

func (h *TemplateRunHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateTemplateRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TemplateType == "" {
		writeError(w, http.StatusBadRequest, "template_type is required")
		return
	}

	tr, err := h.templateRuns.Create(r.Context(), req.TemplateType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toOpenAPITemplateRun(tr))
}

func (h *TemplateRunHandler) List(w http.ResponseWriter, r *http.Request) {
	runs, err := h.templateRuns.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if runs == nil {
		runs = []domain.TemplateRun{}
	}
	items := make([]openapi.TemplateRun, len(runs))
	for i := range runs {
		items[i] = toOpenAPITemplateRun(&runs[i])
	}
	writeJSON(w, http.StatusOK, struct {
		Items []openapi.TemplateRun `json:"items"`
	}{Items: items})
}

func (h *TemplateRunHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	tr, err := h.templateRuns.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateRunNotFound) {
			writeError(w, http.StatusNotFound, "template run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPITemplateRun(tr))
}
