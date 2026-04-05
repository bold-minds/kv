# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed — BREAKING
This release replaces the 0.1.0 option-combinator API with plain generic
functions. Every public name changed. 0.1.0 was published for less than a day
with no known consumers; upgrading is a one-time mechanical find-and-replace.

Rationale: the 0.1.0 API routed every call through a type-erased
`map[any]any` bridge, which produced several load-bearing bugs and a
significant allocation overhead. See the bug fixes below.

#### Migration

| 0.1.0 | 0.2.0 |
|---|---|
| `NewMap(m, PickKeys("a", "b"))` | `Pick(m, "a", "b")` |
| `NewMap(m, OmitKeys("b"))` | `Omit(m, "b")` |
| `NewMap(m, OmitValues(2))` | `OmitValues(m, 2)` |
| `NewMap(m, Invert[K, V]())` | `Invert(m)` — returns `map[V]K` |
| `NewMap(m, Recursive[K, V](), ...)` | removed; use `bold-minds/dig` for nested trees |
| `MutateMap(m, OmitKeys("b"))` | `OmitInPlace(m, "b")` or `Omit(m, "b")` |
| `MergeMaps(a, b)` | `Merge(a, b)` |
| `GetKeys(m)` | `Keys(m)` |
| `GetKeys(m, Sort())` | `SortedKeys(m)` (requires `K cmp.Ordered`) |
| `GetKeys(m, SortDesc())` | `SortedKeysDesc(m)` |
| `GetKeys(m, Filter(fn))` | `FilteredKeys(m, fn)` |
| `GetKeys(m, Omit("b"))` | `FilteredKeys(m, func(k, _) bool { return k != "b" })` |
| `GetValue(m, k)` | `Value(m, k)` |
| `GetValueOr(m, k, def)` | `ValueOr(m, k, def)` |
| `GetValues(m, "a", "c") []V` | `Values(m, "a", "c") map[K]V` — return type changed |

#### Added
- `Pick`, `Omit`, `OmitValues`, `Invert`, `Merge`, `Filter` — immutable map-shape ops.
- `PickInPlace`, `OmitInPlace`, `FilterInPlace` — explicit in-place variants.
- `Keys`, `SortedKeys`, `SortedKeysDesc`, `SortedKeysFunc`, `FilteredKeys`.
- `Value`, `ValueOr`, `Values`.
- Benchmarks covering Pick, Omit, Merge, SortedKeys, Invert at sizes 10 / 1k / 100k.

#### Removed
- `NewMap`, `MutateMap`, `MergeMaps`, `GetKeys`, `GetValue`, `GetValueOr`, `GetValues`.
- `MapOption`, `KeysOption` interfaces and all their implementations (`PickKeys`, `OmitKeys`, `OmitValues` as option, `Invert` as option, `Recursive`, `Filter` as option, `Omit` as option, `Sort`, `SortDesc`).
- The internal `compareAny` type-erased comparator. Sorting now uses `cmp.Ordered` directly; `bool` and other non-Ordered key types are served by `SortedKeysFunc`.

#### Fixed
- **`Invert` now works on any map.** The 0.1.0 `Invert` silently returned an empty map whenever `K != V`, because it routed through `map[any]any` and then re-asserted every entry to the original `K`/`V` types after swapping. `Invert` is now a top-level function returning `map[V]K`, so the type change is expressed in the signature and the bug is impossible to re-introduce.
- **`OmitValues` no longer panics on non-comparable value types.** The 0.1.0 implementation used every value as a map key for its exclusion set, which panicked at runtime if any `V` held a slice, map, or struct containing one. The new implementation uses `reflect.DeepEqual` in a linear scan (O(N·M)), which is safe for any `V`.
- **`OmitInPlace` / `PickInPlace` actually mutate.** The 0.1.0 `MutateMap`, despite its name, built and returned a brand-new map; callers who relied on the name to modify the original map silently dropped their mutations. The new in-place variants operate on the supplied map directly and return the same reference.
- **`Values` preserves key↔value correspondence.** The 0.1.0 `GetValues` returned `[]V` with missing keys silently skipped, so the caller had no way to tell which key each value belonged to. `Values` now returns `map[K]V` containing only the keys that existed.
- **`SortedKeys` handles NaN deterministically.** Sorting goes through `cmp.Compare`, so `NaN` sorts before every non-NaN value — no undefined behavior.
- **No more allocation-per-key boxing.** The old API boxed every key and value into `any` twice per call (once in, once out). Benchmarked `Pick` on `map[string]int` of size 10 is now zero-allocation.

### Tests
- Coverage raised to 98.9% (from a state where several "tests" were `t.Logf` placeholders with no assertions — those are now real assertions).
- Regression tests for both the original `string(rune(val))` int-sort bug and the 0.1.0 Invert silent-empty-map bug.
- Added tests for nil-map tolerance across every read-only function.
- Added tests for non-comparable `OmitValues`, duplicate-value `Invert`, and `SortedKeysFunc` on `bool` keys.

## [0.1.0] — Initial release (superseded by 0.2.0)

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
