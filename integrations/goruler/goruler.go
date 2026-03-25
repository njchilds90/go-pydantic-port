package goruler

import (
	"context"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// RuleEngine defines minimal decision engine behavior.
type RuleEngine interface {
	Evaluate(ctx context.Context, input map[string]any) (bool, error)
}

// ValidateThenEvaluateWithRuler validates first, then evaluates rules.
func ValidateThenEvaluateWithRuler(ctx context.Context, m *pydantic.Model, input map[string]any, engine RuleEngine, opts ...pydantic.ValidateOption) (bool, error) {
	if err := pydantic.ValidateMap(ctx, m, input, opts...); err != nil {
		return false, err
	}
	return engine.Evaluate(ctx, input)
}
