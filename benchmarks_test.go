package pydantic

import (
	"context"
	"testing"
)

func BenchmarkValidateSimpleMap(b *testing.B) {
	m := NewModel("Simple").Field("id", "integer", "required").End().Compile()
	payload := map[string]any{"id": 1}
	for i := 0; i < b.N; i++ {
		_ = m.Validate(context.Background(), payload)
	}
}

func BenchmarkValidateNestedMap(b *testing.B) {
	addr := NewModel("Address").Field("city", "string", "required").End()
	m := NewModel("User").Field("address", addr, "required").End().Compile()
	payload := map[string]any{"address": map[string]any{"city": "NY"}}
	for i := 0; i < b.N; i++ {
		_ = m.Validate(context.Background(), payload)
	}
}

func BenchmarkValidateCoercionPath(b *testing.B) {
	m := NewModel("Payload").SetStrictMode(false).Field("age", "integer", "required").Coerce().End().Compile()
	payload := map[string]any{"age": "21"}
	for i := 0; i < b.N; i++ {
		_ = m.Validate(context.Background(), payload, WithStrict(false))
	}
}
