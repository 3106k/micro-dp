package handler

import (
	"errors"
	"net/http"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type ModuleTypeHandler struct {
	moduleTypes *usecase.ModuleTypeService
}

func NewModuleTypeHandler(moduleTypes *usecase.ModuleTypeService) *ModuleTypeHandler {
	return &ModuleTypeHandler{moduleTypes: moduleTypes}
}

func (h *ModuleTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateModuleTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	mt, err := h.moduleTypes.Create(r.Context(), req.Name, string(req.Category))
	if err != nil {
		if errors.Is(err, domain.ErrModuleTypeNameDuplicate) {
			writeError(w, http.StatusConflict, "module type name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIModuleType(mt))
}

func (h *ModuleTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	types, err := h.moduleTypes.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if types == nil {
		types = []domain.ModuleType{}
	}

	items := make([]openapi.ModuleType, len(types))
	for i := range types {
		items[i] = toOpenAPIModuleType(&types[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.ModuleType `json:"items"`
	}{Items: items})
}

func (h *ModuleTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	mt, err := h.moduleTypes.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrModuleTypeNotFound) {
			writeError(w, http.StatusNotFound, "module type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIModuleType(mt))
}

func (h *ModuleTypeHandler) CreateSchema(w http.ResponseWriter, r *http.Request) {
	moduleTypeID := r.PathValue("id")
	if moduleTypeID == "" {
		writeError(w, http.StatusBadRequest, "missing module type id")
		return
	}

	var req openapi.CreateModuleTypeSchemaRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.JsonSchema == "" {
		writeError(w, http.StatusBadRequest, "json_schema is required")
		return
	}

	schema, err := h.moduleTypes.CreateSchema(r.Context(), moduleTypeID, req.JsonSchema)
	if err != nil {
		if errors.Is(err, domain.ErrModuleTypeNotFound) {
			writeError(w, http.StatusNotFound, "module type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, toOpenAPIModuleTypeSchema(schema))
}

func (h *ModuleTypeHandler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	moduleTypeID := r.PathValue("id")
	if moduleTypeID == "" {
		writeError(w, http.StatusBadRequest, "missing module type id")
		return
	}

	schemas, err := h.moduleTypes.ListSchemas(r.Context(), moduleTypeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if schemas == nil {
		schemas = []domain.ModuleTypeSchema{}
	}

	items := make([]openapi.ModuleTypeSchema, len(schemas))
	for i := range schemas {
		items[i] = toOpenAPIModuleTypeSchema(&schemas[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.ModuleTypeSchema `json:"items"`
	}{Items: items})
}
