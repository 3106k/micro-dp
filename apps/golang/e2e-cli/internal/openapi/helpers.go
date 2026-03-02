package openapi

import openapi_types "github.com/oapi-codegen/runtime/types"

// ListResponse wraps list endpoints that return {"items": [...]}.
type ListResponse[T any] struct {
	Items []T `json:"items"`
}

// Ptr returns a pointer to the given value (for optional fields).
func Ptr[T any](v T) *T { return &v }

// Email converts string to openapi_types.Email.
func Email(s string) openapi_types.Email {
	return openapi_types.Email(s)
}
