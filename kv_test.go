package kv_test

import (
	"cmp"
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/bold-minds/kv"
)

// =============================================================================
// Pick
// =============================================================================

func Test_Pick(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	got := kv.Pick(m, "a", "c")
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Pick() = %v, want %v", got, want)
	}

	// Input is not mutated.
	if len(m) != 3 {
		t.Errorf("Pick mutated input: %v", m)
	}

	// Missing keys are silently skipped.
	got = kv.Pick(m, "a", "missing")
	if !reflect.DeepEqual(got, map[string]int{"a": 1}) {
		t.Errorf("Pick with missing key = %v, want {a:1}", got)
	}

	// Picking zero keys yields an empty map (not nil).
	got = kv.Pick(m)
	if got == nil || len(got) != 0 {
		t.Errorf("Pick() with no keys = %v, want empty non-nil map", got)
	}

	// Pick from nil.
	if g := kv.Pick[string, int](nil, "a"); len(g) != 0 {
		t.Errorf("Pick(nil) = %v, want empty", g)
	}
}

// =============================================================================
// Omit
// =============================================================================

func Test_Omit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	got := kv.Omit(m, "b")
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Omit() = %v, want %v", got, want)
	}

	// Input not mutated.
	if len(m) != 3 || m["b"] != 2 {
		t.Errorf("Omit mutated input: %v", m)
	}

	// Omitting nothing returns a clone, not the same map.
	got = kv.Omit(m)
	if !reflect.DeepEqual(got, m) {
		t.Errorf("Omit() with no keys = %v, want clone %v", got, m)
	}
	got["sentinel"] = 99
	if _, present := m["sentinel"]; present {
		t.Error("Omit() with no keys returned a shared reference, not a clone")
	}

	// Missing keys are tolerated.
	got = kv.Omit(m, "missing")
	if !reflect.DeepEqual(got, m) {
		t.Errorf("Omit with missing key = %v, want %v", got, m)
	}
}

// =============================================================================
// OmitValues
// =============================================================================

func Test_OmitValues_Comparable(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 2}

	got := kv.OmitValues(m, 2)
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OmitValues() = %v, want %v", got, want)
	}

	if len(m) != 4 {
		t.Errorf("OmitValues mutated input: %v", m)
	}

	// Multiple values.
	got = kv.OmitValues(m, 1, 2)
	if !reflect.DeepEqual(got, map[string]int{"c": 3}) {
		t.Errorf("OmitValues(1,2) = %v, want {c:3}", got)
	}

	// Empty values list returns a clone.
	got = kv.OmitValues(m)
	if !reflect.DeepEqual(got, m) {
		t.Errorf("OmitValues() with no values = %v, want clone", got)
	}
}

// This is bug #2 from the review: the old implementation panicked on
// non-comparable value types because it used them as map keys. The new
// implementation uses reflect.DeepEqual, so slices, maps, and function
// values all work.
func Test_OmitValues_NonComparable(t *testing.T) {
	m := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 6},
		"c": {1, 2, 3},
	}

	// Must not panic.
	got := kv.OmitValues(m, []int{1, 2, 3})
	want := map[string][]int{"b": {4, 5, 6}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OmitValues non-comparable = %v, want %v", got, want)
	}
}

// =============================================================================
// Invert
// =============================================================================

// This is bug #1 from the review. The old implementation silently
// returned an empty map for any Invert where K != V, because after the
// swap it re-asserted to the original K/V types and every entry failed.
func Test_Invert(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := kv.Invert(m)
	want := map[int]string{1: "a", 2: "b", 3: "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Invert() = %v, want %v", got, want)
	}

	// Type is actually map[V]K — verify statically with a concrete use.
	if got[1] != "a" {
		t.Errorf("Invert()[1] = %q, want \"a\"", got[1])
	}

	if len(m) != 3 {
		t.Errorf("Invert mutated input: %v", m)
	}
}

func Test_Invert_SquareTypes(t *testing.T) {
	m := map[string]string{"k": "v", "hello": "world"}
	got := kv.Invert(m)
	want := map[string]string{"v": "k", "world": "hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Invert square = %v, want %v", got, want)
	}
}

func Test_Invert_DuplicateValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 1, "c": 2}
	got := kv.Invert(m)
	// Exactly one of "a" / "b" survives — we don't specify which, but
	// the result must be a well-formed inverted map of size 2.
	if len(got) != 2 {
		t.Errorf("Invert with duplicates: len = %d, want 2", len(got))
	}
	if got[2] != "c" {
		t.Errorf("Invert[2] = %q, want \"c\"", got[2])
	}
	if got[1] != "a" && got[1] != "b" {
		t.Errorf("Invert[1] = %q, want \"a\" or \"b\"", got[1])
	}
}

func Test_Invert_Empty(t *testing.T) {
	got := kv.Invert(map[string]int{})
	if len(got) != 0 {
		t.Errorf("Invert(empty) = %v, want empty", got)
	}
}

// =============================================================================
// Merge
// =============================================================================

func Test_Merge(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 3, "c": 4}

	got := kv.Merge(a, b)
	want := map[string]int{"a": 1, "b": 3, "c": 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge() = %v, want %v", got, want)
	}

	// Originals untouched.
	if a["b"] != 2 || len(b) != 2 {
		t.Errorf("Merge mutated inputs: a=%v b=%v", a, b)
	}
}

func Test_Merge_ThreeLayers(t *testing.T) {
	defaults := map[string]int{"retries": 3, "timeout": 30, "workers": 1}
	overrides := map[string]int{"timeout": 60}
	cli := map[string]int{"workers": 8}

	got := kv.Merge(defaults, overrides, cli)
	want := map[string]int{"retries": 3, "timeout": 60, "workers": 8}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge(3 layers) = %v, want %v", got, want)
	}
}

func Test_Merge_Empty(t *testing.T) {
	got := kv.Merge[string, int]()
	if got == nil || len(got) != 0 {
		t.Errorf("Merge() = %v, want empty non-nil map", got)
	}
}

// Merge must tolerate nil maps mixed with real ones — iterating a nil
// map is zero-iter, so the real maps should pass through unchanged.
func Test_Merge_NilInterleaved(t *testing.T) {
	var nilMap map[string]int
	a := map[string]int{"a": 1}
	c := map[string]int{"c": 3}

	got := kv.Merge(a, nilMap, c, nilMap)
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge(real,nil,real,nil) = %v, want %v", got, want)
	}
}

// =============================================================================
// Filter
// =============================================================================

func Test_Filter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	got := kv.Filter(m, func(_ string, v int) bool { return v%2 == 0 })
	want := map[string]int{"b": 2, "d": 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter() = %v, want %v", got, want)
	}

	if len(m) != 4 {
		t.Error("Filter mutated input")
	}
}

// =============================================================================
// In-place variants
// =============================================================================

func Test_PickInPlace(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	same := kv.PickInPlace(m, "a", "c")

	// Same underlying map instance.
	if reflect.ValueOf(m).Pointer() != reflect.ValueOf(same).Pointer() {
		t.Error("PickInPlace returned a different map")
	}
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("PickInPlace result = %v, want %v", m, want)
	}
}

func Test_OmitInPlace(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	same := kv.OmitInPlace(m, "b")

	if reflect.ValueOf(m).Pointer() != reflect.ValueOf(same).Pointer() {
		t.Error("OmitInPlace returned a different map")
	}
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("OmitInPlace result = %v, want %v", m, want)
	}
}

// The three *InPlace variants must all survive a nil map. PickInPlace
// has an explicit guard; OmitInPlace and FilterInPlace rely on
// delete(nil, k) being a legal no-op and nil-map range being empty.
// These tests pin that behavior so a future refactor can't regress it.
func Test_InPlace_NilMap(t *testing.T) {
	var nilMap map[string]int

	if g := kv.PickInPlace(nilMap, "a"); g != nil {
		t.Errorf("PickInPlace(nil) = %v, want nil", g)
	}
	if g := kv.OmitInPlace(nilMap, "a"); g != nil {
		t.Errorf("OmitInPlace(nil) = %v, want nil", g)
	}
	if g := kv.FilterInPlace(nilMap, func(string, int) bool { return true }); g != nil {
		t.Errorf("FilterInPlace(nil) = %v, want nil", g)
	}
}

func Test_FilterInPlace(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	same := kv.FilterInPlace(m, func(_ string, v int) bool { return v%2 == 0 })

	if reflect.ValueOf(m).Pointer() != reflect.ValueOf(same).Pointer() {
		t.Error("FilterInPlace returned a different map")
	}
	want := map[string]int{"b": 2, "d": 4}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("FilterInPlace result = %v, want %v", m, want)
	}
}

// =============================================================================
// Keys / SortedKeys / SortedKeysDesc / SortedKeysFunc / FilteredKeys
// =============================================================================

func Test_Keys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := kv.Keys(m)
	sort.Strings(got)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func Test_SortedKeys_String(t *testing.T) {
	m := map[string]int{"charlie": 3, "alpha": 1, "bravo": 2}
	got := kv.SortedKeys(m)
	want := []string{"alpha", "bravo", "charlie"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}

// Regression test for the string(rune(val)) bug. Under the original
// implementation, int key 10 collated as '\n' and sorted before 2.
// Numeric order must be 1, 2, 10, 20, 100.
func Test_SortedKeys_Int(t *testing.T) {
	m := map[int]string{100: "d", 1: "a", 20: "e", 2: "b", 10: "c"}
	got := kv.SortedKeys(m)
	want := []int{1, 2, 10, 20, 100}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}

func Test_SortedKeys_Float(t *testing.T) {
	m := map[float64]string{3.14: "pi", 2.71: "e", 1.41: "sqrt2"}
	got := kv.SortedKeys(m)
	want := []float64{1.41, 2.71, 3.14}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}

// cmp.Compare puts NaN before every other float. We verify that
// behavior is consistent and that sorting with a NaN key does not
// panic.
func Test_SortedKeys_FloatNaN(t *testing.T) {
	nan := math.NaN()
	m := map[float64]string{3.14: "pi", nan: "nan", 1.41: "s"}
	got := kv.SortedKeys(m)

	if len(got) != 3 {
		t.Fatalf("SortedKeys with NaN: len = %d, want 3", len(got))
	}
	if !math.IsNaN(got[0]) {
		t.Errorf("SortedKeys with NaN: got[0] = %v, want NaN first", got[0])
	}
	if got[1] != 1.41 || got[2] != 3.14 {
		t.Errorf("SortedKeys with NaN: tail = %v, want [1.41 3.14]", got[1:])
	}
}

func Test_SortedKeys_Uint(t *testing.T) {
	m := map[uint]string{10: "a", 2: "b", 100: "c", 1: "d"}
	got := kv.SortedKeys(m)
	want := []uint{1, 2, 10, 100}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys uint = %v, want %v", got, want)
	}
}

// SortedKeysDesc reverses cmp.Compare, so NaN — which cmp.Compare treats
// as less than every non-NaN value — ends up at the tail of the descending
// result. This is the documented asymmetry with Test_SortedKeys_FloatNaN.
func Test_SortedKeysDesc_FloatNaN(t *testing.T) {
	nan := math.NaN()
	m := map[float64]string{3.14: "pi", nan: "nan", 1.41: "s"}
	got := kv.SortedKeysDesc(m)

	if len(got) != 3 {
		t.Fatalf("SortedKeysDesc with NaN: len = %d, want 3", len(got))
	}
	if got[0] != 3.14 || got[1] != 1.41 {
		t.Errorf("SortedKeysDesc with NaN: head = %v, want [3.14 1.41]", got[:2])
	}
	if !math.IsNaN(got[2]) {
		t.Errorf("SortedKeysDesc with NaN: got[2] = %v, want NaN last", got[2])
	}
}

func Test_SortedKeysDesc_Int(t *testing.T) {
	m := map[int]string{100: "d", 1: "a", 20: "e", 2: "b", 10: "c"}
	got := kv.SortedKeysDesc(m)
	want := []int{100, 20, 10, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeysDesc = %v, want %v", got, want)
	}
}

func Test_SortedKeysFunc(t *testing.T) {
	// Sort bool keys (bool isn't cmp.Ordered so the user must supply
	// a comparator). false < true.
	m := map[bool]string{true: "t", false: "f"}
	got := kv.SortedKeysFunc(m, func(a, b bool) int {
		switch {
		case a == b:
			return 0
		case !a:
			return -1
		default:
			return 1
		}
	})
	want := []bool{false, true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeysFunc bool = %v, want %v", got, want)
	}
}

func Test_FilteredKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	got := kv.FilteredKeys(m, func(_ string, v int) bool { return v > 2 })
	sort.Strings(got)
	want := []string{"c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FilteredKeys = %v, want %v", got, want)
	}
}

// =============================================================================
// ValueOr / Values
// =============================================================================

func Test_ValueOr(t *testing.T) {
	m := map[string]int{"a": 1}

	if got := kv.ValueOr(m, "a", 99); got != 1 {
		t.Errorf("ValueOr present = %d, want 1", got)
	}
	if got := kv.ValueOr(m, "missing", 99); got != 99 {
		t.Errorf("ValueOr missing = %d, want 99", got)
	}
}

// This is bug #4 from the review: the old GetValues returned []V with
// no way to detect which keys had been dropped. Values returns a
// map[K]V subset so correspondence is preserved.
func Test_Values(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	got := kv.Values(m, "a", "c")
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Values() = %v, want %v", got, want)
	}

	// Missing key detectable.
	got = kv.Values(m, "a", "missing", "c")
	want = map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Values with missing = %v, want %v", got, want)
	}
	if _, present := got["missing"]; present {
		t.Error("Values: missing key should not appear in result")
	}
}

// =============================================================================
// Nil map tolerance — every read-only function must survive a nil map
// =============================================================================

func Test_Nil_Maps(t *testing.T) {
	var nilMap map[string]int

	if g := kv.Pick(nilMap, "a"); len(g) != 0 {
		t.Errorf("Pick(nil) = %v", g)
	}
	if g := kv.Omit(nilMap, "a"); len(g) != 0 {
		t.Errorf("Omit(nil) = %v", g)
	}
	if g := kv.OmitValues(nilMap, 1); len(g) != 0 {
		t.Errorf("OmitValues(nil) = %v", g)
	}
	if g := kv.Invert(nilMap); len(g) != 0 {
		t.Errorf("Invert(nil) = %v", g)
	}
	if g := kv.Merge(nilMap, nilMap); len(g) != 0 {
		t.Errorf("Merge(nil,nil) = %v", g)
	}
	if g := kv.Filter(nilMap, func(string, int) bool { return true }); len(g) != 0 {
		t.Errorf("Filter(nil) = %v", g)
	}
	if g := kv.Keys(nilMap); len(g) != 0 {
		t.Errorf("Keys(nil) = %v", g)
	}
	if g := kv.SortedKeys(nilMap); len(g) != 0 {
		t.Errorf("SortedKeys(nil) = %v", g)
	}
	if g := kv.FilteredKeys(nilMap, func(string, int) bool { return true }); len(g) != 0 {
		t.Errorf("FilteredKeys(nil) = %v", g)
	}
	if g := kv.ValueOr(nilMap, "a", 42); g != 42 {
		t.Errorf("ValueOr(nil) = %v, want 42", g)
	}
	if g := kv.Values(nilMap, "a"); len(g) != 0 {
		t.Errorf("Values(nil) = %v", g)
	}
}

// =============================================================================
// Composition — verify you can chain operations
// =============================================================================

func Test_Composition(t *testing.T) {
	user := map[string]any{
		"id":       42,
		"email":    "alice@example.com",
		"password": "redacted",
		"role":     "admin",
	}

	public := kv.Pick(kv.Omit(user, "password"), "id", "email")
	want := map[string]any{"id": 42, "email": "alice@example.com"}
	if !reflect.DeepEqual(public, want) {
		t.Errorf("chained Pick(Omit) = %v, want %v", public, want)
	}
}

// =============================================================================
// Sanity: SortedKeysFunc matches SortedKeys for Ordered types
// =============================================================================

func Test_SortedKeysFunc_MatchesSortedKeys(t *testing.T) {
	m := map[int]string{3: "c", 1: "a", 2: "b"}
	a := kv.SortedKeys(m)
	b := kv.SortedKeysFunc(m, cmp.Compare[int])
	if !reflect.DeepEqual(a, b) {
		t.Errorf("SortedKeysFunc(cmp.Compare) = %v, SortedKeys = %v", b, a)
	}
}
