package pydantic

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidModel indicates model input does not match expected format.
var ErrInvalidModel = errors.New("invalid model")

// ValidationError contains all field-level validation failures.
type ValidationError struct {
	Model  string       `json:"model"`
	Fields []FieldError `json:"fields"`
}

// Error implements error.
func (e *ValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s(%s)", f.Name, f.Rule))
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Model, strings.Join(parts, ", "))
}

// FieldError contains one field-level validation failure.
type FieldError struct {
	Name    string `json:"name"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}
