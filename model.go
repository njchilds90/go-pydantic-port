package pydantic

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Field defines a model field including nested structures.
type Field struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Description  string   `json:"description,omitempty"`
	Default      any      `json:"default,omitempty"`
	Rules        []Rule   `json:"rules,omitempty"`
	Items        *Field   `json:"items,omitempty"`
	Properties   []Field  `json:"properties,omitempty"`
	MapValue     *Field   `json:"map_value,omitempty"`
	RequiredFlag bool     `json:"required,omitempty"`
	CoerceFlag   bool     `json:"coerce,omitempty"`
	Custom       []string `json:"custom,omitempty"`
	ModelRef     *Model   `json:"-"`
}

// Model is a fluent validation model for map/object inputs.
type Model struct {
	Name       string                   `json:"name"`
	Fields     []Field                  `json:"fields"`
	StrictMode bool                     `json:"strict_mode,omitempty"`
	validators map[string]ValidatorFunc `json:"-"`
}

// CompiledModel is a precompiled model optimized for repeated validations.
type CompiledModel struct {
	Model *Model
}

// FieldBuilder provides chainable field configuration.
type FieldBuilder struct {
	model *Model
	idx   int
}

// NewModel creates a new fluent model.
func NewModel(name string) *Model {
	return &Model{Name: name, Fields: make([]Field, 0), StrictMode: true, validators: map[string]ValidatorFunc{}}
}

// AddValidator registers a model-scoped custom validator.
func (m *Model) AddValidator(name string, fn ValidatorFunc) *Model {
	if m.validators == nil {
		m.validators = map[string]ValidatorFunc{}
	}
	m.validators[name] = fn
	return m
}

// SetStrictMode toggles strict mode for this model.
func (m *Model) SetStrictMode(strict bool) *Model {
	m.StrictMode = strict
	return m
}

// Field adds a field to the model and returns a builder for chainable configuration.
//
// spec may be a JSON type string ("string", "object", ...), or a nested *Model.
func (m *Model) Field(name string, spec any, rules ...string) *FieldBuilder {
	f := Field{Name: name, Type: "string"}
	switch s := spec.(type) {
	case string:
		f.Type = s
	case *Model:
		f.Type = "object"
		f.ModelRef = s
		f.Properties = s.Fields
	default:
		f.Type = "string"
	}
	for _, r := range rules {
		rs, _ := parseRules(r)
		for _, parsed := range rs {
			switch parsed.Name {
			case "required":
				f.RequiredFlag = true
				f.Rules = append(f.Rules, parsed)
			case "email", "min", "max", "len", "oneof", "regexp":
				f.Rules = append(f.Rules, parsed)
			default:
				f.Custom = append(f.Custom, parsed.Name)
			}
		}
	}
	m.Fields = append(m.Fields, f)
	return &FieldBuilder{model: m, idx: len(m.Fields) - 1}
}

// Compile precompiles the model and nested objects for reuse.
func (m *Model) Compile() *CompiledModel { return &CompiledModel{Model: m} }

// MustSchema returns model schema or panics on impossible serialization issues.
func (m *Model) MustSchema() map[string]any {
	s := m.Schema()
	if _, err := json.Marshal(s); err != nil {
		panic(err)
	}
	return s
}

// Schema returns JSON Schema for the fluent model.
func (m *Model) Schema() map[string]any {
	defs := map[string]any{}
	s := schemaWithDefs(m, defs)
	if len(defs) > 0 {
		s["$defs"] = defs
	}
	return s
}

func schemaWithDefs(m *Model, defs map[string]any) map[string]any {
	props := map[string]any{}
	req := make([]string, 0)
	for _, f := range m.Fields {
		prop := fieldToSchema(f, defs)
		if f.RequiredFlag {
			req = append(req, f.Name)
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

func fieldToSchema(f Field, defs map[string]any) map[string]any {
	prop := map[string]any{"type": f.Type}
	if f.Description != "" {
		prop["description"] = f.Description
	}
	if f.Default != nil {
		prop["default"] = f.Default
	}
	for _, r := range f.Rules {
		switch r.Name {
		case "min":
			if f.Type == "string" || f.Type == "array" {
				prop["minLength"] = atoiDefault(r.Arg)
			} else {
				prop["minimum"] = atofDefault(r.Arg)
			}
		case "max":
			if f.Type == "string" || f.Type == "array" {
				prop["maxLength"] = atoiDefault(r.Arg)
			} else {
				prop["maximum"] = atofDefault(r.Arg)
			}
		case "oneof":
			prop["enum"] = splitWords(r.Arg)
		case "regexp":
			prop["pattern"] = r.Arg
		}
	}
	if f.ModelRef != nil {
		defs[f.ModelRef.Name] = schemaWithDefs(f.ModelRef, defs)
		return map[string]any{"$ref": "#/$defs/" + f.ModelRef.Name}
	}
	if f.Items != nil {
		prop["items"] = fieldToSchema(*f.Items, defs)
	}
	if len(f.Properties) > 0 {
		nested := &Model{Name: f.Name, Fields: f.Properties}
		ns := schemaWithDefs(nested, defs)
		prop["type"] = "object"
		prop["properties"] = ns["properties"]
		if req, ok := ns["required"]; ok {
			prop["required"] = req
		}
	}
	if f.MapValue != nil {
		prop["type"] = "object"
		prop["additionalProperties"] = fieldToSchema(*f.MapValue, defs)
	}
	return prop
}

// Required marks this field as required.
func (fb *FieldBuilder) Required() *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.RequiredFlag = true
	f.Rules = append(f.Rules, Rule{Name: "required"})
	return fb
}

// Coerce enables value coercion for this field when model strict mode is disabled.
func (fb *FieldBuilder) Coerce() *FieldBuilder {
	fb.model.Fields[fb.idx].CoerceFlag = true
	return fb
}

// Min adds a minimum numeric or length constraint.
func (fb *FieldBuilder) Min(v int) *FieldBuilder {
	fb.model.Fields[fb.idx].Rules = append(fb.model.Fields[fb.idx].Rules, Rule{Name: "min", Arg: fmt.Sprintf("%d", v)})
	return fb
}

// Max adds a maximum numeric or length constraint.
func (fb *FieldBuilder) Max(v int) *FieldBuilder {
	fb.model.Fields[fb.idx].Rules = append(fb.model.Fields[fb.idx].Rules, Rule{Name: "max", Arg: fmt.Sprintf("%d", v)})
	return fb
}

// Pattern adds a regex pattern constraint.
func (fb *FieldBuilder) Pattern(expr string) *FieldBuilder {
	fb.model.Fields[fb.idx].Rules = append(fb.model.Fields[fb.idx].Rules, Rule{Name: "regexp", Arg: expr})
	return fb
}

// OneOf adds enum values.
func (fb *FieldBuilder) OneOf(values ...string) *FieldBuilder {
	fb.model.Fields[fb.idx].Rules = append(fb.model.Fields[fb.idx].Rules, Rule{Name: "oneof", Arg: joinWords(values)})
	return fb
}

// Default sets default value.
func (fb *FieldBuilder) Default(v any) *FieldBuilder {
	fb.model.Fields[fb.idx].Default = v
	return fb
}

// Description sets field description.
func (fb *FieldBuilder) Description(v string) *FieldBuilder {
	fb.model.Fields[fb.idx].Description = v
	return fb
}

// ArrayOf configures array item type.
func (fb *FieldBuilder) ArrayOf(itemType string) *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.Type = "array"
	f.Items = &Field{Name: f.Name + "_item", Type: itemType}
	return fb
}

// ArrayOfModel configures array item nested model.
func (fb *FieldBuilder) ArrayOfModel(m *Model) *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.Type = "array"
	f.Items = &Field{Name: f.Name + "_item", Type: "object", ModelRef: m, Properties: m.Fields}
	return fb
}

// Object configures nested object properties.
func (fb *FieldBuilder) Object(fields ...Field) *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.Type = "object"
	f.Properties = fields
	return fb
}

// MapValues configures map value type.
func (fb *FieldBuilder) MapValues(valueType string) *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.Type = "object"
	f.MapValue = &Field{Name: f.Name + "_value", Type: valueType}
	return fb
}

// Custom registers custom validator names for field.
func (fb *FieldBuilder) Custom(names ...string) *FieldBuilder {
	f := &fb.model.Fields[fb.idx]
	f.Custom = append(f.Custom, names...)
	return fb
}

// End returns the model for further chaining.
func (fb *FieldBuilder) End() *Model { return fb.model }

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
	model := modelFromType(t, make(map[reflect.Type]bool), 0, 10)
	return model.Schema(), nil
}
