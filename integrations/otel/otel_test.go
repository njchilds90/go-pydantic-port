package otel

import (
	"context"
	"testing"
)

type span struct{}

func (span) End() {}

type tracer struct{}

func (tracer) Start(ctx context.Context, _ string) (context.Context, Span) { return ctx, span{} }

type payload struct {
	Name string `validate:"required"`
}

func TestValidate(t *testing.T) {
	if err := Validate(context.Background(), tracer{}, payload{Name: "ok"}); err != nil {
		t.Fatal(err)
	}
}
