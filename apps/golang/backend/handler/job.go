package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type JobHandler struct {
	jobs *usecase.JobService
}

func NewJobHandler(jobs *usecase.JobService) *JobHandler {
	return &JobHandler{jobs: jobs}
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	job, err := h.jobs.CreateJob(r.Context(), req.Name, req.Slug, desc)
	if err != nil {
		if errors.Is(err, domain.ErrJobSlugDuplicate) {
			writeError(w, http.StatusConflict, "job slug already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIJob(job))
}

func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobs.ListJobs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}

	items := make([]openapi.Job, len(jobs))
	for i := range jobs {
		items[i] = toOpenAPIJob(&jobs[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Job `json:"items"`
	}{Items: items})
}

func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	job, err := h.jobs.GetJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIJob(job))
}

func (h *JobHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdateJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	job, err := h.jobs.UpdateJob(r.Context(), id, req.Name, req.Slug, desc, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		if errors.Is(err, domain.ErrJobSlugDuplicate) {
			writeError(w, http.StatusConflict, "job slug already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIJob(job))
}

func (h *JobHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "missing job_id")
		return
	}

	var req openapi.CreateJobVersionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Modules) == 0 {
		writeError(w, http.StatusBadRequest, "at least one module is required")
		return
	}

	modules := make([]usecase.CreateModuleInput, len(req.Modules))
	for i, m := range req.Modules {
		configJSON := "{}"
		if m.ConfigJson != nil {
			configJSON = *m.ConfigJson
		}
		var posX, posY float64
		if m.PositionX != nil {
			posX = float64(*m.PositionX)
		}
		if m.PositionY != nil {
			posY = float64(*m.PositionY)
		}
		modules[i] = usecase.CreateModuleInput{
			ModuleTypeID:       m.ModuleTypeId,
			ModuleTypeSchemaID: m.ModuleTypeSchemaId,
			ConnectionID:       m.ConnectionId,
			Name:               m.Name,
			ConfigJSON:         configJSON,
			PositionX:          posX,
			PositionY:          posY,
		}
	}

	var edges []usecase.CreateEdgeInput
	if req.Edges != nil {
		edges = make([]usecase.CreateEdgeInput, len(*req.Edges))
		for i, e := range *req.Edges {
			edges[i] = usecase.CreateEdgeInput{
				SourceModuleIndex: e.SourceModuleIndex,
				TargetModuleIndex: e.TargetModuleIndex,
			}
		}
	}

	version, err := h.jobs.CreateVersion(r.Context(), jobID, modules, edges)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIJobVersion(version))
}

func (h *JobHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "missing job_id")
		return
	}

	versions, err := h.jobs.ListVersions(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if versions == nil {
		versions = []domain.JobVersion{}
	}

	items := make([]openapi.JobVersion, len(versions))
	for i := range versions {
		items[i] = toOpenAPIJobVersion(&versions[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.JobVersion `json:"items"`
	}{Items: items})
}

func (h *JobHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	versionID := r.PathValue("version_id")
	if jobID == "" || versionID == "" {
		writeError(w, http.StatusBadRequest, "missing job_id or version_id")
		return
	}

	v, mods, edges, err := h.jobs.GetVersionDetail(r.Context(), jobID, versionID)
	if err != nil {
		if errors.Is(err, domain.ErrJobVersionNotFound) {
			writeError(w, http.StatusNotFound, "job version not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if mods == nil {
		mods = []domain.JobModule{}
	}
	if edges == nil {
		edges = []domain.JobModuleEdge{}
	}

	apiMods := make([]openapi.JobModule, len(mods))
	for i := range mods {
		apiMods[i] = toOpenAPIJobModule(&mods[i])
	}
	apiEdges := make([]openapi.JobModuleEdge, len(edges))
	for i := range edges {
		apiEdges[i] = toOpenAPIJobModuleEdge(&edges[i])
	}

	writeJSON(w, http.StatusOK, openapi.JobVersionDetail{
		Version: toOpenAPIJobVersion(v),
		Modules: apiMods,
		Edges:   apiEdges,
	})
}

func (h *JobHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	versionID := r.PathValue("version_id")
	if jobID == "" || versionID == "" {
		writeError(w, http.StatusBadRequest, "missing job_id or version_id")
		return
	}

	v, err := h.jobs.PublishVersion(r.Context(), jobID, versionID)
	if err != nil {
		if errors.Is(err, domain.ErrJobVersionNotFound) {
			writeError(w, http.StatusNotFound, "job version not found")
			return
		}
		if errors.Is(err, domain.ErrJobVersionImmutable) {
			writeError(w, http.StatusConflict, "version already published")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIJobVersion(v))
}
