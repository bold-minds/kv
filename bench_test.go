package kv_test

import (
	"strconv"
	"testing"

	"github.com/bold-minds/kv"
)

// buildStringIntMap returns a map[string]int of the requested size with
// deterministic contents.
func buildStringIntMap(n int) map[string]int {
	m := make(map[string]int, n)
	for i := 0; i < n; i++ {
		m["key"+strconv.Itoa(i)] = i
	}
	return m
}

func buildIntStringMap(n int) map[int]string {
	m := make(map[int]string, n)
	for i := 0; i < n; i++ {
		m[i] = "v" + strconv.Itoa(i)
	}
	return m
}

// -----------------------------------------------------------------------------
// Pick
// -----------------------------------------------------------------------------

func BenchmarkPick_10(b *testing.B)   { benchPick(b, 10) }
func BenchmarkPick_1k(b *testing.B)   { benchPick(b, 1000) }
func BenchmarkPick_100k(b *testing.B) { benchPick(b, 100_000) }

func benchPick(b *testing.B, n int) {
	m := buildStringIntMap(n)
	// Pick half the keys.
	keys := make([]string, 0, n/2)
	for i := 0; i < n/2; i++ {
		keys = append(keys, "key"+strconv.Itoa(i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.Pick(m, keys...)
	}
}

// -----------------------------------------------------------------------------
// Omit
// -----------------------------------------------------------------------------

func BenchmarkOmit_10(b *testing.B)   { benchOmit(b, 10) }
func BenchmarkOmit_1k(b *testing.B)   { benchOmit(b, 1000) }
func BenchmarkOmit_100k(b *testing.B) { benchOmit(b, 100_000) }

func benchOmit(b *testing.B, n int) {
	m := buildStringIntMap(n)
	keys := []string{"key0", "key1", "key2"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.Omit(m, keys...)
	}
}

// -----------------------------------------------------------------------------
// Merge
// -----------------------------------------------------------------------------

func BenchmarkMerge_10(b *testing.B)   { benchMerge(b, 10) }
func BenchmarkMerge_1k(b *testing.B)   { benchMerge(b, 1000) }
func BenchmarkMerge_100k(b *testing.B) { benchMerge(b, 100_000) }

func benchMerge(b *testing.B, n int) {
	a := buildStringIntMap(n)
	c := buildStringIntMap(n)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.Merge(a, c)
	}
}

// -----------------------------------------------------------------------------
// SortedKeys
// -----------------------------------------------------------------------------

func BenchmarkSortedKeys_10(b *testing.B)   { benchSortedKeys(b, 10) }
func BenchmarkSortedKeys_1k(b *testing.B)   { benchSortedKeys(b, 1000) }
func BenchmarkSortedKeys_100k(b *testing.B) { benchSortedKeys(b, 100_000) }

func benchSortedKeys(b *testing.B, n int) {
	m := buildIntStringMap(n)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.SortedKeys(m)
	}
}

// -----------------------------------------------------------------------------
// Invert
// -----------------------------------------------------------------------------

func BenchmarkInvert_10(b *testing.B)   { benchInvert(b, 10) }
func BenchmarkInvert_1k(b *testing.B)   { benchInvert(b, 1000) }
func BenchmarkInvert_100k(b *testing.B) { benchInvert(b, 100_000) }

func benchInvert(b *testing.B, n int) {
	m := buildStringIntMap(n)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.Invert(m)
	}
}

// -----------------------------------------------------------------------------
// OmitValues — the reflect.DeepEqual hot loop. This is the one operation
// where per-entry cost is non-trivial, so it matters most for tuning.
// -----------------------------------------------------------------------------

func BenchmarkOmitValues_10(b *testing.B)   { benchOmitValues(b, 10) }
func BenchmarkOmitValues_1k(b *testing.B)   { benchOmitValues(b, 1000) }
func BenchmarkOmitValues_100k(b *testing.B) { benchOmitValues(b, 100_000) }

func benchOmitValues(b *testing.B, n int) {
	m := buildStringIntMap(n)
	// Exclude a handful of values to exercise the inner loop.
	vals := []int{0, 1, 2, 3}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.OmitValues(m, vals...)
	}
}

// -----------------------------------------------------------------------------
// Filter
// -----------------------------------------------------------------------------

func BenchmarkFilter_10(b *testing.B)   { benchFilter(b, 10) }
func BenchmarkFilter_1k(b *testing.B)   { benchFilter(b, 1000) }
func BenchmarkFilter_100k(b *testing.B) { benchFilter(b, 100_000) }

func benchFilter(b *testing.B, n int) {
	m := buildStringIntMap(n)
	pred := func(_ string, v int) bool { return v%2 == 0 }
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = kv.Filter(m, pred)
	}
}

// -----------------------------------------------------------------------------
// In-place variants — should allocate substantially less than their
// immutable counterparts. Each iteration rebuilds the source map so
// mutation does not compound across iterations.
// -----------------------------------------------------------------------------

func BenchmarkPickInPlace_1k(b *testing.B)   { benchPickInPlace(b, 1000) }
func BenchmarkPickInPlace_100k(b *testing.B) { benchPickInPlace(b, 100_000) }

func benchPickInPlace(b *testing.B, n int) {
	keys := make([]string, 0, n/2)
	for i := 0; i < n/2; i++ {
		keys = append(keys, "key"+strconv.Itoa(i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := buildStringIntMap(n)
		b.StartTimer()
		_ = kv.PickInPlace(m, keys...)
	}
}

func BenchmarkOmitInPlace_1k(b *testing.B)   { benchOmitInPlace(b, 1000) }
func BenchmarkOmitInPlace_100k(b *testing.B) { benchOmitInPlace(b, 100_000) }

func benchOmitInPlace(b *testing.B, n int) {
	keys := []string{"key0", "key1", "key2"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := buildStringIntMap(n)
		b.StartTimer()
		_ = kv.OmitInPlace(m, keys...)
	}
}

func BenchmarkFilterInPlace_1k(b *testing.B)   { benchFilterInPlace(b, 1000) }
func BenchmarkFilterInPlace_100k(b *testing.B) { benchFilterInPlace(b, 100_000) }

func benchFilterInPlace(b *testing.B, n int) {
	pred := func(_ string, v int) bool { return v%2 == 0 }
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := buildStringIntMap(n)
		b.StartTimer()
		_ = kv.FilterInPlace(m, pred)
	}
}
