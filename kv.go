// Package kv provides ergonomic operations on typed Go maps:
// pick/omit/invert keys, merge, sort keys, filter, and extract values
// with caller-supplied defaults.
//
// Every function operates directly on typed map[K]V — no type-erased
// bridge, no reflection except where explicitly needed for non-
// comparable value equality (OmitValues), no hidden allocations for
// boxing K/V into any.
//
// Immutable by default. Each function returns a new map or slice and
// leaves its input untouched. Where in-place mutation is worth the
// ergonomic cost, it is exposed as an explicit *InPlace variant.
//
// This package operates on typed map[K]V — not on the heterogeneous
// any-trees produced by json.Unmarshal. For nested data navigation
// through map[string]any / map[any]any / []any, see bold-minds/dig.
// For type coercion at extracted values, chain with bold-minds/to.
package kv

import (
	"cmp"
	"reflect"
	"slices"
)

// =============================================================================
// Map-shape operations (immutable — return a new map)
// =============================================================================

// Pick returns a new map containing only the entries whose keys appear
// in keys. Keys that are not present in m are skipped silently.
//
// The order in which keys are supplied has no effect on the returned
// map's iteration order — Go maps are unordered.
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit returns a new map with the specified keys removed. Keys not
// present in m have no effect.
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	if len(keys) == 0 {
		return clone(m)
	}
	excluded := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		excluded[k] = struct{}{}
	}
	result := make(map[K]V, len(m))
	for k, v := range m {
		if _, drop := excluded[k]; !drop {
			result[k] = v
		}
	}
	return result
}

// OmitValues returns a new map with entries removed whose value equals
// any of the supplied values under reflect.DeepEqual. DeepEqual is used
// (rather than ==) so that V may legally include non-comparable types
// such as slices, maps, or structs containing them. The cost is
// O(len(m) × len(values)); for large maps with many exclusions, prefer
// Filter with a purpose-built predicate.
func OmitValues[K comparable, V any](m map[K]V, values ...V) map[K]V {
	if len(values) == 0 {
		return clone(m)
	}
	result := make(map[K]V, len(m))
entries:
	for k, v := range m {
		for _, excluded := range values {
			if reflect.DeepEqual(v, excluded) {
				continue entries
			}
		}
		result[k] = v
	}
	return result
}

// Invert returns a new map whose keys and values are swapped. V must
// be comparable so it can serve as a map key. If multiple keys in m
// share the same value, exactly one wins; which one is unspecified,
// matching Go's own map iteration order.
func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Merge returns a new map containing the union of all provided maps.
// For keys present in more than one map, the value from the last map
// containing the key wins.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	total := 0
	for _, m := range maps {
		total += len(m)
	}
	result := make(map[K]V, total)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Filter returns a new map containing only the entries of m for which
// pred returns true.
func Filter[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		if pred(k, v) {
			result[k] = v
		}
	}
	return result
}

// clone is an internal helper — the stdlib maps.Clone exists in 1.21+
// but allocating explicitly keeps our dependency surface at zero beyond
// cmp, reflect, and slices.
func clone[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// =============================================================================
// Map-shape operations (in place — mutate and return the same map)
// =============================================================================

// PickInPlace removes every entry from m whose key is not in keys, and
// returns m for call-site chaining. If m is nil it is returned
// unchanged.
func PickInPlace[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	if m == nil {
		return m
	}
	keep := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		keep[k] = struct{}{}
	}
	for k := range m {
		if _, ok := keep[k]; !ok {
			delete(m, k)
		}
	}
	return m
}

// OmitInPlace deletes the specified keys from m in place and returns m.
// Missing keys are silently ignored. If m is nil it is returned
// unchanged.
func OmitInPlace[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	for _, k := range keys {
		delete(m, k)
	}
	return m
}

// FilterInPlace removes entries from m for which pred returns false,
// and returns m. The semantics are the inverse of stdlib
// maps.DeleteFunc (which removes where the predicate is true) and match
// the immutable Filter.
func FilterInPlace[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	for k, v := range m {
		if !pred(k, v) {
			delete(m, k)
		}
	}
	return m
}

// =============================================================================
// Key extraction
// =============================================================================

// Keys returns a slice containing every key of m in unspecified order.
func Keys[K comparable, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// SortedKeys returns the keys of m in ascending order, compared with
// cmp.Compare. K must satisfy cmp.Ordered — strings, integer types,
// and float types. NaN floats sort before any non-NaN value, matching
// cmp.Compare's defined behavior.
//
// For key types that are comparable but not cmp.Ordered (bool, custom
// structs, pointers), use SortedKeysFunc.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	result := Keys(m)
	slices.Sort(result)
	return result
}

// SortedKeysDesc returns the keys of m in descending order.
func SortedKeysDesc[K cmp.Ordered, V any](m map[K]V) []K {
	result := Keys(m)
	slices.SortFunc(result, func(a, b K) int { return cmp.Compare(b, a) })
	return result
}

// SortedKeysFunc returns the keys of m sorted by the supplied
// comparator, which must return -1/0/+1 as a<b / a==b / a>b.
func SortedKeysFunc[K comparable, V any](m map[K]V, cmpFn func(a, b K) int) []K {
	result := Keys(m)
	slices.SortFunc(result, cmpFn)
	return result
}

// FilteredKeys returns the keys of m for which pred returns true, in
// unspecified order.
func FilteredKeys[K comparable, V any](m map[K]V, pred func(K, V) bool) []K {
	result := make([]K, 0, len(m))
	for k, v := range m {
		if pred(k, v) {
			result = append(result, k)
		}
	}
	return result
}

// =============================================================================
// Value extraction
// =============================================================================

// Value returns m[key]. If key is absent the zero value of V is
// returned. Equivalent to plain m[key]; provided for naming symmetry
// with ValueOr and Values.
func Value[K comparable, V any](m map[K]V, key K) V {
	return m[key]
}

// ValueOr returns m[key] or def if key is absent.
func ValueOr[K comparable, V any](m map[K]V, key K, def V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

// Values returns the subset of m consisting of those keys that exist.
// The return type preserves key↔value correspondence, making missing
// keys detectable (compare len(result) to len(keys), or probe by key).
// This replaces an earlier []V-returning variant that silently dropped
// missing keys and destroyed positional correspondence.
func Values[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}
