package pydantic

import (
	"context"
	"encoding/json"
	"testing"
)

type user struct {
	Name  string `validate:"required,min=2"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18,max=130"`
	Role  string `validate:"oneof=admin user"`
}

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		u    user
		ok   bool
	}{
		{name: "valid", u: user{Name: "Jane", Email: "jane@example.com", Age: 20, Role: "admin"}, ok: true},
		{name: "invalid email", u: user{Name: "Jane", Email: "bad", Age: 20, Role: "admin"}},
		{name: "missing required", u: user{Email: "jane@example.com", Age: 20, Role: "admin"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(context.Background(), tt.u)
			if tt.ok && err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected err")
			}
		})
	}
}

func TestParseAndValidate(t *testing.T) {
	t.Parallel()
	_, err := ParseAndValidate[user](context.Background(), []byte(`{"name":"Jay","email":"jay@example.com","age":33,"role":"user"}`))
	if err != nil {
		t.Fatalf("ParseAndValidate() err = %v", err)
	}
}

func TestModelBuilderAndSchema(t *testing.T) {
	t.Parallel()
	m := NewModel("Ticket").
		Field("title", "string", "required", "min=3").
		Field("priority", "string", "oneof=low medium high")

	payload := map[string]any{"title": "foo", "priority": "medium"}
	if err := ValidateMap(context.Background(), m, payload); err != nil {
		t.Fatalf("ValidateMap() err = %v", err)
	}
	if got := m.Schema()["type"]; got != "object" {
		t.Fatalf("schema type = %v", got)
	}

	b, err := json.Marshal(m.Schema())
	if err != nil || len(b) == 0 {
		t.Fatalf("schema marshal err = %v", err)
	}
}

func TestJSONSchema(t *testing.T) {
	t.Parallel()
	s, err := JSONSchema[user]()
	if err != nil {
		t.Fatalf("JSONSchema() err = %v", err)
	}
	if s["title"] != "user" {
		t.Fatalf("unexpected title: %v", s["title"])
	}
}

func BenchmarkValidate(b *testing.B) {
	u := user{Name: "Jane", Email: "jane@example.com", Age: 30, Role: "admin"}
	for i := 0; i < b.N; i++ {
		_ = Validate(context.Background(), u)
	}
}
