package pydantic

import "reflect"

func modelFromType(t reflect.Type, seen map[reflect.Type]bool, depth, maxDepth int) *Model {
	m := NewModel(t.Name())
	if depth > maxDepth {
		return m
	}
	if seen[t] {
		return m
	}
	seen[t] = true
	for i := range t.NumField() {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		name := sf.Name
		if j := sf.Tag.Get("json"); j != "" && j != "-" {
			name = firstPart(j)
		}
		f := Field{Name: name, Type: mapGoTypeToJSON(sf.Type), Description: sf.Tag.Get("description")}
		rules, _ := parseRules(sf.Tag.Get("validate"))
		for _, r := range rules {
			if r.Name == "required" {
				f.RequiredFlag = true
			}
		}
		f.Rules = rules
		base := derefType(sf.Type)
		switch base.Kind() {
		case reflect.Struct:
			if base.PkgPath() != "time" {
				nested := modelFromType(base, seen, depth+1, maxDepth)
				f.Type = "object"
				f.Properties = nested.Fields
			}
		case reflect.Slice, reflect.Array:
			f.Type = "array"
			itemType := derefType(base.Elem())
			f.Items = &Field{Name: name + "_item", Type: mapGoTypeToJSON(itemType)}
			if itemType.Kind() == reflect.Struct {
				nested := modelFromType(itemType, seen, depth+1, maxDepth)
				f.Items.Type = "object"
				f.Items.Properties = nested.Fields
			}
		case reflect.Map:
			f.Type = "object"
			mv := derefType(base.Elem())
			f.MapValue = &Field{Name: name + "_value", Type: mapGoTypeToJSON(mv)}
		}
		m.Fields = append(m.Fields, f)
	}
	return m
}

func mapGoTypeToJSON(t reflect.Type) string {
	t = derefType(t)
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

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func firstPart(v string) string {
	for i, c := range v {
		if c == ',' {
			return v[:i]
		}
	}
	return v
}
