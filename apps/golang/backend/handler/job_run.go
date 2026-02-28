package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/usecase"
)

type JobRunHandler struct {
	jobRuns *usecase.JobRunService
}

func NewJobRunHandler(jobRuns *usecase.JobRunService) *JobRunHandler {
	return &JobRunHandler{jobRuns: jobRuns}
}

func (h *JobRunHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string `json:"project_id"`
		JobID     string `json:"job_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ProjectID == "" || req.JobID == "" {
		writeError(w, http.StatusBadRequest, "project_id and job_id are required")
		return
	}

	jr, err := h.jobRuns.Create(r.Context(), req.ProjectID, req.JobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, jr)
}

func (h *JobRunHandler) List(w http.ResponseWriter, r *http.Request) {
	jobRuns, err := h.jobRuns.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if jobRuns == nil {
		jobRuns = []domain.JobRun{}
	}

	writeJSON(w, http.StatusOK, jobRuns)
}

func (h *JobRunHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	jr, err := h.jobRuns.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrJobRunNotFound) {
			writeError(w, http.StatusNotFound, "job run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, jr)
}
