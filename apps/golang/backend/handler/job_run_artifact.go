package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type JobRunArtifactHandler struct {
	runArtifacts *usecase.JobRunArtifactService
}

func NewJobRunArtifactHandler(runArtifacts *usecase.JobRunArtifactService) *JobRunArtifactHandler {
	return &JobRunArtifactHandler{runArtifacts: runArtifacts}
}

func (h *JobRunArtifactHandler) List(w http.ResponseWriter, r *http.Request) {
	jobRunID := r.PathValue("job_run_id")
	if jobRunID == "" {
		writeError(w, http.StatusBadRequest, "missing job_run_id")
		return
	}

	artifacts, err := h.runArtifacts.ListByJobRun(r.Context(), jobRunID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if artifacts == nil {
		artifacts = []domain.JobRunArtifact{}
	}

	items := make([]openapi.JobRunArtifact, len(artifacts))
	for i := range artifacts {
		items[i] = toOpenAPIJobRunArtifact(&artifacts[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.JobRunArtifact `json:"items"`
	}{Items: items})
}

func (h *JobRunArtifactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	a, err := h.runArtifacts.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrJobRunArtifactNotFound) {
			writeError(w, http.StatusNotFound, "job run artifact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIJobRunArtifact(a))
}
