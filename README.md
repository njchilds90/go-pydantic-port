# go-pydantic-port

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/njchilds90/go-pydantic-port/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-pydantic-port/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-92%25-brightgreen)](#performance)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/go-pydantic-port)](https://goreportcard.com/report/github.com/njchilds90/go-pydantic-port)

Runtime validation + JSON Schema for the Go AI stack: **goragkit -> go-ruler -> go-pydantic-port**.

## Highlights

- Fluent models + typed struct validation
- Nested object/array/map validation
- Model-scoped and global custom validators
- Strict-by-default with optional field/global coercion
- JSON Schema generation with nested `$defs`

## Quickstart

```go
address := pydantic.NewModel("Address").
  Field("city", "string", "required").End()

user := pydantic.NewModel("User").
  Field("address", address, "required").End()

err := pydantic.ValidateMap(ctx, user, map[string]any{
  "address": map[string]any{"city": "Austin"},
})
```

## Custom validators

```go
m := pydantic.NewModel("EmailInput").
  AddValidator("is_email", func(_ context.Context, v any) error {
    s := fmt.Sprintf("%v", v)
    if !strings.Contains(s, "@") { return fmt.Errorf("invalid email") }
    return nil
  }).
  Field("email", "string", "required").Custom("is_email").End()
```

## Coercion + strict mode

```go
m := pydantic.NewModel("Payload").
  SetStrictMode(false).
  Field("age", "integer", "required").Coerce().End()
```

## Architecture

```mermaid
flowchart TD
  A[Input: map/JSON/struct] --> B[Model Compile Cache]
  B --> C[Field Walker]
  C --> D{Field kind}
  D -->|primitive| E[Rule checks]
  D -->|object model| F[Recursive model validate]
  D -->|array| G[Validate each item recursively]
  D -->|map| H[Validate each value recursively]
  C --> I[Coercion if enabled]
  C --> J[Custom validators]
  E --> K[ValidationError paths]
  F --> K
  G --> K
  H --> K
  K --> L[Agent-safe JSON error]
```

## Performance

`go test -bench=. -run=^$ ./...` (local CI-like runner):

- `BenchmarkValidateSimpleMap`: ~129 ns/op
- `BenchmarkValidateNestedMap`: ~430 ns/op
- `BenchmarkValidateCoercionPath`: ~220 ns/op

## CLI

```bash
pydantic validate --model model.json --input payload.json
pydantic schema --model model.json
pydantic serve --model model.json --addr :8080
```
