package sort

import (
	stdsort "sort"

	"github.com/liyue201/gostl/utils/comparator"
	"github.com/liyue201/gostl/utils/iterator"
)

// StableBy sorts the range [first, last) by the keys extracted with keyOf.
//
// All keys are snapshotted before the first comparison happens, so mutating
// the key source while the sort is in progress does not affect the result.
// Elements whose keys compare equal keep their original relative order.
func StableBy[T any, K any](first, last iterator.RandomAccessIterator[T], keyOf func(T) K, cmp comparator.Comparator[K]) {
	n := last.Position() - first.Position()
	if n < 2 {
		return
	}
	values := make([]T, n)
	keys := make([]K, n)
	it := first.Clone().(iterator.RandomAccessIterator[T])
	for i := 0; i < n; i++ {
		values[i] = it.Value()
		keys[i] = keyOf(values[i])
		it.Next()
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	stdsort.SliceStable(order, func(a, b int) bool {
		return cmp(keys[order[a]], keys[order[b]]) < 0
	})
	it = first.Clone().(iterator.RandomAccessIterator[T])
	for _, src := range order {
		it.SetValue(values[src])
		it.Next()
	}
}
