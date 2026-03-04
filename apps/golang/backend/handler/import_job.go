package handler

import (
	"net/http"

	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type ImportJobHandler struct {
	importJob *usecase.ImportJobService
}

func NewImportJobHandler(importJob *usecase.ImportJobService) *ImportJobHandler {
	return &ImportJobHandler{importJob: importJob}
}

func (h *ImportJobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateImportJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if req.ConnectionId == "" {
		writeError(w, http.StatusBadRequest, "connection_id is required")
		return
	}
	if req.SpreadsheetId == "" {
		writeError(w, http.StatusBadRequest, "spreadsheet_id is required")
		return
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	sheetName := ""
	if req.SheetName != nil {
		sheetName = *req.SheetName
	}
	rangeStr := ""
	if req.Range != nil {
		rangeStr = *req.Range
	}

	input := usecase.CreateImportJobInput{
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    desc,
		ConnectionID:   req.ConnectionId,
		SpreadsheetID:  req.SpreadsheetId,
		SheetName:      sheetName,
		Range:          rangeStr,
	}

	result, err := h.importJob.CreateImportJob(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, openapi.CreateImportJobResponse{
		Job:     toOpenAPIJob(result.Job),
		Version: toOpenAPIJobVersion(result.Version),
	})
}
