# kv

[![Go Reference](https://pkg.go.dev/badge/github.com/bold-minds/kv.svg)](https://pkg.go.dev/github.com/bold-minds/kv)
[![Build](https://img.shields.io/github/actions/workflow/status/bold-minds/kv/test.yaml?branch=main&label=tests)](https://github.com/bold-minds/kv/actions/workflows/test.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bold-minds/kv)](go.mod)

**Map operations Go's stdlib leaves out — Pick, Omit, Invert, Merge, Filter, Sort — on typed `map[K]V`.**

Go 1.21's `maps` package gave you `Keys`, `Values`, `Clone`, `Copy`, `DeleteFunc`, `Equal`, and a few others. Useful, but missing most of what you actually reach for when shaping data: keep-these-keys, drop-these-keys, invert, merge with override, filter by predicate, sorted key extraction. `kv` is the rest of the Ruby-style map vocabulary for typed Go maps.

```go
// Pick a subset of keys
trimmed := kv.Pick(user, "id", "email")

// Merge with later maps overriding earlier
final := kv.Merge(defaults, overrides, cliFlags)

// Invert keys and values — K and V swap, return type is map[V]K
byName := kv.Invert(byID) // map[int]string → map[string]int

// Keys sorted in true numeric order
keys := kv.SortedKeys(scores) // []int: [1, 2, 10, 20, 100]
```

## ✨ Why kv?

- 🎯 **Operates on typed `map[K]V`** — no `map[string]any` bridge, no reflection in the hot path, full type safety at the call site.
- 🧩 **Plain functions, not combinators** — every operation is a top-level generic function. Compose by nesting: `kv.Pick(kv.Omit(m, "password"), "id", "email")`.
- 🔢 **Correct numeric sort** — `SortedKeys` / `SortedKeysDesc` use `cmp.Compare` via the `cmp.Ordered` constraint. A `map[int]T` sorts as `1, 2, 10, 20, 100`, not by codepoint.
- 🛡️ **Immutable by default** — every function returns a new map or slice. Input is never mutated. Explicit `*InPlace` variants are available where in-place mutation is worth it.
- 🪶 **Zero dependencies** — pure Go stdlib (`cmp`, `reflect`, `slices`). `reflect` is used only by `OmitValues` to permit non-comparable value types.

## 📦 Installation

```bash
go get github.com/bold-minds/kv
```

Requires Go 1.21 or later.

## 🎯 Quick Start

```go
package main

import (
    "fmt"

    "github.com/bold-minds/kv"
)

func main() {
    user := map[string]any{
        "id":       42,
        "email":    "alice@example.com",
        "password": "redacted",
        "role":     "admin",
    }

    // Strip sensitive fields before logging
    safe := kv.Omit(user, "password")
    fmt.Println(safe)

    // Keep only whitelisted fields for a public API response
    public := kv.Pick(user, "id", "email", "role")
    fmt.Println(public)

    // Merge config layers — later maps override earlier
    defaults := map[string]int{"retries": 3, "timeout": 30, "workers": 1}
    overrides := map[string]int{"timeout": 60}
    final := kv.Merge(defaults, overrides)
    fmt.Println(final) // retries:3, timeout:60, workers:1

    // Sorted keys in true numeric order
    scores := map[int]string{100: "a", 1: "b", 20: "c", 2: "d", 10: "e"}
    fmt.Println(kv.SortedKeys(scores))
    // → [1 2 10 20 100]
}
```

## 📚 API

### Map-shape operations (immutable)

| Function | Purpose |
|---|---|
| `Pick[K, V](m, keys...) map[K]V` | New map containing only the listed keys |
| `Omit[K, V](m, keys...) map[K]V` | New map with the listed keys removed |
| `OmitValues[K, V](m, values...) map[K]V` | New map with entries whose values match any of `values` removed (uses `reflect.DeepEqual`, so non-comparable V types are fine) |
| `Invert[K, V comparable](m) map[V]K` | Swap keys and values; if duplicate values exist, one arbitrary winner |
| `Merge[K, V](maps...) map[K]V` | Shallow merge; later values override |
| `Filter[K, V](m, pred) map[K]V` | New map containing entries where `pred(k, v)` is true |

### Map-shape operations (in place)

| Function | Purpose |
|---|---|
| `PickInPlace[K, V](m, keys...) map[K]V` | Delete every key not in `keys` from `m`, return `m` |
| `OmitInPlace[K, V](m, keys...) map[K]V` | Delete the listed keys from `m`, return `m` |
| `FilterInPlace[K, V](m, pred) map[K]V` | Delete entries where `pred(k, v)` is false, return `m` |

`FilterInPlace` is the "keep where true" dual of stdlib `maps.DeleteFunc` (which is "delete where true"). Pick whichever polarity reads better at the call site.

### Key extraction

| Function | Purpose |
|---|---|
| `Keys[K, V](m) []K` | All keys, unspecified order |
| `SortedKeys[K cmp.Ordered, V](m) []K` | Ascending order via `cmp.Compare` |
| `SortedKeysDesc[K cmp.Ordered, V](m) []K` | Descending order |
| `SortedKeysFunc[K, V](m, cmp func(a, b K) int) []K` | Custom comparator (for key types that aren't `cmp.Ordered`, e.g. `bool`, custom structs) |
| `FilteredKeys[K, V](m, pred) []K` | Keys where `pred(k, v)` is true |

### Value extraction

| Function | Purpose |
|---|---|
| `Value[K, V](m, key) V` | `m[key]` or zero |
| `ValueOr[K, V](m, key, def) V` | `m[key]` or caller default |
| `Values[K, V](m, keys...) map[K]V` | Subset of `m` consisting of the supplied keys that exist. Returns a map (not a slice), preserving key↔value correspondence and making missing keys detectable. |

## 🔍 How sorting works

`SortedKeys` / `SortedKeysDesc` take `K cmp.Ordered`, so they accept any built-in ordered type: strings, all integer widths, and both float types. Sorting goes through `cmp.Compare`, which produces deterministic ordering even for `NaN` (NaN sorts before every non-NaN value).

For key types that are `comparable` but not `cmp.Ordered` — `bool`, custom structs, pointers, arrays — use `SortedKeysFunc` with your own comparator.

## 🧭 `kv` vs `dig` vs stdlib `maps`

| Use case | Reach for |
|---|---|
| Walk `map[string]any` / `map[any]any` / `[]any` from `json.Unmarshal` | [`bold-minds/dig`](https://github.com/bold-minds/dig) |
| Pick/omit/invert/merge/filter on typed `map[K]V` | `kv` (this library) |
| `Clone`, `Copy`, `DeleteFunc`, `Equal` | stdlib `maps` |
| Sorted keys of a typed map | `kv.SortedKeys(m)` |

`kv` and `dig` operate on different data shapes and are complementary — it's normal to import both.

## 🔗 Related bold-minds libraries

- [`bold-minds/dig`](https://github.com/bold-minds/dig) — nested-data navigation for `map[string]any` / `map[any]any` / `[]any`.
- [`bold-minds/each`](https://github.com/bold-minds/each) — slice operations like find/filter/group.
- [`bold-minds/list`](https://github.com/bold-minds/list) — set operations (union, intersect, difference) on slices. Useful on the results of `Keys`.
- [`bold-minds/to`](https://github.com/bold-minds/to) — safe type conversion, for coercing extracted values.

## 🚫 Non-goals

- **No nested-map walker.** `kv` intentionally does not descend into nested maps or slices. That's [`bold-minds/dig`](https://github.com/bold-minds/dig)'s job. If you need both, import both.
- **No `Must*` variants.** Every function either returns a zero/empty result on missing data or accepts a caller-supplied default.
- **No reflection-based introspection.** The only use of `reflect` is `DeepEqual` inside `OmitValues`, to permit non-comparable V types. Everything else is plain generics.
- **No cycles.** `kv` operates on the top level of a single map — there's nothing to cycle through.

## 📄 License

MIT — see [LICENSE](LICENSE).
