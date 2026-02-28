package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type JobRunHandler struct {
	jobRuns *usecase.JobRunService
}

func NewJobRunHandler(jobRuns *usecase.JobRunService) *JobRunHandler {
	return &JobRunHandler{jobRuns: jobRuns}
}

func (h *JobRunHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateJobRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.JobId == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	jr, err := h.jobRuns.Create(r.Context(), req.JobId, req.JobVersionId)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIJobRun(jr))
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

	items := make([]openapi.JobRun, len(jobRuns))
	for i := range jobRuns {
		items[i] = toOpenAPIJobRun(&jobRuns[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.JobRun `json:"items"`
	}{Items: items})
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

	writeJSON(w, http.StatusOK, toOpenAPIJobRun(jr))
}
