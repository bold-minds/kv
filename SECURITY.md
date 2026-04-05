# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** Create a Public Issue

Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

### 2. Report Privately

Send an email to **security@boldminds.tech** with the following information:

- **Subject**: Security Vulnerability in bold-minds/kv
- **Description**: Detailed description of the vulnerability
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Impact**: Potential impact and severity assessment
- **Suggested Fix**: If you have ideas for a fix (optional)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Varies based on complexity, typically within 30 days

### 4. Disclosure Process

1. We will acknowledge receipt of your vulnerability report
2. We will investigate and validate the vulnerability
3. We will develop and test a fix
4. We will coordinate disclosure timing with you
5. We will release a security update
6. We will publicly acknowledge your responsible disclosure (if desired)

## Security Considerations

`kv` is a pure-computation library with a very small attack surface:

- **No network I/O.** `kv` does not make network calls.
- **No file I/O.** `kv` does not read or write files.
- **No reflection.** `kv` uses generic type parameters and concrete type switches only.
- **No external dependencies.** `kv` is pure Go stdlib.
- **Input immutability.** `kv` returns new maps/slices; it never mutates caller-owned input.
- **Nil-safe.** Operations on `nil` maps return empty results.

### Known Limitations

- **Memory on large merges.** `MergeMaps` allocates a new map sized to the combined input. Pathological input (e.g., thousands of large maps) allocates proportionally. Validate input sizes at boundaries where necessary.
- **Recursive option on self-referential maps.** The `Recursive` flag for `NewMap`/`MutateMap` walks nested `map[string]any` and `map[any]any` children. A self-referential map (`m["k"] = m`) will recurse indefinitely and stack-overflow. `kv` does not implement cycle detection. Do not pass self-referential maps, and do not apply `Recursive` to data from untrusted sources without validating structure first.
- **Sort order on heterogeneous keys.** `Sort` / `SortDesc` use type-aware comparison when both keys share a supported type (string, integer, float, bool). When types differ or are unsupported, the comparison falls back to the lexicographic order of `fmt.Sprintf("%v", ...)`. This guarantees determinism but may produce non-intuitive orderings for mixed-type key sets — in practice, maps with a single concrete key type are the common case and sort naturally.
- **Filter functions can panic.** `Filter` takes a user-supplied `func(K, V) bool`. If that function panics, `kv` does not recover. Callers passing filter functions that operate on untrusted data should handle the recover themselves.

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 0.1.1)
- Documented in the CHANGELOG.md
- Announced through GitHub releases
- Tagged with security labels

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who help improve the security of this project.

Thank you for helping keep our project and users safe!
