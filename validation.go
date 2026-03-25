package pydantic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var modelCache sync.Map

// ValidatorFunc defines a pluggable custom validator.
type ValidatorFunc func(ctx context.Context, value any) error

var (
	validatorMu       sync.RWMutex
	customValidators  = map[string]ValidatorFunc{}
	defaultLocalizer  = NewLocalizer("en")
	defaultMaxDepth   = 10
	defaultStrictMode = true
)

// ValidationOptions controls validation behavior.
type ValidationOptions struct {
	Strict    bool
	MaxDepth  int
	Locale    string
	Localizer *Localizer
	Tracer    Tracer
}

type ValidateOption func(*ValidationOptions)

func WithStrict(strict bool) ValidateOption { return func(o *ValidationOptions) { o.Strict = strict } }
func WithMaxDepth(depth int) ValidateOption { return func(o *ValidationOptions) { o.MaxDepth = depth } }
func WithLocale(locale string) ValidateOption {
	return func(o *ValidationOptions) { o.Locale = locale }
}
func WithLocalizer(l *Localizer) ValidateOption {
	return func(o *ValidationOptions) { o.Localizer = l }
}
func WithTracer(t Tracer) ValidateOption { return func(o *ValidationOptions) { o.Tracer = t } }

type Tracer interface {
	Start(ctx context.Context, operation string) (context.Context, Span)
}
type Span interface{ End() }

func newOptions(opts ...ValidateOption) ValidationOptions {
	cfg := ValidationOptions{Strict: defaultStrictMode, MaxDepth: defaultMaxDepth, Locale: "en", Localizer: defaultLocalizer}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = defaultMaxDepth
	}
	if cfg.Localizer == nil {
		cfg.Localizer = defaultLocalizer
	}
	return cfg
}

// RegisterValidator registers a global custom validator by name.
func RegisterValidator(name string, fn ValidatorFunc) {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	customValidators[name] = fn
}

func Validate(ctx context.Context, model any, opts ...ValidateOption) error {
	cfg := newOptions(opts...)
	if cfg.Tracer != nil {
		var span Span
		ctx, span = cfg.Tracer.Start(ctx, "Validate")
		defer span.End()
	}
	return validateAny(ctx, model, cfg)
}

func ValidateStruct[T any](ctx context.Context, model T, opts ...ValidateOption) error {
	return Validate(ctx, model, opts...)
}

func validateAny(ctx context.Context, model any, cfg ValidationOptions) error {
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
	verr := &ValidationError{Model: meta.Name, Locale: cfg.Locale}
	for _, f := range meta.Fields {
		if err := validateField(ctx, val.FieldByIndex(f.Index), f, cfg, f.Name, 0); err != nil {
			collectFieldErr(verr, err, f.Name)
		}
	}
	if len(verr.Fields) > 0 {
		return verr
	}
	return nil
}

func ParseAndValidate[T any](ctx context.Context, raw []byte, opts ...ValidateOption) (T, error) {
	cfg := newOptions(opts...)
	if cfg.Tracer != nil {
		var span Span
		ctx, span = cfg.Tracer.Start(ctx, "ParseAndValidate")
		defer span.End()
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode input: %w", err)
	}
	if err := Validate(ctx, &out, opts...); err != nil {
		return out, err
	}
	return out, nil
}

func ParseMapToStruct[T any](ctx context.Context, input map[string]any, opts ...ValidateOption) (T, error) {
	var out T
	cfg := newOptions(opts...)
	if cfg.Strict {
		raw, _ := json.Marshal(input)
		if err := json.Unmarshal(raw, &out); err != nil {
			return out, err
		}
	} else {
		coerced, err := coerceMapToType(input, reflect.TypeOf(out))
		if err != nil {
			return out, err
		}
		raw, _ := json.Marshal(coerced)
		if err := json.Unmarshal(raw, &out); err != nil {
			return out, err
		}
	}
	if err := Validate(ctx, out, opts...); err != nil {
		return out, err
	}
	return out, nil
}

func ValidateMap(ctx context.Context, m *Model, input map[string]any, opts ...ValidateOption) error {
	cfg := newOptions(opts...)
	if !m.StrictMode {
		cfg.Strict = false
	}
	compiled := m.Compile()
	return compiled.Validate(ctx, input, append(opts, WithStrict(cfg.Strict))...)
}

func (c *CompiledModel) Validate(ctx context.Context, input map[string]any, opts ...ValidateOption) error {
	cfg := newOptions(opts...)
	verr := &ValidationError{Model: c.Model.Name, Locale: cfg.Locale}
	for _, field := range c.Model.Fields {
		v, ok := input[field.Name]
		if !ok {
			v = field.Default
		}
		if field.RequiredFlag && v == nil {
			verr.Fields = append(verr.Fields, FieldError{Path: field.Name, Rule: "required", Message: cfg.Localizer.Message(cfg.Locale, "required"), Value: "<nil>"})
			continue
		}
		if err := validateDynamicField(ctx, c.Model, field, v, cfg, field.Name, 0); err != nil {
			collectFieldErr(verr, err, field.Name)
		}
	}
	if len(verr.Fields) > 0 {
		return verr
	}
	return nil
}

func validateDynamicField(ctx context.Context, root *Model, f Field, v any, cfg ValidationOptions, path string, depth int) error {
	if depth > cfg.MaxDepth {
		return RuleError{Field: path, Rule: "max_depth", Msg: cfg.Localizer.Message(cfg.Locale, "max_depth")}
	}
	if v == nil {
		return nil
	}
	if !cfg.Strict || f.CoerceFlag {
		cv, err := coerceValue(v, f.Type)
		if err != nil {
			return RuleError{Field: path, Rule: "coerce", Msg: err.Error()}
		}
		v = cv
	}
	if err := ensureType(f.Type, v); err != nil {
		return RuleError{Field: path, Rule: "type", Msg: err.Error()}
	}
	cv := reflect.ValueOf(v)
	if err := validateRules(path, cv, f.Rules, cfg.Localizer, cfg.Locale); err != nil {
		return err
	}
	for _, name := range f.Custom {
		fn, ok := lookupValidator(root, name)
		if !ok {
			return RuleError{Field: path, Rule: name, Msg: "custom validator not found"}
		}
		if err := fn(ctx, v); err != nil {
			return RuleError{Field: path, Rule: name, Msg: err.Error()}
		}
	}
	if f.ModelRef != nil {
		obj, ok := v.(map[string]any)
		if !ok {
			return RuleError{Field: path, Rule: "type", Msg: cfg.Localizer.Message(cfg.Locale, "object_type")}
		}
		if err := f.ModelRef.Compile().Validate(ctx, obj, WithStrict(cfg.Strict), WithLocale(cfg.Locale), WithLocalizer(cfg.Localizer), WithMaxDepth(cfg.MaxDepth)); err != nil {
			return prefixValidationPath(path, err)
		}
	}
	if f.Items != nil {
		arr, ok := v.([]any)
		if !ok {
			return RuleError{Field: path, Rule: "type", Msg: cfg.Localizer.Message(cfg.Locale, "array_type")}
		}
		for i, item := range arr {
			if err := validateDynamicField(ctx, root, *f.Items, item, cfg, fmt.Sprintf("%s[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
	}
	if len(f.Properties) > 0 {
		obj, ok := v.(map[string]any)
		if !ok {
			return RuleError{Field: path, Rule: "type", Msg: cfg.Localizer.Message(cfg.Locale, "object_type")}
		}
		nested := &CompiledModel{Model: &Model{Name: f.Name, Fields: f.Properties, validators: root.validators, StrictMode: root.StrictMode}}
		if err := nested.Validate(ctx, obj, WithStrict(cfg.Strict), WithLocale(cfg.Locale), WithLocalizer(cfg.Localizer), WithMaxDepth(cfg.MaxDepth)); err != nil {
			return prefixValidationPath(path, err)
		}
	}
	if f.MapValue != nil {
		obj, ok := v.(map[string]any)
		if !ok {
			return RuleError{Field: path, Rule: "type", Msg: cfg.Localizer.Message(cfg.Locale, "object_type")}
		}
		for key, mv := range obj {
			if err := validateDynamicField(ctx, root, *f.MapValue, mv, cfg, path+"."+key, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func prefixValidationPath(prefix string, err error) error {
	if ve, ok := err.(*ValidationError); ok && len(ve.Fields) > 0 {
		fe := ve.Fields[0]
		return RuleError{Field: prefix + "." + fe.Path, Rule: fe.Rule, Msg: fe.Message}
	}
	return err
}

func lookupValidator(m *Model, name string) (ValidatorFunc, bool) {
	if m != nil && m.validators != nil {
		if fn, ok := m.validators[name]; ok {
			return fn, true
		}
	}
	validatorMu.RLock()
	fn, ok := customValidators[name]
	validatorMu.RUnlock()
	return fn, ok
}

func validateField(ctx context.Context, v reflect.Value, f fieldMeta, cfg ValidationOptions, path string, depth int) error {
	if depth > cfg.MaxDepth {
		return RuleError{Field: path, Rule: "max_depth", Msg: cfg.Localizer.Message(cfg.Locale, "max_depth")}
	}
	if err := validateRules(path, v, f.Rules, cfg.Localizer, cfg.Locale); err != nil {
		return err
	}
	if v.Kind() == reflect.Struct {
		meta, _ := getOrBuildMetadata(v.Type())
		for _, nf := range meta.Fields {
			if err := validateField(ctx, v.FieldByIndex(nf.Index), nf, cfg, path+"."+nf.Name, depth+1); err != nil {
				return err
			}
		}
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if item.Kind() == reflect.Struct {
				meta, _ := getOrBuildMetadata(item.Type())
				for _, nf := range meta.Fields {
					if err := validateField(ctx, item.FieldByIndex(nf.Index), nf, cfg, fmt.Sprintf("%s[%d].%s", path, i, nf.Name), depth+1); err != nil {
						return err
					}
				}
			}
		}
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

func collectFieldErr(verr *ValidationError, err error, fallback string) {
	var re RuleError
	if errors.As(err, &re) {
		verr.Fields = append(verr.Fields, FieldError{Path: re.Field, Rule: re.Rule, Message: re.Error()})
		return
	}
	verr.Fields = append(verr.Fields, FieldError{Path: fallback, Message: err.Error()})
}

func coerceValue(v any, wantType string) (any, error) {
	switch wantType {
	case "integer":
		switch x := v.(type) {
		case string:
			i, err := strconv.Atoi(x)
			if err != nil {
				return nil, fmt.Errorf("cannot coerce %q to integer", x)
			}
			return i, nil
		case float64:
			return int(x), nil
		}
	case "number":
		switch x := v.(type) {
		case string:
			f, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot coerce %q to number", x)
			}
			return f, nil
		case int:
			return float64(x), nil
		}
	case "boolean":
		switch x := v.(type) {
		case string:
			b, err := strconv.ParseBool(x)
			if err != nil {
				return nil, fmt.Errorf("cannot coerce %q to boolean", x)
			}
			return b, nil
		}
	case "string":
		return fmt.Sprintf("%v", v), nil
	}
	return v, nil
}

func coerceMapToType(in map[string]any, t reflect.Type) (map[string]any, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return in, nil
	}
	out := make(map[string]any, len(in))
	for i := range t.NumField() {
		sf := t.Field(i)
		name := sf.Name
		if j := sf.Tag.Get("json"); j != "" && j != "-" {
			name = firstPart(j)
		}
		v, ok := in[name]
		if !ok {
			continue
		}
		cv, err := coerceValue(v, mapGoTypeToJSON(sf.Type))
		if err != nil {
			return nil, err
		}
		out[name] = cv
	}
	return out, nil
}

func ensureType(want string, v any) error {
	switch want {
	case "string":
		if _, ok := v.(string); !ok {
			return fmt.Errorf("expected string")
		}
	case "integer":
		switch v.(type) {
		case int, int32, int64, float64:
		default:
			return fmt.Errorf("expected integer")
		}
	case "number":
		switch v.(type) {
		case int, int32, int64, float32, float64:
		default:
			return fmt.Errorf("expected number")
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected boolean")
		}
	case "array":
		if _, ok := v.([]any); !ok {
			return fmt.Errorf("expected array")
		}
	case "object":
		if _, ok := v.(map[string]any); !ok {
			return fmt.Errorf("expected object")
		}
	}
	return nil
}

func ClearModelCache() { modelCache = sync.Map{} }
