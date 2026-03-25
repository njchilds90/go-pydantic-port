package goruler

import (
	"context"
	"testing"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

type e struct{}

func (e) Evaluate(_ context.Context, input map[string]any) (bool, error) {
	return input["ok"] == true, nil
}

func TestValidateThenEvaluateWithRuler(t *testing.T) {
	m := pydantic.NewModel("R").Field("ok", "boolean").Required().End()
	ok, err := ValidateThenEvaluateWithRuler(context.Background(), m, map[string]any{"ok": true}, e{})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
