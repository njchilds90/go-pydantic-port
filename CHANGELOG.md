# Changelog

## v0.2.0 - 2026-03-24

### Added
- Fluent model builder for dynamic AI-agent payload validation.
- JSON Schema generation from fluent models and typed structs.
- Rich `ValidationError` with field-level details.
- Struct tag validation engine (`required`, `email`, `min`, `max`, `len`, `oneof`, `regexp`).
- Model metadata caching for faster repeated validations.
- `ParseAndValidate[T]` helper for LLM structured output parsing.
- CLI (`validate`, `schema`, `serve`) in `cmd/pydantic`.
- Optional integrations for goragkit, go-ruler, and OpenTelemetry.
- CI workflow with race detector, golangci-lint, and coverage.

### Changed
- Module path standardized to `github.com/njchilds90/go-pydantic-port`.
- README rewritten for AI-first workflows and production usage.
