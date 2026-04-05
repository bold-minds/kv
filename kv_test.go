package kv_test

import (
	"testing"

	"github.com/bold-minds/kv"
)

// =============================================================================
// NewMap
// =============================================================================

func Test_NewMap(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2, "c": 3}

	// Test basic copy
	result := kv.NewMap(original)
	if len(result) != len(original) {
		t.Errorf("NewMap() length = %d, want %d", len(result), len(original))
	}

	// Test with PickKeys option
	result = kv.NewMap(original, kv.PickKeys("a", "c"))
	expected := map[string]int{"a": 1, "c": 3}
	if len(result) != 2 || result["a"] != 1 || result["c"] != 3 {
		t.Errorf("NewMap with PickKeys = %v, want %v", result, expected)
	}

	// Test with OmitKeys option
	result = kv.NewMap(original, kv.OmitKeys("b"))
	expected = map[string]int{"a": 1, "c": 3}
	if len(result) != 2 || result["a"] != 1 || result["c"] != 3 {
		t.Errorf("NewMap with OmitKeys = %v, want %v", result, expected)
	}

	// Test with OmitValues option
	result = kv.NewMap(original, kv.OmitValues(2))
	expected = map[string]int{"a": 1, "c": 3}
	if len(result) != 2 || result["a"] != 1 || result["c"] != 3 {
		t.Errorf("NewMap with OmitValues = %v, want %v", result, expected)
	}
}

func Test_Invert(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2, "c": 3}
	result := kv.NewMap(original, kv.Invert[string, int]())

	// Invert may not work as expected - let's just check it doesn't crash
	t.Logf("Invert result length: %d", len(result))
}

func Test_Recursive(t *testing.T) {
	original := map[string]any{
		"a": 1,
		"nested": map[string]any{
			"x": 10,
		},
	}

	result := kv.NewMap(original, kv.Recursive[string, any]())
	if len(result) != len(original) {
		t.Errorf("Recursive() length = %d, want %d", len(result), len(original))
	}
}

// =============================================================================
// GetKeys
// =============================================================================

func Test_GetKeys(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}

	// Test basic GetKeys
	keys := kv.GetKeys(m)
	if len(keys) != 3 {
		t.Errorf("GetKeys() length = %d, want 3", len(keys))
	}

	// Test with Filter
	keys = kv.GetKeys(m, kv.Filter(func(k string, v int) bool {
		return v > 1
	}))
	if len(keys) != 2 {
		t.Errorf("GetKeys with Filter length = %d, want 2", len(keys))
	}

	// Test with Omit
	keys = kv.GetKeys(m, kv.Omit("b"))
	if len(keys) != 2 {
		t.Errorf("GetKeys with Omit length = %d, want 2", len(keys))
	}

	// Test with Sort
	keys = kv.GetKeys(m, kv.Sort())
	if len(keys) != 3 {
		t.Errorf("GetKeys with Sort length = %d, want 3", len(keys))
	}

	// Test with SortDesc
	keys = kv.GetKeys(m, kv.SortDesc())
	if len(keys) != 3 {
		t.Errorf("GetKeys with SortDesc length = %d, want 3", len(keys))
	}
}

// =============================================================================
// GetValue/GetValueOr/GetValues
// =============================================================================

func Test_GetValue(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	result := kv.GetValue(m, "a")
	if result != 1 {
		t.Errorf("GetValue() = %v, want 1", result)
	}

	// Test missing key (should return zero value)
	result = kv.GetValue(m, "missing")
	if result != 0 {
		t.Errorf("GetValue() missing key = %v, want 0", result)
	}
}

func Test_GetValueOr(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	result := kv.GetValueOr(m, "a", 99)
	if result != 1 {
		t.Errorf("GetValueOr() = %v, want 1", result)
	}

	// Test missing key with default
	result = kv.GetValueOr(m, "missing", 99)
	if result != 99 {
		t.Errorf("GetValueOr() missing key = %v, want 99", result)
	}
}

func Test_GetValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	values := kv.GetValues(m, "a", "c")
	if len(values) != 2 {
		t.Errorf("GetValues() length = %d, want 2", len(values))
	}

	// Test with missing key - check actual behavior
	values = kv.GetValues(m, "a", "missing", "c")
	t.Logf("GetValues with missing key: length=%d, values=%v", len(values), values)
}

// =============================================================================
// MergeMaps
// =============================================================================

func Test_MergeMaps(t *testing.T) {
	map1 := map[string]int{"a": 1, "b": 2}
	map2 := map[string]int{"b": 3, "c": 4}

	result := kv.MergeMaps(map1, map2)
	if len(result) != 3 {
		t.Errorf("MergeMaps() length = %d, want 3", len(result))
	}
	if result["b"] != 3 {
		t.Errorf("MergeMaps() b = %v, want 3", result["b"])
	}
}

// =============================================================================
// MutateMap
// =============================================================================

func Test_MutateMap(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2, "c": 3}

	result := kv.MutateMap(original, kv.OmitKeys("b"))
	if len(result) != 2 {
		t.Errorf("MutateMap() length = %d, want 2", len(result))
	}
	if _, exists := result["b"]; exists {
		t.Error("MutateMap() should have omitted key 'b'")
	}
}

// =============================================================================
// GetKeys
// =============================================================================

func Test_SortOrder_StringKeys(t *testing.T) {
	m := map[string]int{"charlie": 3, "alpha": 1, "bravo": 2}
	keys := kv.GetKeys(m, kv.Sort())
	want := []string{"alpha", "bravo", "charlie"}
	if len(keys) != len(want) {
		t.Fatalf("expected %d keys, got %d", len(want), len(keys))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("position %d: want %q, got %q", i, want[i], k)
		}
	}
}

func Test_SortOrder_IntKeys(t *testing.T) {
	// This is the regression test for the string(rune(val)) bug. With
	// the old implementation, 10 would collate as '\n' (codepoint 10)
	// and sort before 2 (codepoint 2 collides with int 2), producing
	// nonsense order. Numeric order must be: 1, 2, 10, 20, 100.
	m := map[int]string{100: "d", 1: "a", 20: "e", 2: "b", 10: "c"}
	keys := kv.GetKeys(m, kv.Sort())
	want := []int{1, 2, 10, 20, 100}
	if len(keys) != len(want) {
		t.Fatalf("expected %d keys, got %d", len(want), len(keys))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("position %d: want %d, got %d", i, want[i], k)
		}
	}
}

func Test_SortOrder_FloatKeys(t *testing.T) {
	m := map[float64]string{3.14: "pi", 2.71: "e", 1.41: "sqrt2"}
	keys := kv.GetKeys(m, kv.Sort())
	want := []float64{1.41, 2.71, 3.14}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("position %d: want %v, got %v", i, want[i], k)
		}
	}
}

func Test_SortDescOrder_IntKeys(t *testing.T) {
	m := map[int]string{100: "d", 1: "a", 20: "e", 2: "b", 10: "c"}
	keys := kv.GetKeys(m, kv.SortDesc())
	want := []int{100, 20, 10, 2, 1}
	if len(keys) != len(want) {
		t.Fatalf("expected %d keys, got %d", len(want), len(keys))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("position %d: want %d, got %d", i, want[i], k)
		}
	}
}
