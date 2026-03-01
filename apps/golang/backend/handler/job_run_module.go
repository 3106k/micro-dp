package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type JobRunModuleHandler struct {
	runModules *usecase.JobRunModuleService
}

func NewJobRunModuleHandler(runModules *usecase.JobRunModuleService) *JobRunModuleHandler {
	return &JobRunModuleHandler{runModules: runModules}
}

func (h *JobRunModuleHandler) List(w http.ResponseWriter, r *http.Request) {
	jobRunID := r.PathValue("job_run_id")
	if jobRunID == "" {
		writeError(w, http.StatusBadRequest, "missing job_run_id")
		return
	}

	modules, err := h.runModules.ListByJobRun(r.Context(), jobRunID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if modules == nil {
		modules = []domain.JobRunModule{}
	}

	items := make([]openapi.JobRunModule, len(modules))
	for i := range modules {
		items[i] = toOpenAPIJobRunModule(&modules[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.JobRunModule `json:"items"`
	}{Items: items})
}

func (h *JobRunModuleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	m, err := h.runModules.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrJobRunModuleNotFound) {
			writeError(w, http.StatusNotFound, "job run module not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIJobRunModule(m))
}
