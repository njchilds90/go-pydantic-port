package pydantic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidModel indicates model input does not match expected format.
var ErrInvalidModel = errors.New("invalid model")

// ValidationError contains all field-level validation failures.
type ValidationError struct {
	Model  string       `json:"model"`
	Locale string       `json:"locale,omitempty"`
	Fields []FieldError `json:"fields"`
}

// Error implements error.
func (e *ValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s(%s)", f.Path, f.Rule))
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Model, strings.Join(parts, ", "))
}

// JSON returns machine-friendly JSON representation.
func (e *ValidationError) JSON() []byte {
	out, _ := json.Marshal(e)
	return out
}

// FieldError contains one field-level validation failure.
type FieldError struct {
	Path    string `json:"path"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}
