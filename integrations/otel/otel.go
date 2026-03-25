package otel

import (
	"context"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// Span is an OTEL-compatible span interface.
type Span interface{ End() }

// Tracer is an OTEL-compatible tracer interface.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

type adapter struct{ tracer Tracer }

func (a adapter) Start(ctx context.Context, operation string) (context.Context, pydantic.Span) {
	ctx, span := a.tracer.Start(ctx, operation)
	return ctx, span
}

// Validate validates with tracing enabled.
func Validate(ctx context.Context, tracer Tracer, model any, opts ...pydantic.ValidateOption) error {
	opts = append(opts, pydantic.WithTracer(adapter{tracer: tracer}))
	return pydantic.Validate(ctx, model, opts...)
}

// ParseAndValidate parses and validates with tracing enabled.
func ParseAndValidate[T any](ctx context.Context, tracer Tracer, payload []byte, opts ...pydantic.ValidateOption) (T, error) {
	opts = append(opts, pydantic.WithTracer(adapter{tracer: tracer}))
	return pydantic.ParseAndValidate[T](ctx, payload, opts...)
}
