// Package kv provides ergonomic operations on typed Go maps:
// pick/omit/invert keys, merge, sort keys with type-aware comparison,
// filter, and extract values with caller-supplied defaults.
//
// This package intentionally does not depend on any other bold-minds
// library. It operates on typed map[K]V — not on the heterogeneous
// any-trees produced by json.Unmarshal. For nested data navigation
// through map[string]any / map[any]any / []any, see bold-minds/dig.
// For type coercion at extracted values, chain with bold-minds/to.
package kv

import (
	"cmp"
	"fmt"
	"slices"
)

// MapOption is an operation applied to a map during NewMap or
// MutateMap. Options are applied in order.
type MapOption interface {
	Apply(m map[any]any) map[any]any
}

// KeysOption is an operation applied to a list of keys during GetKeys.
// Options receive both the key list and the source map (for filters
// that need to inspect values) and return the transformed key list.
type KeysOption interface {
	Apply(m map[any]any, keys []any) []any
}

// compareAny returns -1/0/+1 for a<b / a==b / a>b using type-aware
// comparison. When both operands share a supported ordered type —
// string, any integer type, any float type, or bool — the natural
// ordering for that type is used. When the types differ or are
// unsupported, the comparison falls back to the lexicographic order
// of their fmt.Sprintf("%v", ...) representation so the sort remains
// deterministic.
//
// This replaces an earlier string(rune(val)) approach which produced
// nonsense ordering for integers (sorting by unicode codepoint of the
// lowest byte instead of numeric value).
func compareAny(a, b any) int {
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return cmp.Compare(av, bv)
		}
	case int:
		if bv, ok := b.(int); ok {
			return cmp.Compare(av, bv)
		}
	case int8:
		if bv, ok := b.(int8); ok {
			return cmp.Compare(av, bv)
		}
	case int16:
		if bv, ok := b.(int16); ok {
			return cmp.Compare(av, bv)
		}
	case int32:
		if bv, ok := b.(int32); ok {
			return cmp.Compare(av, bv)
		}
	case int64:
		if bv, ok := b.(int64); ok {
			return cmp.Compare(av, bv)
		}
	case uint:
		if bv, ok := b.(uint); ok {
			return cmp.Compare(av, bv)
		}
	case uint8:
		if bv, ok := b.(uint8); ok {
			return cmp.Compare(av, bv)
		}
	case uint16:
		if bv, ok := b.(uint16); ok {
			return cmp.Compare(av, bv)
		}
	case uint32:
		if bv, ok := b.(uint32); ok {
			return cmp.Compare(av, bv)
		}
	case uint64:
		if bv, ok := b.(uint64); ok {
			return cmp.Compare(av, bv)
		}
	case float32:
		if bv, ok := b.(float32); ok {
			return cmp.Compare(av, bv)
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return cmp.Compare(av, bv)
		}
	case bool:
		if bv, ok := b.(bool); ok {
			switch {
			case av == bv:
				return 0
			case !av:
				return -1
			default:
				return 1
			}
		}
	}
	// Mixed or unsupported types: stable deterministic fallback.
	return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// =============================================================================
// NewMap
// =============================================================================

// NewMap creates a new map with the specified options applied.
func NewMap[K comparable, V any](m map[K]V, opts ...MapOption) map[K]V {
	// Convert to any map for processing
	anyMap := make(map[any]any)
	for k, v := range m {
		anyMap[any(k)] = any(v)
	}

	// Check if recursive flag is present
	recursive := false
	for _, opt := range opts {
		if _, isRecursive := opt.(recursiveOption[K, V]); isRecursive {
			recursive = true
			break
		}
	}

	// Apply each option in sequence with recursive context
	for _, opt := range opts {
		if _, isRecursive := opt.(recursiveOption[K, V]); !isRecursive {
			anyMap = applyWithRecursive(opt, anyMap, recursive)
		}
	}

	// Convert back to typed map
	result := make(map[K]V)
	for k, v := range anyMap {
		if key, ok := k.(K); ok {
			if val, ok := v.(V); ok {
				result[key] = val
			}
		}
	}

	return result
}


// PickKeys keeps only the specified keys
func PickKeys[K comparable](keys ...K) MapOption {
	return pickKeysOption[K]{keys: keys}
}

type pickKeysOption[K comparable] struct {
	keys []K
}

func (o pickKeysOption[K]) Apply(m map[any]any) map[any]any {
	result := make(map[any]any)
	for _, key := range o.keys {
		if val, exists := m[any(key)]; exists {
			result[any(key)] = val
		}
	}
	return result
}

// OmitKeys removes the specified keys
func OmitKeys[K comparable](keys ...K) MapOption {
	return omitKeysOption[K]{keys: keys}
}

type omitKeysOption[K comparable] struct {
	keys []K
}

func (o omitKeysOption[K]) Apply(m map[any]any) map[any]any {
	excluded := make(map[any]bool)
	for _, key := range o.keys {
		excluded[any(key)] = true
	}

	result := make(map[any]any)
	for k, v := range m {
		if !excluded[k] {
			result[k] = v
		}
	}
	return result
}

// OmitValues removes entries with the specified values
func OmitValues[K comparable](values ...K) MapOption {
	return omitValuesOption[K]{values: values}
}

type omitValuesOption[K comparable] struct {
	values []K
}

func (o omitValuesOption[K]) Apply(m map[any]any) map[any]any {
	excluded := make(map[any]bool)
	for _, value := range o.values {
		excluded[any(value)] = true
	}

	result := make(map[any]any)
	for k, v := range m {
		if !excluded[v] {
			result[k] = v
		}
	}
	return result
}

// Invert swaps keys and values
func Invert[K, V comparable]() MapOption {
	return invertOption[K, V]{}
}

type invertOption[K, V comparable] struct{}

func (o invertOption[K, V]) Apply(m map[any]any) map[any]any {
	result := make(map[any]any)
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Recursive enables recursive operation on nested maps
func Recursive[K comparable, V any]() MapOption {
	return recursiveOption[K, V]{}
}

type recursiveOption[K comparable, V any] struct{}

func (o recursiveOption[K, V]) Apply(m map[any]any) map[any]any {
	// This is a flag option - actual work is done in applyWithRecursive
	return m
}

// applyWithRecursive applies an operation with optional recursive behavior
func applyWithRecursive(opt MapOption, m map[any]any, recursive bool) map[any]any {
	if !recursive {
		return opt.Apply(m)
	}

	// Apply recursively
	result := opt.Apply(m)
	
	// Process nested maps
	for k, v := range result {
		if nestedMap, ok := v.(map[string]any); ok {
			// Convert to map[any]any for processing
			anyNestedMap := make(map[any]any)
			for nk, nv := range nestedMap {
				anyNestedMap[any(nk)] = any(nv)
			}
			
			// Apply operation recursively
			processedNested := applyWithRecursive(opt, anyNestedMap, true)
			
			// Convert back to map[string]any
			convertedNested := make(map[string]any)
			for nk, nv := range processedNested {
				if strKey, ok := nk.(string); ok {
					convertedNested[strKey] = nv
				}
			}
			
			result[k] = convertedNested
		} else if nestedMap, ok := v.(map[any]any); ok {
			// Handle map[any]any directly
			result[k] = applyWithRecursive(opt, nestedMap, true)
		}
	}
	
	return result
}

// =============================================================================
// GetKeys
// =============================================================================

// GetKeys extracts keys from a map with optional filtering and sorting
func GetKeys[K comparable, V any](m map[K]V, opts ...KeysOption) []K {
	// Convert map to any for processing
	anyMap := make(map[any]any)
	for k, v := range m {
		anyMap[any(k)] = any(v)
	}

	// Get all keys initially
	keys := make([]any, 0, len(anyMap))
	for k := range anyMap {
		keys = append(keys, k)
	}

	// Apply options
	for _, opt := range opts {
		keys = opt.Apply(anyMap, keys)
	}

	// Convert back to []K
	result := make([]K, 0, len(keys))
	for _, key := range keys {
		if k, ok := key.(K); ok {
			result = append(result, k)
		}
	}

	return result
}


// Filter keeps keys matching the condition
func Filter[K comparable, V any](fn func(K, V) bool) KeysOption {
	return filterOption[K, V]{fn: fn}
}

type filterOption[K comparable, V any] struct {
	fn func(K, V) bool
}

func (o filterOption[K, V]) Apply(m map[any]any, keys []any) []any {
	var result []any
	for _, key := range keys {
		if k, ok := key.(K); ok {
			if v, exists := m[key]; exists {
				if val, ok := v.(V); ok {
					if o.fn(k, val) {
						result = append(result, key)
					}
				}
			}
		}
	}
	return result
}

// Omit excludes specified keys
func Omit[K comparable](keys ...K) KeysOption {
	return omitKeysFilterOption[K]{keys: keys}
}

type omitKeysFilterOption[K comparable] struct {
	keys []K
}

func (o omitKeysFilterOption[K]) Apply(m map[any]any, keys []any) []any {
	excluded := make(map[any]bool)
	for _, key := range o.keys {
		excluded[any(key)] = true
	}

	var result []any
	for _, key := range keys {
		if !excluded[key] {
			result = append(result, key)
		}
	}
	return result
}

// Sort sorts keys in ascending order using type-aware comparison.
// Keys that share a supported ordered type (string, any integer, any
// float, bool) are compared naturally for that type. Mixed-type or
// unsupported-type keys fall back to lexicographic comparison of
// their fmt "%v" representation for determinism.
func Sort() KeysOption {
	return sortKeysOption{}
}

type sortKeysOption struct{}

func (o sortKeysOption) Apply(m map[any]any, keys []any) []any {
	result := make([]any, len(keys))
	copy(result, keys)
	slices.SortFunc(result, compareAny)
	return result
}

// SortDesc sorts keys in descending order using the same type-aware
// comparison as Sort.
func SortDesc() KeysOption {
	return sortKeysDescOption{}
}

type sortKeysDescOption struct{}

func (o sortKeysDescOption) Apply(m map[any]any, keys []any) []any {
	result := make([]any, len(keys))
	copy(result, keys)
	slices.SortFunc(result, func(a, b any) int { return compareAny(b, a) })
	return result
}

// =============================================================================
// GetValue
// =============================================================================

// GetValue gets a value from a map and converts it to type V
func GetValue[K comparable, V any](m map[K]V, key K) V {
	if val, exists := m[key]; exists {
		return val
	}
	var zero V
	return zero
}

// =============================================================================
// GetValueOr
// =============================================================================

// GetValueOr gets a value from a map or returns the default
func GetValueOr[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if val, exists := m[key]; exists {
		return val
	}
	return defaultValue
}

// =============================================================================
// GetValues
// =============================================================================

// GetValues gets multiple values by their keys
func GetValues[K comparable, V any](m map[K]V, keys ...K) []V {
	result := make([]V, 0, len(keys))
	for _, key := range keys {
		if val, exists := m[key]; exists {
			result = append(result, val)
		}
	}
	return result
}

// =============================================================================
// MergeMaps
// =============================================================================

// MergeMaps combines multiple maps, with later values overriding earlier ones
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// =============================================================================
// MutateMap
// =============================================================================

// MutateMap applies modifications to an existing map
func MutateMap[K comparable, V any](m map[K]V, opts ...MapOption) map[K]V {
	// Convert to any map for processing
	anyMap := make(map[any]any)
	for k, v := range m {
		anyMap[any(k)] = any(v)
	}

	// Apply options
	for _, opt := range opts {
		anyMap = opt.Apply(anyMap)
	}

	// Convert back to typed map
	result := make(map[K]V)
	for k, v := range anyMap {
		if key, ok := k.(K); ok {
			if val, ok := v.(V); ok {
				result[key] = val
			}
		}
	}

	return result
}

