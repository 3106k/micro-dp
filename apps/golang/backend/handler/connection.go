package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type ConnectionHandler struct {
	connections *usecase.ConnectionService
}

func NewConnectionHandler(connections *usecase.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{connections: connections}
}

func (h *ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateConnectionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}

	configJSON := "{}"
	if req.ConfigJson != nil {
		configJSON = *req.ConfigJson
	}

	c, err := h.connections.Create(r.Context(), req.Name, req.Type, configJSON, req.SecretRef)
	if err != nil {
		if errors.Is(err, domain.ErrConnectionNameDuplicate) {
			writeError(w, http.StatusConflict, "connection name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIConnection(c))
}

func (h *ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
	connections, err := h.connections.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if connections == nil {
		connections = []domain.Connection{}
	}

	items := make([]openapi.Connection, len(connections))
	for i := range connections {
		items[i] = toOpenAPIConnection(&connections[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Connection `json:"items"`
	}{Items: items})
}

func (h *ConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	c, err := h.connections.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIConnection(c))
}

func (h *ConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req openapi.UpdateConnectionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}

	configJSON := "{}"
	if req.ConfigJson != nil {
		configJSON = *req.ConfigJson
	}

	c, err := h.connections.Update(r.Context(), id, req.Name, req.Type, configJSON, req.SecretRef)
	if err != nil {
		if errors.Is(err, domain.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		if errors.Is(err, domain.ErrConnectionNameDuplicate) {
			writeError(w, http.StatusConflict, "connection name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIConnection(c))
}

func (h *ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.connections.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
