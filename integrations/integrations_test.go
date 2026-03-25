package integrations_test

import (
	"context"
	"testing"

	pydantic "github.com/njchilds90/go-pydantic-port"
	"github.com/njchilds90/go-pydantic-port/integrations/goragkit"
	"github.com/njchilds90/go-pydantic-port/integrations/goruler"
)

type engine struct{}

func (engine) Evaluate(_ context.Context, input map[string]any) (bool, error) {
	return input["status"] == "approved", nil
}

func TestIntegrations(t *testing.T) {
	m := pydantic.NewModel("Decision").Field("status", "string", "required", "oneof=approved denied")
	if err := goragkit.ValidateResult(context.Background(), m, map[string]any{"status": "approved"}); err != nil {
		t.Fatalf("ValidateResult err=%v", err)
	}
	ok, err := goruler.ValidateThenEvaluate(context.Background(), m, map[string]any{"status": "approved"}, engine{})
	if err != nil || !ok {
		t.Fatalf("ValidateThenEvaluate ok=%v err=%v", ok, err)
	}
}
