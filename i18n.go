package pydantic

// Localizer stores message catalogs by locale.
type Localizer struct {
	defaultLocale string
	messages      map[string]map[string]string
}

// NewLocalizer creates a localizer with built-in English and Spanish defaults.
func NewLocalizer(defaultLocale string) *Localizer {
	l := &Localizer{defaultLocale: defaultLocale, messages: map[string]map[string]string{}}
	l.messages["en"] = map[string]string{
		"required":    "value is required",
		"email":       "must be a valid email",
		"oneof":       "value not in allowed set",
		"regexp":      "value does not match pattern",
		"array_type":  "must be an array",
		"object_type": "must be an object",
		"max_depth":   "maximum nesting depth exceeded",
	}
	l.messages["es"] = map[string]string{
		"required":    "el valor es obligatorio",
		"email":       "debe ser un correo válido",
		"oneof":       "valor fuera del conjunto permitido",
		"regexp":      "el valor no coincide con el patrón",
		"array_type":  "debe ser un arreglo",
		"object_type": "debe ser un objeto",
		"max_depth":   "se excedió la profundidad máxima",
	}
	return l
}

// Register adds or updates a locale catalog.
func (l *Localizer) Register(locale string, messages map[string]string) {
	l.messages[locale] = messages
}

// Message resolves a localized message key.
func (l *Localizer) Message(locale, key string) string {
	if locale == "" {
		locale = l.defaultLocale
	}
	if m, ok := l.messages[locale][key]; ok {
		return m
	}
	if m, ok := l.messages[l.defaultLocale][key]; ok {
		return m
	}
	return key
}
