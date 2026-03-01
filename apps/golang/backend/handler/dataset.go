package handler

import (
	"errors"
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
