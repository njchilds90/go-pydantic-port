package pydantic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type profile struct {
	Bio string `json:"bio" validate:"required,min=3"`
}
type user struct {
	Name    string  `json:"name" validate:"required,min=2"`
	Profile profile `json:"profile" validate:"required"`
}

func TestNestedModelReference(t *testing.T) {
	address := NewModel("Address").Field("city", "string", "required").End()
	m := NewModel("User").Field("address", address, "required").End()
	if err := ValidateMap(context.Background(), m, map[string]any{"address": map[string]any{"city": "NY"}}); err != nil {
		t.Fatal(err)
	}
	err := ValidateMap(context.Background(), m, map[string]any{"address": map[string]any{}})
	if err == nil {
		t.Fatal("expected nested error")
	}
}

func TestCustomValidatorTable(t *testing.T) {
	m := NewModel("Email").
		AddValidator("is_email", func(_ context.Context, v any) error {
			s := fmt.Sprintf("%v", v)
			if len(s) < 3 || !strings.Contains(s, "@") {
				return fmt.Errorf("invalid email")
			}
			return nil
		}).
		Field("email", "string", "required").Custom("is_email").End()

	tests := []struct {
		in any
		ok bool
	}{{"a@x.com", true}, {"xx", false}}
	for _, tt := range tests {
		err := ValidateMap(context.Background(), m, map[string]any{"email": tt.in})
		if (err == nil) != tt.ok {
			t.Fatalf("email=%v err=%v", tt.in, err)
		}
	}
}

func TestCoercionStrictMode(t *testing.T) {
	m := NewModel("Payload").SetStrictMode(false).Field("age", "integer", "min=18").Coerce().End()
	if err := ValidateMap(context.Background(), m, map[string]any{"age": "22"}); err != nil {
		t.Fatal(err)
	}
	strict := NewModel("Payload").Field("age", "integer", "min=18").End()
	if err := ValidateMap(context.Background(), strict, map[string]any{"age": "22"}); err == nil {
		t.Fatal("expected strict mode failure")
	}
}

func TestStructNestedValidationPath(t *testing.T) {
	err := Validate(context.Background(), user{Name: "ok", Profile: profile{Bio: "no"}})
	if err == nil {
		t.Fatal("expected error")
	}
	var verr *ValidationError
	if !errors.As(err, &verr) || verr.Fields[0].Path == "" {
		t.Fatalf("invalid error: %v", err)
	}
}

func TestSchemaDefs(t *testing.T) {
	child := NewModel("Child").Field("value", "string", "required").End()
	m := NewModel("Parent").Field("child", child, "required").End()
	s := m.Schema()
	if _, ok := s["$defs"]; !ok {
		t.Fatal("expected $defs")
	}
}
