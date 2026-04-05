# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — Initial release

### Added
- `NewMap[K, V](m map[K]V, opts ...MapOption) map[K]V` — produce a new map with the given options applied in order.
- `MutateMap[K, V](m map[K]V, opts ...MapOption) map[K]V` — same as `NewMap` but reads like mutation at the call site.
- `MergeMaps[K, V](maps ...map[K]V) map[K]V` — shallow-merge with later values overriding earlier ones.
- Map options: `PickKeys`, `OmitKeys`, `OmitValues`, `Invert`, `Recursive` (nested-map traversal flag).
- `GetKeys[K, V](m, opts ...KeysOption) []K` — extract keys with filtering and sorting.
- Keys options: `Filter(fn)`, `Omit(keys...)`, `Sort()`, `SortDesc()`.
- `GetValue[K, V](m, key) V` — zero-value on missing key.
- `GetValueOr[K, V](m, key, defaultValue) V` — caller-supplied default on missing key.
- `GetValues[K, V](m, keys...) []V` — batch value extraction, skipping missing keys.
- `MapOption` and `KeysOption` interfaces defined locally (no external `opts` package).
- Type-aware comparison for `Sort` / `SortDesc`: strings, all integer types, all float types, and bool use natural ordering; mixed or unsupported types fall back to `fmt "%v"` lexicographic order for determinism.
- Regression tests locking in numeric sort order across `int`, `float64`, and `string` key types, plus descending variants.
- Zero external dependencies.

### Requires
- Go 1.21 or later
