# go-pydantic-port

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/njchilds90/go-pydantic-port/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-pydantic-port/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen)](#performance)
[![Version](https://img.shields.io/badge/version-v0.2.0-purple)](CHANGELOG.md)

Production-grade runtime validation and schema enforcement for Go AI systems.

## Why Go AI teams need this

LLM tool calling and structured output workflows need deterministic runtime validation in Go—without Python sidecars, serialization glue, or dynamic runtime surprises. `go-pydantic-port` gives AI agents:

- **Runtime validation** for typed structs and dynamic tool payloads.
- **JSON Schema generation** for prompts, tool definitions, and contract docs.
- **Rich structured errors** for autonomous retries and debug loops.
- **Optional AI-stack integrations** for goragkit/go-ruler and OTEL traces.

## Installation

```bash
go get github.com/njchilds90/go-pydantic-port@v0.2.0
```

## Quickstart

```go
package main

import (
  "context"
  "fmt"

  pydantic "github.com/njchilds90/go-pydantic-port"
)

func main() {
  m := pydantic.NewModel("ToolPayload").
    Field("query", "string", "required", "min=3").
    Field("top_k", "integer", "min=1", "max=20")

  input := map[string]any{"query": "golang otel", "top_k": 5}
  if err := pydantic.ValidateMap(context.Background(), m, input); err != nil {
    panic(err)
  }

  schema := m.Schema()
  fmt.Println(schema["$schema"])
}
```

## CLI

```bash
pydantic validate --model model.json --input payload.json
pydantic schema --model model.json
pydantic serve --model model.json --addr :8080
```

`serve` exposes:
- `GET /schema`
- `POST /validate`

## Architecture

```mermaid
flowchart LR
  A[Typed Structs / Maps] --> B[Validation Engine]
  B --> C[ValidationError (structured)]
  B --> D[JSON Schema Generator]
  D --> E[LLM Tool/Prompt Contracts]
  B --> F[Optional Integrations]
  F --> G[goragkit]
  F --> H[go-ruler]
  F --> I[OpenTelemetry]
```

## AI-agent examples

- Validate LLM JSON responses via `ParseAndValidate[T]`.
- Build tool input contracts with `NewModel(...).Field(...)` and emit schema.
- Run `goruler.ValidateThenEvaluate` to gate policy decisions.
- Wrap validation spans using `integrations/otel` for observability.

## Performance

Benchmarks are included with `go test -bench=. ./...`.

Current target characteristics:
- reflection metadata cached via `sync.Map`
- zero external runtime dependency in the core package
- deterministic tag parser and low-allocation validation loops

## Roadmap

- Nested object/array schema refinements.
- Localization/i18n for validation errors.
- Pluggable custom validators.

## Ecosystem

Pairs well with:
- [goragkit](https://github.com/njchilds90/goragkit)
- [go-ruler](https://github.com/njchilds90/go-ruler)
- goretry
- go-result

## License

MIT
