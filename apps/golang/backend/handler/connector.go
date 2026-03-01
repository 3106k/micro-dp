package handler

import (
	"net/http"

	"github.com/user/micro-dp/internal/connector"
	"github.com/user/micro-dp/internal/openapi"
)

type ConnectorHandler struct {
	registry *connector.Registry
}

func NewConnectorHandler(registry *connector.Registry) *ConnectorHandler {
	return &ConnectorHandler{registry: registry}
}

func (h *ConnectorHandler) List(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")

	defs := h.registry.List(kind)
	items := make([]openapi.ConnectorDefinition, len(defs))
	for i, d := range defs {
		items[i] = toOpenAPIConnectorDefinition(d)
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.ConnectorDefinition `json:"items"`
	}{Items: items})
}

func (h *ConnectorHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	def := h.registry.Get(id)
	if def == nil {
		writeError(w, http.StatusNotFound, "connector not found")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIConnectorDefinitionDetail(def))
}
