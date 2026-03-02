package handler

import (
	"net/http"

	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type TransformHandler struct {
	transform *usecase.TransformService
}

func NewTransformHandler(transform *usecase.TransformService) *TransformHandler {
	return &TransformHandler{transform: transform}
}

func (h *TransformHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req openapi.TransformValidateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Sql == "" {
		writeError(w, http.StatusBadRequest, "sql is required")
		return
	}
	if len(req.DatasetIds) == 0 {
		writeError(w, http.StatusBadRequest, "dataset_ids is required")
		return
	}

	result, err := h.transform.ValidateSQL(r.Context(), req.Sql, req.DatasetIds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := openapi.TransformValidateResponse{
		Valid: result.Valid,
	}
	if result.Error != "" {
		resp.Error = &result.Error
	}
	if len(result.Columns) > 0 {
		cols := make([]openapi.DatasetColumn, len(result.Columns))
		for i, c := range result.Columns {
			cols[i] = openapi.DatasetColumn{Name: c.Name, Type: c.Type}
		}
		resp.Columns = &cols
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TransformHandler) Preview(w http.ResponseWriter, r *http.Request) {
	var req openapi.TransformPreviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Sql == "" {
		writeError(w, http.StatusBadRequest, "sql is required")
		return
	}
	if len(req.DatasetIds) == 0 {
		writeError(w, http.StatusBadRequest, "dataset_ids is required")
		return
	}

	limit := 100
	if req.Limit != nil {
		limit = *req.Limit
	}

	result, err := h.transform.PreviewSQL(r.Context(), req.Sql, req.DatasetIds, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cols := make([]openapi.DatasetColumn, len(result.Columns))
	for i, c := range result.Columns {
		cols[i] = openapi.DatasetColumn{Name: c.Name, Type: c.Type}
	}

	rows := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		rows[i] = row
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, openapi.TransformPreviewResponse{
		Columns:  cols,
		Rows:     rows,
		RowCount: result.RowCount,
	})
}

func (h *TransformHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateTransformJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if req.Sql == "" {
		writeError(w, http.StatusBadRequest, "sql is required")
		return
	}
	if len(req.DatasetIds) == 0 {
		writeError(w, http.StatusBadRequest, "dataset_ids is required")
		return
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	execution := "save_only"
	if req.Execution != nil {
		execution = string(*req.Execution)
	}

	input := usecase.CreateTransformJobInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: desc,
		SQL:         req.Sql,
		DatasetIDs:  req.DatasetIds,
		Execution:   execution,
		ScheduledAt: req.ScheduledAt,
	}

	result, err := h.transform.CreateTransformJob(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := openapi.CreateTransformJobResponse{
		Job:     toOpenAPIJob(result.Job),
		Version: toOpenAPIJobVersion(result.Version),
	}
	if result.JobRun != nil {
		jr := toOpenAPIJobRun(result.JobRun)
		resp.JobRun = &jr
	}

	writeJSON(w, http.StatusCreated, resp)
}
