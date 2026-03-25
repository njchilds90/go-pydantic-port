# Changelog

## v0.3.0 - 2026-03-25

- Added nested model references in fluent builder (`Field("address", nestedModel, ...)`) with recursive map/array/object validation.
- Added model-scoped custom validators (`AddValidator`) and fluent `.Custom(...)` binding.
- Added strict-mode + field coercion controls (`SetStrictMode`, `.Coerce()`) with clear coercion errors.
- Added nested JSON Schema generation with `$defs` for referenced models.
- Added benchmark suite (`benchmarks_test.go`) and README performance numbers.

## v0.2.0 - 2026-03-24

- Initial AI-first runtime validation baseline.
