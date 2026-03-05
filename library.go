
// Package pydantic provides a simple data validation system.
package pydantic

import (
    "context"
    "errors"
    "github.com/go-playground/validator/v10"
)

// ErrInvalidModel is returned when the model is invalid.
var ErrInvalidModel = errors.New("model is invalid")

// ErrEmptyModelName is returned when the model name is empty.
var ErrEmptyModelName = errors.New("model name is empty")

// Validate validates a data model against a set of validation rules.
// Validate returns ErrInvalidModel if the model is invalid.
func Validate(ctx context.Context, model interface{}) error {
    if model == nil {
        return errors.New("model is nil")
    }
    validate := validator.New()
    err := validate.Struct(model)
    if err != nil {
        return fmt.Errorf("validate model: %w", err)
    }
    return nil
}

// ModelOption represents an option for the Model function.
type ModelOption func(*modelOptions)

// modelOptions holds the configuration for the Model function.
type modelOptions struct {
    name   string
    fields map[string]interface{}
}

// WithName sets the name of the model.
func WithName(name string) ModelOption {
    return func(o *modelOptions) {
        o.name = name
    }
}

// WithFields sets the fields of the model.
func WithFields(fields map[string]interface{}) ModelOption {
    return func(o *modelOptions) {
        o.fields = fields
    }
}

// Model defines a new data model.
// Model returns ErrEmptyModelName if the model name is empty.
func Model(opts ...ModelOption) (interface{}, error) {
    options := &modelOptions{}
    for _, opt := range opts {
        opt(options)
    }
    if options.name == "" {
        return nil, ErrEmptyModelName
    }
    return options.fields, nil
}
