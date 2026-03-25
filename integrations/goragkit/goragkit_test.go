package goragkit

import (
	"context"
	"testing"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

func TestValidateGoragkitResult(t *testing.T) {
	m := pydantic.NewModel("R").Field("status", "string").Required().End()
	if err := ValidateGoragkitResult(context.Background(), m, map[string]any{"status": "ok"}); err != nil {
		t.Fatal(err)
	}
}
