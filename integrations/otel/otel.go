// Package otel provides OpenTelemetry-compatible hooks without adding core dependencies.
package otel

import (
	"context"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// Tracer is the minimal interface needed for tracing integrations.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// Span is a minimal end-able tracing span.
type Span interface {
	End()
}

// Validate traces and validates a typed model.
func Validate(ctx context.Context, tracer Tracer, model any) error {
	ctx, span := tracer.Start(ctx, "pydantic.validate")
	defer span.End()
	return pydantic.Validate(ctx, model)
}

// ValidateMap traces and validates a dynamic model.
func ValidateMap(ctx context.Context, tracer Tracer, m *pydantic.Model, input map[string]any) error {
	ctx, span := tracer.Start(ctx, "pydantic.validate_map")
	defer span.End()
	return pydantic.ValidateMap(ctx, m, input)
}
