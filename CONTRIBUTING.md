# Contributing to `kv`

Thanks for your interest in contributing. This guide covers the operational process. For the **why** — the design principles every contribution is tested against — see **[bold-minds/oss/PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md)**.

## 🎯 Before You Start

Every contribution is measured against the four Bold Minds principles: **outcome naming**, **one way to do each thing**, **get out of the way**, and **non-goals explicit**. If your proposed change doesn't honor these, it will not be merged — not because the maintainers are precious, but because these principles are what make the libraries worth using.

**Read [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md) first.** It's the load-bearing document.

## 🔧 Development Setup

**Requirements:** Go 1.21 or later, Git, Bash.

```bash
git clone https://github.com/bold-minds/kv.git
cd kv
go test ./...              # unit tests
go test -race ./...        # race detection
go test -bench=. ./...     # benchmarks
./scripts/validate.sh      # full validation pipeline (local mode)
./scripts/validate.sh ci   # strict CI mode
```

Your contribution must pass `./scripts/validate.sh ci` before submitting.

## 📁 Project Structure

```
kv/
├── kv.go                  # Implementation (single file)
├── kv_test.go             # Unit tests
├── scripts/
│   └── validate.sh        # Validation pipeline
├── README.md
├── CONTRIBUTING.md        # This file
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── SECURITY.md
├── LICENSE
└── go.mod
```

Keep it flat. No `internal/` directory unless the library grows significantly.

## 🎨 Code Style

### Naming
- Outcome naming per PRINCIPLES.md. If you reach for a dispatcher name (`Apply`, `Mutate`, `Process`, `Handle`), stop and rename.
- The options pattern used by `NewMap`, `MutateMap`, and `GetKeys` is the established idiom — new transformations should be options, not top-level functions, when they fit.

### Error Handling
- Base functions **must not panic**. Nil maps, missing keys, and empty option lists should produce empty-but-valid results.
- `Or` variants take a fallback argument and return it on failure (e.g., `GetValueOr`).
- **No `Must*` variants.**
- User-supplied filter functions may panic — `kv` does not recover. Document this for any new function that accepts a caller callback.

### Documentation
- Every exported function has a doc comment starting with the function name, describing the outcome (not the implementation), and noting edge cases.
- Package-level doc comment in `kv.go`.
- If a new option changes the sort order or comparison semantics, document the exact rule.

### Dependencies
- **Zero external dependencies.** `kv` is pure stdlib (`cmp`, `fmt`, `slices`).
- The `MapOption` and `KeysOption` interfaces are defined locally — do not re-introduce an external `opts` package.

## 🧪 Testing

**Coverage target: 100% of exported functions.**

```bash
go test -v ./...                   # verbose
go test -race ./...                # race detection
go test -cover ./...               # coverage
go test -bench=. -benchmem ./...   # benchmarks with allocations
```

- Table-driven tests preferred for option combinations
- **Sort tests must assert actual ordering**, not just length. The library has history of a bug (string-of-rune codepoint encoding) that length-only tests failed to catch — regression tests locking in numeric ordering are now mandatory.
- Any new test that covers a map/key/value transformation should assert both shape and content of the result.

## 📝 Pull Request Process

### PR Checklist

Before submitting, verify your PR against the four principles:

- [ ] **Outcome naming** — does the function name describe what the caller gets?
- [ ] **One way** — does any existing function (this library, stdlib `maps`, or `bold-minds/dig`) already do this? If yes, stop.
- [ ] **Get out of the way** — can a Go dev use this from the signature alone?
- [ ] **Non-goals** — does this violate any of the library's stated non-goals?

Additionally:
- [ ] Tests cover 100% of new code, including order assertions for any sort-related change
- [ ] Benchmarks added for new exported functions
- [ ] README updated (if adding or changing exported functions)
- [ ] CHANGELOG.md updated
- [ ] `./scripts/validate.sh ci` passes locally

### PR Scope
- **One function per PR** when adding new functionality
- Bug fixes can be grouped if they share a root cause
- Documentation-only changes can be batched

### PR Description Template

```
## What
One sentence describing the change.

## Why
Real-world evidence of the pain this solves. Link to code, open-source example,
or specific stdlib gap.

## Principles Check
- Outcome naming: [how the name passes the "say it aloud" test]
- One way: [verified no stdlib or existing function does this]
- Get out of the way: [signature alone is enough]
- Non-goals: [confirmed no non-goal violated]

## Breaking Changes
None / [describe]
```

## 🆕 Adding a New Function

`kv` covers map-level operations on typed `map[K]V`. New additions are rare and must clear a high bar:

1. Read the library's non-goals in [README.md](README.md) and [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md). If the function violates one, stop.
2. **Check if this belongs in `bold-minds/dig` instead.** If the function walks nested `any` rather than operating on a typed map, it belongs in `dig`, not here.
3. **Check stdlib `maps`.** Go 1.21's `maps` package has `Keys`, `Values`, `Clone`, `Copy`, `DeleteFunc`, `Equal`, `Collect`, `Insert`. If your function duplicates any of those, stop.
4. Apply the four-principles checklist above.
5. **Show real-world evidence.** Either a codebase using the pattern today, or a verifiable open-source example. Theoretical usefulness is not enough.
6. Draft the function signature and README section first. Open a discussion issue for feedback before writing implementation.
7. Implement, test, benchmark, document.
8. Submit PR with one function per PR.

## 🏷️ Versioning and Releases

- **Semantic versioning**: `vMAJOR.MINOR.PATCH`
- **v0.x**: API may change between minor versions (pre-1.0 signaling)
- **v1.0+**: breaking changes require a major version bump
- Every release updates `CHANGELOG.md`
- Releases are tagged in git and published via `go mod` automatically

## 🙏 Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## 📄 License

By contributing, you agree your contributions are licensed under the MIT License (see [LICENSE](LICENSE)).

## Questions?

Open a discussion issue in this repository.
