package pydantic

import (
	"fmt"
	"reflect"
)

// Field defines a dynamic model field.
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Rules       []Rule `json:"rules,omitempty"`
}

// Model is a fluent validation model for map/object inputs.
type Model struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

// NewModel creates a new fluent model.
func NewModel(name string) *Model {
	return &Model{Name: name, Fields: make([]Field, 0)}
}

// Field adds a field to the model.
func (m *Model) Field(name, typ string, rules ...string) *Model {
	parsed := make([]Rule, 0, len(rules))
	for _, r := range rules {
		rs, _ := parseRules(r)
		parsed = append(parsed, rs...)
	}
	m.Fields = append(m.Fields, Field{Name: name, Type: typ, Rules: parsed})
	return m
}

// DescribeField sets description for an existing field.
func (m *Model) DescribeField(name, description string) *Model {
	for i := range m.Fields {
		if m.Fields[i].Name == name {
			m.Fields[i].Description = description
		}
	}
	return m
}

// Schema returns JSON Schema for the fluent model.
func (m *Model) Schema() map[string]any {
	props := map[string]any{}
	req := make([]string, 0)
	for _, f := range m.Fields {
		prop := map[string]any{"type": f.Type}
		if f.Description != "" {
			prop["description"] = f.Description
		}
		for _, r := range f.Rules {
			switch r.Name {
			case "min":
				if f.Type == "string" {
					prop["minLength"] = atoiDefault(r.Arg)
				} else {
					prop["minimum"] = atofDefault(r.Arg)
				}
			case "max":
				if f.Type == "string" {
					prop["maxLength"] = atoiDefault(r.Arg)
				} else {
					prop["maximum"] = atofDefault(r.Arg)
				}
			case "oneof":
				prop["enum"] = splitWords(r.Arg)
			case "regexp":
				prop["pattern"] = r.Arg
			case "required":
				req = append(req, f.Name)
			}
		}
		props[f.Name] = prop
	}
	s := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"title":                m.Name,
		"additionalProperties": false,
		"properties":           props,
	}
	if len(req) > 0 {
		s["required"] = req
	}
	return s
}

// JSONSchema generates JSON Schema from a typed struct.
func JSONSchema[T any]() (map[string]any, error) {
	var v T
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("JSONSchema requires struct type")
	}
	m := NewModel(t.Name())
	for i := range t.NumField() {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		ft := mapGoTypeToJSON(sf.Type)
		rules, _ := parseRules(sf.Tag.Get("validate"))
		m.Fields = append(m.Fields, Field{Name: sf.Name, Type: ft, Rules: rules, Description: sf.Tag.Get("description")})
	}
	return m.Schema(), nil
}

func mapGoTypeToJSON(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}
