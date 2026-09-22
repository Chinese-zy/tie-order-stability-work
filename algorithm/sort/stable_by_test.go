package sort

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/liyue201/gostl/ds/vector"
	"github.com/liyue201/gostl/utils/comparator"
)

type tieRecord struct {
	id  int
	key string
}

func recordIDs(v *vector.Vector[tieRecord]) []int {
	ids := make([]int, 0, v.Size())
	for i := 0; i < v.Size(); i++ {
		ids = append(ids, v.At(i).id)
	}
	return ids
}

func newRecordVector(records []tieRecord) *vector.Vector[tieRecord] {
	v := vector.New[tieRecord]()
	for _, r := range records {
		v.PushBack(r)
	}
	return v
}

// Equal keys must keep registration order (the order records were
// registered), not the traversal order of the key map.
func TestStableByKeepsRegistrationOrder(t *testing.T) {
	keys := map[int]string{
		1: "bb",
		2: "aa",
		3: "bb",
		4: "aa",
		5: "bb",
	}
	registration := []tieRecord{{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}}
	v := newRecordVector(registration)

	StableBy[tieRecord, string](v.Begin(), v.End(),
		func(r tieRecord) string { return keys[r.id] },
		comparator.StringComparator)

	want := []int{2, 4, 1, 3, 5}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Missing keys, empty keys and whitespace-only keys are three distinct
// classes. None of them may be folded into another class or thrown ahead
// of real keys.
func TestStableByEmptyKeyClasses(t *testing.T) {
	type classified struct {
		class int
		key   string
	}
	classOf := func(r tieRecord) classified {
		switch {
		case r.key == "\x00missing":
			return classified{class: 1}
		case r.key == "":
			return classified{class: 2}
		case strings.TrimSpace(r.key) == "":
			return classified{class: 3}
		default:
			return classified{class: 0, key: r.key}
		}
	}
	cmp := func(a, b classified) int {
		if a.class != b.class {
			return comparator.IntComparator(a.class, b.class)
		}
		return comparator.StringComparator(a.key, b.key)
	}

	registration := []tieRecord{
		{id: 1, key: "\x00missing"},
		{id: 2, key: "bb"},
		{id: 3, key: ""},
		{id: 4, key: "aa"},
		{id: 5, key: "  "},
		{id: 6, key: "\x00missing"},
		{id: 7, key: ""},
		{id: 8, key: "\t"},
		{id: 9, key: "aa"},
	}
	v := newRecordVector(registration)

	StableBy[tieRecord, classified](v.Begin(), v.End(), classOf, cmp)

	want := []int{4, 9, 2, 1, 6, 3, 7, 5, 8}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Keys are snapshotted the moment the sort starts. Rewriting the key
// source halfway through must not let the new keys cut in.
func TestStableBySnapshotsKeys(t *testing.T) {
	keys := map[int]int{1: 3, 2: 1, 3: 2, 4: 1}
	registration := []tieRecord{{id: 1}, {id: 2}, {id: 3}, {id: 4}}
	v := newRecordVector(registration)

	rewritten := false
	cmp := func(a, b int) int {
		if !rewritten {
			rewritten = true
			for id := range keys {
				keys[id] = 1000 + id
			}
		}
		return comparator.IntComparator(a, b)
	}

	StableBy[tieRecord, int](v.Begin(), v.End(),
		func(r tieRecord) int { return keys[r.id] }, cmp)

	want := []int{2, 4, 3, 1}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// When a key is a group of numbers, mutating the original array after the
// snapshot must not change the outcome.
func TestStableByArrayKeySnapshot(t *testing.T) {
	keys := map[int][3]int{
		1: {2, 0, 0},
		2: {1, 9, 9},
		3: {1, 9, 8},
	}
	registration := []tieRecord{{id: 1}, {id: 2}, {id: 3}}
	v := newRecordVector(registration)

	cmpArray := func(a, b [3]int) int {
		for i := 0; i < 3; i++ {
			if c := comparator.IntComparator(a[i], b[i]); c != 0 {
				return c
			}
		}
		return 0
	}
	rewritten := false
	cmp := func(a, b [3]int) int {
		if !rewritten {
			rewritten = true
			keys[2] = [3]int{9, 9, 9}
			keys[3] = [3]int{0, 0, 0}
		}
		return cmpArray(a, b)
	}

	StableBy[tieRecord, [3]int](v.Begin(), v.End(),
		func(r tieRecord) [3]int { return keys[r.id] }, cmp)

	want := []int{3, 2, 1}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Sorting an already sorted run again must not swap equal elements.
func TestStableByResortKeepsEqualOrder(t *testing.T) {
	registration := []tieRecord{
		{id: 1, key: "bb"},
		{id: 2, key: "aa"},
		{id: 3, key: "bb"},
		{id: 4, key: "aa"},
		{id: 5, key: "cc"},
	}
	v := newRecordVector(registration)
	keyOf := func(r tieRecord) string { return r.key }

	StableBy[tieRecord, string](v.Begin(), v.End(), keyOf, comparator.StringComparator)
	first := recordIDs(v)
	StableBy[tieRecord, string](v.Begin(), v.End(), keyOf, comparator.StringComparator)
	second := recordIDs(v)

	want := []int{2, 4, 1, 3, 5}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("first sort got %v, want %v", first, want)
	}
	if !reflect.DeepEqual(second, first) {
		t.Fatalf("resort changed order: %v -> %v", first, second)
	}
}

// Inserting a single record into the middle of a sorted run must not swap
// the equal elements on either side of the insertion point.
func TestStableByInsertOneKeepsNeighbors(t *testing.T) {
	sorted := []tieRecord{
		{id: 1, key: "aa"},
		{id: 2, key: "bb"},
		{id: 3, key: "bb"},
		{id: 4, key: "cc"},
	}
	v := newRecordVector(sorted)
	v.InsertAt(3, tieRecord{id: 5, key: "bb"})

	StableBy[tieRecord, string](v.Begin(), v.End(),
		func(r tieRecord) string { return r.key }, comparator.StringComparator)

	want := []int{1, 2, 3, 5, 4}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

type levelKey struct {
	l1 string
	l2 string
	l3 string
}

// Only look at the next level when all previous levels are truly equal.
// An empty second level stops the comparison; the third level must never
// flip an order already decided by the first level.
func compareLevelKey(a, b levelKey) int {
	if c := comparator.StringComparator(a.l1, b.l1); c != 0 {
		return c
	}
	if a.l2 == "" && b.l2 == "" {
		return 0
	}
	if c := comparator.StringComparator(a.l2, b.l2); c != 0 {
		return c
	}
	return comparator.StringComparator(a.l3, b.l3)
}

func TestStableByMultiLevelKeys(t *testing.T) {
	type lvlRecord struct {
		id  int
		key levelKey
	}
	registration := []lvlRecord{
		{id: 1, key: levelKey{l1: "b", l2: "", l3: "aaa"}},
		{id: 2, key: levelKey{l1: "a", l2: "", l3: "zzz"}},
		{id: 3, key: levelKey{l1: "a", l2: "", l3: "aaa"}},
		{id: 4, key: levelKey{l1: "a", l2: "x", l3: "zzz"}},
		{id: 5, key: levelKey{l1: "a", l2: "x", l3: "aaa"}},
	}
	v := vector.New[lvlRecord]()
	for _, r := range registration {
		v.PushBack(r)
	}

	StableBy[lvlRecord, levelKey](v.Begin(), v.End(),
		func(r lvlRecord) levelKey { return r.key }, compareLevelKey)

	got := make([]int, 0, v.Size())
	for i := 0; i < v.Size(); i++ {
		got = append(got, v.At(i).id)
	}
	// l1 decides first: id 1 (l1 "b") goes last even though its l3 is
	// smallest. ids 2 and 3 both stop at the empty l2, so l3 is never
	// consulted and they keep registration order. ids 4 and 5 share l2
	// "x", so l3 orders 5 before 4.
	want := []int{2, 3, 5, 4, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Deleting a key and registering it back in the original order must not
// swap the equal elements that were never touched.
func TestStableByDeleteAndReregister(t *testing.T) {
	registration := []tieRecord{
		{id: 1, key: "aa"},
		{id: 2, key: "bb"},
		{id: 3, key: "bb"},
		{id: 4, key: "bb"},
		{id: 5, key: "cc"},
	}
	v := newRecordVector(registration)

	// delete id 3, then register it back at its original spot
	v.EraseAt(2)
	v.InsertAt(2, tieRecord{id: 3, key: "bb"})

	StableBy[tieRecord, string](v.Begin(), v.End(),
		func(r tieRecord) string { return r.key }, comparator.StringComparator)

	want := []int{1, 2, 3, 4, 5}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Negative zero and positive zero compare equal; registration order
// between them must be preserved.
func TestStableByNegativeZero(t *testing.T) {
	type fRecord struct {
		id  int
		key float64
	}
	negZero := math.Copysign(0, -1)
	registration := []fRecord{
		{id: 1, key: negZero},
		{id: 2, key: 1.5},
		{id: 3, key: 0.0},
		{id: 4, key: -2.5},
		{id: 5, key: negZero},
	}
	v := vector.New[fRecord]()
	for _, r := range registration {
		v.PushBack(r)
	}

	StableBy[fRecord, float64](v.Begin(), v.End(),
		func(r fRecord) float64 { return r.key }, comparator.Float64Comparator)

	got := make([]int, 0, v.Size())
	for i := 0; i < v.Size(); i++ {
		got = append(got, v.At(i).id)
	}
	want := []int{4, 1, 3, 5, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Pages are sorted independently. Equal elements on a page boundary stay
// on their own page; no page borrows registration order from another.
func TestStableByPagesAreIndependent(t *testing.T) {
	registration := []tieRecord{
		// page 1
		{id: 1, key: "bb"},
		{id: 2, key: "aa"},
		{id: 3, key: "bb"},
		// page 2
		{id: 4, key: "bb"},
		{id: 5, key: "aa"},
		{id: 6, key: "bb"},
	}
	v := newRecordVector(registration)
	keyOf := func(r tieRecord) string { return r.key }

	const pageSize = 3
	for start := 0; start < v.Size(); start += pageSize {
		first := v.Begin().IteratorAt(start)
		last := v.Begin().IteratorAt(start + pageSize)
		StableBy[tieRecord, string](first, last, keyOf, comparator.StringComparator)
	}

	want := []int{2, 1, 3, 5, 4, 6}
	if got := recordIDs(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
