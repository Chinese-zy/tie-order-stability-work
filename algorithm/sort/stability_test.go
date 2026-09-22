package sort

import (
	"math"
	"testing"

	"github.com/liyue201/gostl/ds/vector"
	"github.com/liyue201/gostl/utils/comparator"
)

type rec struct {
	key string
	seq int // registration order
}

func recCmp(a, b rec) int {
	return comparator.StringComparator(a.key, b.key)
}

func recVectorOf(items ...rec) *vector.Vector[rec] {
	v := vector.New[rec]()
	for _, item := range items {
		v.PushBack(item)
	}
	return v
}

func recsOf(v *vector.Vector[rec]) []rec {
	out := make([]rec, 0, v.Size())
	for i := 0; i < v.Size(); i++ {
		out = append(out, v.At(i))
	}
	return out
}

func assertRecsEqual(t *testing.T, got []rec, want []rec) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("position %d = %+v, want %+v; full got %v", i, got[i], want[i], got)
		}
	}
}

// Equal keys must keep their registration order, for both Sort and Stable.
func TestStableSortKeepsRegistrationOrder(t *testing.T) {
	build := func() *vector.Vector[rec] {
		return recVectorOf(
			rec{"b", 0}, rec{"a", 1}, rec{"b", 2},
			rec{"a", 3}, rec{"b", 4}, rec{"a", 5},
		)
	}
	want := []rec{
		{"a", 1}, {"a", 3}, {"a", 5},
		{"b", 0}, {"b", 2}, {"b", 4},
	}
	v := build()
	Sort[rec](v.Begin(), v.End(), recCmp)
	assertRecsEqual(t, recsOf(v), want)

	v = build()
	Stable[rec](v.Begin(), v.End(), recCmp)
	assertRecsEqual(t, recsOf(v), want)
}

// Missing keys, empty keys and blank-only keys are three distinct
// categories: none of them may be folded into another or dumped up front.
func TestSortThreeEmptyKeyCategories(t *testing.T) {
	str := func(s string) *string { return &s }
	rank := func(k *string) int {
		switch {
		case k == nil:
			return 1 // missing
		case *k == "":
			return 2 // empty
		case trimSpace(*k) == "":
			return 3 // blank only
		default:
			return 0 // normal
		}
	}
	type krec struct {
		key *string
		seq int
	}
	cmp := func(a, b krec) int {
		ra, rb := rank(a.key), rank(b.key)
		if ra != rb {
			return comparator.IntComparator(ra, rb)
		}
		if ra == 0 {
			return comparator.StringComparator(*a.key, *b.key)
		}
		return 0
	}
	v := vector.New[krec]()
	for _, item := range []krec{
		{str(""), 0}, {nil, 1}, {str("  "), 2}, {str("m"), 3},
		{str(""), 4}, {nil, 5}, {str("\t"), 6}, {str("a"), 7},
		{str(""), 8}, {nil, 9}, {str(" "), 10},
	} {
		v.PushBack(item)
	}
	Sort[krec](v.Begin(), v.End(), cmp)
	var got []krec
	for i := 0; i < v.Size(); i++ {
		got = append(got, v.At(i))
	}
	want := []krec{
		{str("a"), 7}, {str("m"), 3}, // normal keys, by value
		{nil, 1}, {nil, 5}, {nil, 9}, // missing, registration order
		{str(""), 0}, {str(""), 4}, {str(""), 8}, // empty, registration order
		{str("  "), 2}, {str("\t"), 6}, {str(" "), 10}, // blank, registration order
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].seq != want[i].seq {
			t.Fatalf("position %d seq = %d, want %d; full got %v", i, got[i].seq, want[i].seq, got)
		}
		if (got[i].key == nil) != (want[i].key == nil) {
			t.Fatalf("position %d nil mismatch; full got %v", i, got)
		}
		if got[i].key != nil && *got[i].key != *want[i].key {
			t.Fatalf("position %d key = %q, want %q", i, *got[i].key, *want[i].key)
		}
	}
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// Keys replaced in the middle of a sort must not cut in: the sort runs
// to completion on the keys snapshotted at the moment it started.
func TestSortUsesKeysSnapshottedAtStart(t *testing.T) {
	v := recVectorOf(
		rec{"c", 0}, rec{"a", 1}, rec{"b", 2}, rec{"a", 3}, rec{"c", 4},
	)
	swapped := false
	cmp := func(a, b rec) int {
		if !swapped {
			swapped = true
			// Replace every key mid-sort; late keys must not cut in.
			for i := 0; i < v.Size(); i++ {
				r := v.At(i)
				r.key = "z"
				v.SetAt(i, r)
			}
		}
		return recCmp(a, b)
	}
	Sort[rec](v.Begin(), v.End(), cmp)
	assertRecsEqual(t, recsOf(v), []rec{
		{"a", 1}, {"a", 3}, {"b", 2}, {"c", 0}, {"c", 4},
	})
}

// Array keys are copied into the snapshot; mutating the original arrays
// afterwards must not change the order.
func TestSortArrayKeysCopiedAtSnapshot(t *testing.T) {
	type arec struct {
		key [2]int
		seq int
	}
	v := vector.New[arec]()
	for _, item := range []arec{
		{[2]int{1, 2}, 0}, {[2]int{1, 1}, 1}, {[2]int{0, 9}, 2}, {[2]int{1, 1}, 3},
	} {
		v.PushBack(item)
	}
	mutated := false
	cmp := func(a, b arec) int {
		if !mutated {
			mutated = true
			for i := 0; i < v.Size(); i++ {
				v.SetAt(i, arec{[2]int{99, 99}, v.At(i).seq})
			}
		}
		if c := comparator.IntComparator(a.key[0], b.key[0]); c != 0 {
			return c
		}
		return comparator.IntComparator(a.key[1], b.key[1])
	}
	Sort[arec](v.Begin(), v.End(), cmp)
	want := []arec{
		{[2]int{0, 9}, 2}, {[2]int{1, 1}, 1}, {[2]int{1, 1}, 3}, {[2]int{1, 2}, 0},
	}
	for i := 0; i < v.Size(); i++ {
		if v.At(i) != want[i] {
			t.Fatalf("position %d = %+v, want %+v", i, v.At(i), want[i])
		}
	}
}

// Sorting an already sorted segment again must not swap equal items.
func TestSortAlreadySortedSegmentTwice(t *testing.T) {
	v := recVectorOf(
		rec{"b", 0}, rec{"a", 1}, rec{"b", 2}, rec{"c", 3}, rec{"a", 4}, rec{"b", 5},
	)
	Sort[rec](v.Begin(), v.End(), recCmp)
	once := recsOf(v)
	Sort[rec](v.Begin(), v.End(), recCmp)
	assertRecsEqual(t, recsOf(v), once)
	assertRecsEqual(t, recsOf(v), []rec{
		{"a", 1}, {"a", 4}, {"b", 0}, {"b", 2}, {"b", 5}, {"c", 3},
	})
}

// Inserting a single record into a sorted segment must not swap the
// equal items on either side of the insertion point.
func TestSortInsertOneIntoSortedSegment(t *testing.T) {
	v := recVectorOf(
		rec{"a", 0}, rec{"b", 1}, rec{"b", 2}, rec{"c", 3},
	)
	v.InsertAt(2, rec{"b", 4})
	Sort[rec](v.Begin(), v.End(), recCmp)
	assertRecsEqual(t, recsOf(v), []rec{
		{"a", 0}, {"b", 1}, {"b", 4}, {"b", 2}, {"c", 3},
	})
}

// Multi-level keys: the next level is consulted only when every previous
// level is truly equal, and an empty level stops the descent instead of
// flipping an already decided earlier level.
func TestSortMultiLevelKeys(t *testing.T) {
	type mrec struct {
		l1  int
		l2  string
		l3  string
		seq int
	}
	cmp := func(a, b mrec) int {
		if c := comparator.IntComparator(a.l1, b.l1); c != 0 {
			return c
		}
		if a.l2 == "" || b.l2 == "" {
			return 0
		}
		if c := comparator.StringComparator(a.l2, b.l2); c != 0 {
			return c
		}
		if a.l3 == "" || b.l3 == "" {
			return 0
		}
		return comparator.StringComparator(a.l3, b.l3)
	}
	v := vector.New[mrec]()
	for _, item := range []mrec{
		{2, "", "a", 0},  // third level must not flip the first level
		{1, "", "z", 1},  // empty second level: stop, do not look at third
		{1, "x", "", 2},  // empty third level: stop
		{1, "x", "b", 3}, // equal to seq 2 by the stop rule, keeps order
		{1, "w", "a", 4},
	} {
		v.PushBack(item)
	}
	Sort[mrec](v.Begin(), v.End(), cmp)
	want := []mrec{
		{1, "", "z", 1},
		{1, "w", "a", 4},
		{1, "x", "", 2},
		{1, "x", "b", 3},
		{2, "", "a", 0},
	}
	for i := 0; i < v.Size(); i++ {
		if v.At(i) != want[i] {
			t.Fatalf("position %d = %+v, want %+v", i, v.At(i), want[i])
		}
	}
}

// Deleting a key and registering it back at the end must not shuffle the
// equal items that were never touched.
func TestSortDeleteAndReregisterKeepsUntouchedEquals(t *testing.T) {
	v := recVectorOf(
		rec{"a", 0}, rec{"b", 1}, rec{"b", 2}, rec{"b", 3}, rec{"c", 4},
	)
	v.EraseAt(1)            // delete {"b", 1}
	v.PushBack(rec{"b", 5}) // register it back at the end
	Sort[rec](v.Begin(), v.End(), recCmp)
	assertRecsEqual(t, recsOf(v), []rec{
		{"a", 0}, {"b", 2}, {"b", 3}, {"b", 5}, {"c", 4},
	})
}

// Negative zero and positive zero compare equal, and their registration
// order is preserved.
func TestSortNegativeZeroEqualsPositiveZero(t *testing.T) {
	type frec struct {
		key float64
		seq int
	}
	negZero := math.Copysign(0, -1)
	v := vector.New[frec]()
	for _, item := range []frec{
		{negZero, 0}, {1, 1}, {0, 2}, {negZero, 3}, {-1, 4}, {0, 5},
	} {
		v.PushBack(item)
	}
	cmp := func(a, b frec) int {
		return comparator.Float64Comparator(a.key, b.key)
	}
	Sort[frec](v.Begin(), v.End(), cmp)
	want := []frec{
		{-1, 4}, {negZero, 0}, {0, 2}, {negZero, 3}, {0, 5}, {1, 1},
	}
	for i := 0; i < v.Size(); i++ {
		got := v.At(i)
		if got.seq != want[i].seq || got.key != want[i].key {
			t.Fatalf("position %d = %+v, want %+v", i, got, want[i])
		}
	}
}

// Pages are sorted separately: equal items never borrow registration
// order from another page, and an equal item on a page boundary stays
// on its own page.
func TestSortPagesStayIndependent(t *testing.T) {
	v := recVectorOf(
		rec{"b", 0}, rec{"a", 1}, rec{"b", 2}, // page 1
		rec{"b", 3}, rec{"a", 4}, rec{"c", 5}, // page 2
	)
	pageSize := 3
	Sort[rec](v.IterAt(0), v.IterAt(pageSize), recCmp)
	Sort[rec](v.IterAt(pageSize), v.IterAt(v.Size()), recCmp)
	assertRecsEqual(t, recsOf(v), []rec{
		{"a", 1}, {"b", 0}, {"b", 2},
		{"a", 4}, {"b", 3}, {"c", 5},
	})
}
