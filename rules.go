package pydantic

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Rule defines a single validation rule.
type Rule struct {
	Name string `json:"name"`
	Arg  string `json:"arg,omitempty"`
}

// RuleError identifies the failing rule.
type RuleError struct {
	Field string
	Rule  string
	Msg   string
}

func (e RuleError) Error() string { return e.Msg }

func parseRules(tag string) ([]Rule, error) {
	parts := strings.Split(tag, ",")
	out := make([]Rule, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name := p
		arg := ""
		if strings.Contains(p, "=") {
			s := strings.SplitN(p, "=", 2)
			name, arg = s[0], s[1]
		}
		out = append(out, Rule{Name: name, Arg: arg})
	}
	return out, nil
}

func validateRules(field string, v reflect.Value, rules []Rule, l *Localizer, locale string) error {
	for _, r := range rules {
		switch r.Name {
		case "required":
			if isZero(v) {
				return RuleError{Field: field, Rule: r.Name, Msg: l.Message(locale, "required")}
			}
		case "email":
			s := toString(v)
			if !strings.Contains(s, "@") || !strings.Contains(strings.SplitN(s, "@", 2)[1], ".") {
				return RuleError{Field: field, Rule: r.Name, Msg: l.Message(locale, "email")}
			}
		case "min":
			if err := checkMin(v, r.Arg); err != nil {
				return RuleError{Field: field, Rule: r.Name, Msg: err.Error()}
			}
		case "max":
			if err := checkMax(v, r.Arg); err != nil {
				return RuleError{Field: field, Rule: r.Name, Msg: err.Error()}
			}
		case "len":
			n, _ := strconv.Atoi(r.Arg)
			if l := valueLen(v); l != n {
				return RuleError{Field: field, Rule: r.Name, Msg: fmt.Sprintf("length must be %d", n)}
			}
		case "oneof":
			allowed := strings.Fields(r.Arg)
			actual := toString(v)
			found := false
			for _, a := range allowed {
				if actual == a {
					found = true
					break
				}
			}
			if !found {
				return RuleError{Field: field, Rule: r.Name, Msg: l.Message(locale, "oneof")}
			}
		case "regexp":
			re, err := regexp.Compile(r.Arg)
			if err != nil {
				return fmt.Errorf("invalid regexp rule: %w", err)
			}
			if !re.MatchString(toString(v)) {
				return RuleError{Field: field, Rule: r.Name, Msg: l.Message(locale, "regexp")}
			}
		}
	}
	return nil
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		return v.IsNil()
	}
	return v.IsZero()
}

func toString(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprintf("%v", v.Interface())
}

func valueLen(v reflect.Value) int {
	if !v.IsValid() {
		return 0
	}
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len()
	default:
		return len(toString(v))
	}
}

func checkMin(v reflect.Value, arg string) error {
	n, _ := strconv.ParseFloat(arg, 64)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		if float64(v.Len()) < n {
			return fmt.Errorf("length must be >= %s", arg)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < n {
			return fmt.Errorf("must be >= %s", arg)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < n {
			return fmt.Errorf("must be >= %s", arg)
		}
	}
	return nil
}

func checkMax(v reflect.Value, arg string) error {
	n, _ := strconv.ParseFloat(arg, 64)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		if float64(v.Len()) > n {
			return fmt.Errorf("length must be <= %s", arg)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > n {
			return fmt.Errorf("must be <= %s", arg)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > n {
			return fmt.Errorf("must be <= %s", arg)
		}
	}
	return nil
}
