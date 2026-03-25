// Package goruler connects pydantic validation with rule-engine decisions.
package goruler

import (
	"context"

	pydantic "github.com/njchilds90/go-pydantic-port"
)

// RuleEngine defines minimal decision engine behavior.
type RuleEngine interface {
	Evaluate(ctx context.Context, input map[string]any) (bool, error)
}

// ValidateThenEvaluate validates first, then evaluates rules.
func ValidateThenEvaluate(ctx context.Context, m *pydantic.Model, input map[string]any, engine RuleEngine) (bool, error) {
	if err := pydantic.ValidateMap(ctx, m, input); err != nil {
		return false, err
	}
	return engine.Evaluate(ctx, input)
}
