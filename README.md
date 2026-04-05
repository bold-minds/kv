# kv

[![Go Reference](https://pkg.go.dev/badge/github.com/bold-minds/kv.svg)](https://pkg.go.dev/github.com/bold-minds/kv)
[![Build](https://img.shields.io/github/actions/workflow/status/bold-minds/kv/test.yaml?branch=main&label=tests)](https://github.com/bold-minds/kv/actions/workflows/test.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bold-minds/kv)](go.mod)

**Map operations Go's stdlib leaves out — Pick, Omit, Invert, Merge, Filter, Sort — on typed `map[K]V`.**

Go 1.21's `maps` package gave you `Keys`, `Values`, `Clone`, `Copy`, `DeleteFunc`, `Equal`, `Collect`, and `Insert`. Useful, but missing most of what you actually reach for when shaping data: keep-these-keys, drop-these-keys, invert, merge with override, filter by predicate, sort keys deterministically. `kv` is the rest of the Ruby-style map vocabulary for typed Go maps.

```go
// Pick a subset of keys
trimmed := kv.NewMap(user, kv.PickKeys("id", "email"))

// Merge with later maps overriding earlier
final := kv.MergeMaps(defaults, overrides, cliFlags)

// Invert keys and values
byName := kv.NewMap(byID, kv.Invert[int, string]())

// Keys sorted in true numeric order — not string-of-rune codepoint order
keys := kv.GetKeys(scores, kv.Sort())
```

## ✨ Why kv?

- 🎯 **Typed `map[K]V`, not `map[string]any`** — `kv` operates on your own data structures, not JSON-decoder output. For nested `any` navigation, use [`bold-minds/dig`](https://github.com/bold-minds/dig) instead.
- 🧩 **Options pattern** — `NewMap(m, PickKeys(...), OmitValues(...))` reads top-down. Composable, no method chains.
- 🔢 **Correct numeric sort** — `Sort` / `SortDesc` use `cmp.Compare`-based type-aware comparison. A map keyed by `int` sorts as `1, 2, 10, 20, 100`, not by codepoint.
- 🛡️ **Immutable** — every function returns a new map or slice. Input is never mutated.
- 🪶 **Zero dependencies** — pure Go stdlib (`cmp`, `fmt`, `slices`).

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
    safe := kv.NewMap(user, kv.OmitKeys("password"))
    fmt.Println(safe)

    // Keep only whitelisted fields for a public API response
    public := kv.NewMap(user, kv.PickKeys("id", "email", "role"))
    fmt.Println(public)

    // Merge config layers — later maps override earlier
    defaults := map[string]int{"retries": 3, "timeout": 30, "workers": 1}
    overrides := map[string]int{"timeout": 60}
    final := kv.MergeMaps(defaults, overrides)
    fmt.Println(final) // retries:3, timeout:60, workers:1

    // Sorted keys in true numeric order
    scores := map[int]string{100: "a", 1: "b", 20: "c", 2: "d", 10: "e"}
    fmt.Println(kv.GetKeys(scores, kv.Sort()))
    // → [1 2 10 20 100]
}
```

## 📚 API

### Map-shape operations

| Function | Purpose |
|---|---|
| `NewMap[K, V](m, opts...) map[K]V` | Produce a new map with options applied in order |
| `MutateMap[K, V](m, opts...) map[K]V` | Same as `NewMap` but named for call-site clarity |
| `MergeMaps[K, V](maps...) map[K]V` | Shallow merge; later values override |

**Map options** (`MapOption`):

| Option | Effect |
|---|---|
| `PickKeys(keys...)` | Keep only the listed keys |
| `OmitKeys(keys...)` | Drop the listed keys |
| `OmitValues(values...)` | Drop entries whose values match |
| `Invert[K, V]()` | Swap keys and values |
| `Recursive[K, V]()` | Apply subsequent options to nested `map[string]any` / `map[any]any` children |

### Key operations

| Function | Purpose |
|---|---|
| `GetKeys[K, V](m, opts...) []K` | Extract keys with optional filtering and sorting |

**Keys options** (`KeysOption`):

| Option | Effect |
|---|---|
| `Filter(fn func(K, V) bool)` | Keep keys whose `(key, value)` satisfies `fn` |
| `Omit(keys...)` | Drop the listed keys from the result |
| `Sort()` | Ascending order, type-aware comparison |
| `SortDesc()` | Descending order, same comparison |

### Value operations

| Function | Purpose |
|---|---|
| `GetValue[K, V](m, key) V` | Value or zero |
| `GetValueOr[K, V](m, key, defaultValue) V` | Value or caller default |
| `GetValues[K, V](m, keys...) []V` | Batch lookup; missing keys are skipped |

## 🔍 How sorting works

`Sort` and `SortDesc` use type-aware comparison:

1. If both keys share a supported type — `string`, any integer type, any float type, or `bool` — they're compared with `cmp.Compare` (`false < true` for booleans).
2. Otherwise, the comparison falls back to the lexicographic order of `fmt.Sprintf("%v", ...)`, guaranteeing determinism even for mixed or unsupported key types.

In practice, any map with a single concrete key type (`map[int]T`, `map[string]T`, `map[float64]T`, `map[MyEnum]T` with enum backed by a supported type) sorts naturally without special cases.

## 🧭 `kv` vs `dig` vs stdlib `maps`

| Use case | Reach for |
|---|---|
| Walk `map[string]any` / `map[any]any` / `[]any` from `json.Unmarshal` | [`bold-minds/dig`](https://github.com/bold-minds/dig) |
| Pick/omit/invert/merge on typed `map[K]V` | `kv` (this library) |
| `Keys`, `Values`, `Clone`, `Copy`, `DeleteFunc`, `Equal` | stdlib `maps` |
| Sort keys of a typed map with correct natural order | `kv.GetKeys(m, kv.Sort())` |

`kv` and `dig` operate on different data shapes and are complementary — it's normal to import both.

## 🔗 Related bold-minds libraries

- [`bold-minds/dig`](https://github.com/bold-minds/dig) — nested-data navigation for `map[string]any` / `map[any]any` / `[]any`.
- [`bold-minds/each`](https://github.com/bold-minds/each) — slice operations like find/filter/group. Pairs with `kv.GetValues`.
- [`bold-minds/list`](https://github.com/bold-minds/list) — set operations (union, intersect, difference) on slices. Useful on the results of `GetKeys`.
- [`bold-minds/to`](https://github.com/bold-minds/to) — safe type conversion, for coercing extracted values.

## 🚫 Non-goals

- **No `Dig`.** `kv` intentionally does not ship a nested-data walker. That's `bold-minds/dig`'s job, and `dig` does it better. If you need both, import both.
- **No `Must*` variants.** Every function either returns a zero/empty result on failure or accepts a caller-supplied default.
- **No mutation.** `kv` never mutates the map you pass in. If you want in-place semantics, assign the result back to your variable.
- **No reflection-based introspection.** Options work via type parameters and `any`-map adaptation, not reflection.
- **No cycle detection for `Recursive`.** Self-referential nested maps will recurse indefinitely. Validate input structure beforehand.

## 📄 License

MIT — see [LICENSE](LICENSE).
