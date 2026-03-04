package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type DatasetHandler struct {
	datasets *usecase.DatasetService
}

func NewDatasetHandler(datasets *usecase.DatasetService) *DatasetHandler {
	return &DatasetHandler{datasets: datasets}
}

func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var filter domain.DatasetListFilter
	filter.Query = q.Get("q")
	filter.SourceType = q.Get("source_type")

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		filter.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		filter.Offset = n
	}

	datasets, err := h.datasets.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if datasets == nil {
		datasets = []domain.Dataset{}
	}

	items := make([]openapi.Dataset, len(datasets))
	for i := range datasets {
		items[i] = toOpenAPIDataset(&datasets[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Dataset `json:"items"`
	}{Items: items})
}

func (h *DatasetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	d, err := h.datasets.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIDataset(d))
}

func (h *DatasetHandler) GetRows(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 500 {
			writeError(w, http.StatusBadRequest, "invalid limit (1-500)")
			return
		}
		limit = n
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = n
	}

	result, err := h.datasets.GetRows(r.Context(), id, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		log.Printf("dataset get rows error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read dataset rows")
		return
	}

	columns := make([]openapi.DatasetColumn, len(result.Columns))
	for i, c := range result.Columns {
		columns[i] = openapi.DatasetColumn{Name: c.Name, Type: c.Type}
	}

	writeJSON(w, http.StatusOK, openapi.DatasetRowsResponse{
		Columns:   columns,
		Rows:      result.Rows,
		TotalRows: result.TotalRows,
		Limit:     limit,
		Offset:    offset,
	})
}

func (h *DatasetHandler) UpdateColumns(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdateDatasetColumnsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inputs := make([]usecase.UpdateColumnInput, len(req.Columns))
	for i, c := range req.Columns {
		inputs[i] = usecase.UpdateColumnInput{
			Name:        c.Name,
			Description: ptrStr(c.Description),
		}
		if c.SemanticType != nil {
			inputs[i].SemanticType = string(*c.SemanticType)
		}
		if c.Tags != nil {
			inputs[i].Tags = *c.Tags
		}
	}

	d, err := h.datasets.UpdateColumns(r.Context(), id, inputs)
	if err != nil {
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		if errors.Is(err, domain.ErrColumnNotFound) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("update columns error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIDataset(d))
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
