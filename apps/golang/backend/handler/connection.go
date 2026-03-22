package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/connector"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type ConnectionHandler struct {
	connections *usecase.ConnectionService
	credentials *usecase.CredentialService
	registry    *connector.Registry
}

func NewConnectionHandler(connections *usecase.ConnectionService, credentials *usecase.CredentialService, registry *connector.Registry) *ConnectionHandler {
	return &ConnectionHandler{connections: connections, credentials: credentials, registry: registry}
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

	c, err := h.connections.Create(r.Context(), req.Name, req.Type, configJSON, req.SecretRef, req.CredentialId)
	if err != nil {
		if errors.Is(err, domain.ErrConnectorTypeUnknown) {
			writeError(w, http.StatusBadRequest, "unknown connector type")
			return
		}
		if errors.Is(err, domain.ErrConnectionConfigInvalid) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
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

	c, err := h.connections.Update(r.Context(), id, req.Name, req.Type, configJSON, req.SecretRef, req.CredentialId)
	if err != nil {
		if errors.Is(err, domain.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		if errors.Is(err, domain.ErrConnectorTypeUnknown) {
			writeError(w, http.StatusBadRequest, "unknown connector type")
			return
		}
		if errors.Is(err, domain.ErrConnectionConfigInvalid) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
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

func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
	var req openapi.TestConnectionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Type == "" || req.ConfigJson == "" {
		writeError(w, http.StatusBadRequest, "type and config_json are required")
		return
	}

	if !h.registry.Exists(req.Type) {
		writeError(w, http.StatusBadRequest, "unknown connector type")
		return
	}

	// Validate config against JSON Schema
	validationResult := openapi.ValidationResult{Status: openapi.ValidationResultStatusOk}
	if err := h.registry.ValidateConfig(req.Type, req.ConfigJson); err != nil {
		msg := err.Error()
		validationResult = openapi.ValidationResult{
			Status:  openapi.ValidationResultStatusFailed,
			Message: &msg,
		}
		// Return early: no point testing connectivity with invalid config
		skipMsg := "skipped due to validation failure"
		writeJSON(w, http.StatusOK, openapi.TestConnectionResponse{
			Validation: validationResult,
			Connectivity: openapi.ConnectivityResult{
				Status:  openapi.ConnectivityResultStatusSkipped,
				Message: &skipMsg,
			},
		})
		return
	}

	// If a real tester is registered, use it
	tester := h.registry.GetTester(req.Type)
	if tester != nil {
		accessToken := ""
		if req.CredentialId != nil && *req.CredentialId != "" && h.credentials != nil {
			tenantID, ok := domain.TenantIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "missing tenant")
				return
			}
			token, err := h.credentials.GetValidAccessToken(r.Context(), tenantID, *req.CredentialId)
			if err != nil {
				code := "unauthorized"
				msg := "failed to retrieve access token"
				writeJSON(w, http.StatusOK, openapi.TestConnectionResponse{
					Validation: validationResult,
					Connectivity: openapi.ConnectivityResult{
						Status:  openapi.ConnectivityResultStatusFailed,
						Code:    &code,
						Message: &msg,
					},
				})
				return
			}
			accessToken = token
		}

		result := tester.Test(r.Context(), req.ConfigJson, accessToken)
		connStatus := openapi.ConnectivityResultStatusOk
		if !result.OK {
			connStatus = openapi.ConnectivityResultStatusFailed
		}
		writeJSON(w, http.StatusOK, openapi.TestConnectionResponse{
			Validation: validationResult,
			Connectivity: openapi.ConnectivityResult{
				Status:  connStatus,
				Code:    &result.Code,
				Message: &result.Message,
			},
		})
		return
	}

	// No tester registered: report connectivity as skipped
	skipMsg := "no connectivity test available for this connector type"
	writeJSON(w, http.StatusOK, openapi.TestConnectionResponse{
		Validation: validationResult,
		Connectivity: openapi.ConnectivityResult{
			Status:  openapi.ConnectivityResultStatusSkipped,
			Message: &skipMsg,
		},
	})
}

func (h *ConnectionHandler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	connID := r.PathValue("connection_id")
	if connID == "" {
		writeError(w, http.StatusBadRequest, "missing connection_id")
		return
	}

	conn, err := h.connections.Get(r.Context(), connID)
	if err != nil {
		if errors.Is(err, domain.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	fetcher := h.registry.GetFetcher(conn.Type)
	if fetcher == nil {
		writeError(w, http.StatusBadRequest, "schema fetching not supported for this connector type")
		return
	}

	// Merge spreadsheet_id query param into configJSON if provided
	configJSON := conn.ConfigJSON
	if qsID := r.URL.Query().Get("spreadsheet_id"); qsID != "" {
		var cfgMap map[string]interface{}
		if err := json.Unmarshal([]byte(configJSON), &cfgMap); err != nil {
			cfgMap = map[string]interface{}{}
		}
		cfgMap["spreadsheet_id"] = qsID
		if merged, err := json.Marshal(cfgMap); err == nil {
			configJSON = string(merged)
		}
	}

	accessToken := ""
	if conn.CredentialID != nil && *conn.CredentialID != "" && h.credentials != nil {
		tenantID, ok := domain.TenantIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing tenant")
			return
		}
		token, err := h.credentials.GetValidAccessToken(r.Context(), tenantID, *conn.CredentialID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "credential_expired")
			return
		}
		accessToken = token
	}

	result, err := fetcher.FetchSchema(r.Context(), configJSON, accessToken)
	if err != nil {
		if err.Error() == "credential_expired" {
			writeError(w, http.StatusBadRequest, "credential_expired")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items := make([]openapi.SchemaItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = openapi.SchemaItem{
			Name: item.Name,
			Type: openapi.SchemaItemType(item.Type),
		}
		if len(item.Columns) > 0 {
			cols := make([]openapi.SchemaColumn, len(item.Columns))
			for j, c := range item.Columns {
				cols[j] = openapi.SchemaColumn{
					Name: c.Name,
					Type: c.Type,
				}
				if c.Nullable {
					cols[j].Nullable = &c.Nullable
				}
				if c.PrimaryKey {
					cols[j].PrimaryKey = &c.PrimaryKey
				}
				if c.CursorCandidate {
					cols[j].CursorCandidate = &c.CursorCandidate
				}
			}
			items[i].Columns = &cols
		}
		if len(item.PrimaryKey) > 0 {
			items[i].PrimaryKey = &item.PrimaryKey
		}
		if item.CursorField != "" {
			items[i].CursorField = &item.CursorField
		}
		if item.SupportsIncremental {
			items[i].SupportsIncremental = &item.SupportsIncremental
		}
		if len(item.Metadata) > 0 {
			m := map[string]interface{}(item.Metadata)
			items[i].Metadata = &m
		}
	}

	writeJSON(w, http.StatusOK, openapi.ConnectionSchemasResponse{
		Title: result.Title,
		Items: items,
	})
}
