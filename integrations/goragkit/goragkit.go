// Package goragkit provides helpers for validating LLM/RAG structured responses.
package goragkit

import (
	"context"
	"encoding/json"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// ValidateResult validates a generic RAG result map against a pydantic model.
func ValidateResult(ctx context.Context, m *pydantic.Model, result map[string]any) error {
	return pydantic.ValidateMap(ctx, m, result)
}

// ParseResult decodes and validates a JSON response to a typed output.
func ParseResult[T any](ctx context.Context, payload []byte) (T, error) {
	return pydantic.ParseAndValidate[T](ctx, payload)
}

// MarshalSchema marshals schema for prompt/tool definitions.
func MarshalSchema(m *pydantic.Model) ([]byte, error) {
	return json.MarshalIndent(m.Schema(), "", "  ")
}
