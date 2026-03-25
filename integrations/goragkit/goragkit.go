package goragkit

import (
	"context"
	"encoding/json"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// ValidateGoragkitResult validates a generic RAG result map against a pydantic model.
func ValidateGoragkitResult(ctx context.Context, m *pydantic.Model, result map[string]any, opts ...pydantic.ValidateOption) error {
	return pydantic.ValidateMap(ctx, m, result, opts...)
}

// ParseResult decodes and validates a JSON response to a typed output.
func ParseResult[T any](ctx context.Context, payload []byte, opts ...pydantic.ValidateOption) (T, error) {
	return pydantic.ParseAndValidate[T](ctx, payload, opts...)
}

// MarshalSchema marshals schema for prompt/tool definitions.
func MarshalSchema(m *pydantic.Model) ([]byte, error) {
	return json.MarshalIndent(m.Schema(), "", "  ")
}
