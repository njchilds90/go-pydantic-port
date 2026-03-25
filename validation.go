// Package pydantic provides zero-dependency runtime data validation,
// parsing, and JSON Schema generation for Go applications and AI agents.
package pydantic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var modelCache sync.Map

// Validate validates a struct using `validate` tags.
//
// Supported rules include required, email, min, max, len, oneof, and regexp.
func Validate(ctx context.Context, model any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("validation canceled: %w", err)
	}
	if model == nil {
		return ErrInvalidModel
	}
	val := reflect.ValueOf(model)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ErrInvalidModel
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected struct, got %s", ErrInvalidModel, val.Kind())
	}

	meta, err := getOrBuildMetadata(val.Type())
	if err != nil {
		return err
	}
	verr := &ValidationError{Model: meta.Name}
	for _, f := range meta.Fields {
		if err := validateRules(f.Name, val.FieldByIndex(f.Index), f.Rules); err != nil {
			var fr RuleError
			if errors.As(err, &fr) {
				verr.Fields = append(verr.Fields, FieldError{Name: f.Name, Rule: fr.Rule, Message: fr.Error(), Value: fmt.Sprintf("%v", val.FieldByIndex(f.Index).Interface())})
				continue
			}
			verr.Fields = append(verr.Fields, FieldError{Name: f.Name, Message: err.Error()})
		}
	}
	if len(verr.Fields) > 0 {
		return verr
	}
	return nil
}

// ParseAndValidate unmarshals JSON into T and validates it.
func ParseAndValidate[T any](ctx context.Context, raw []byte) (T, error) {
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode input: %w", err)
	}
	if err := Validate(ctx, &out); err != nil {
		return out, err
	}
	return out, nil
}

// ValidateMap validates input data against a fluent Model definition.
func ValidateMap(ctx context.Context, m *Model, input map[string]any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("validation canceled: %w", err)
	}
	if m == nil {
		return ErrInvalidModel
	}
	verr := &ValidationError{Model: m.Name}
	for _, field := range m.Fields {
		v, ok := input[field.Name]
		if !ok {
			v = nil
		}
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			rv = reflect.ValueOf("")
		}
		if err := validateRules(field.Name, rv, field.Rules); err != nil {
			var re RuleError
			if errors.As(err, &re) {
				verr.Fields = append(verr.Fields, FieldError{Name: field.Name, Rule: re.Rule, Message: re.Error(), Value: fmt.Sprintf("%v", v)})
			} else {
				verr.Fields = append(verr.Fields, FieldError{Name: field.Name, Message: err.Error(), Value: fmt.Sprintf("%v", v)})
			}
		}
	}
	if len(verr.Fields) > 0 {
		return verr
	}
	return nil
}

type fieldMeta struct {
	Name  string
	Index []int
	Rules []Rule
}

type structMeta struct {
	Name   string
	Fields []fieldMeta
}

func getOrBuildMetadata(t reflect.Type) (structMeta, error) {
	if m, ok := modelCache.Load(t); ok {
		return m.(structMeta), nil
	}
	meta := structMeta{Name: t.Name()}
	for i := range t.NumField() {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		tag := sf.Tag.Get("validate")
		if strings.TrimSpace(tag) == "" {
			continue
		}
		rules, err := parseRules(tag)
		if err != nil {
			return structMeta{}, fmt.Errorf("parse rules for field %s: %w", sf.Name, err)
		}
		meta.Fields = append(meta.Fields, fieldMeta{Name: sf.Name, Index: sf.Index, Rules: rules})
	}
	modelCache.Store(t, meta)
	return meta, nil
}

// ClearModelCache removes cached reflected model metadata.
func ClearModelCache() {
	modelCache = sync.Map{}
}
