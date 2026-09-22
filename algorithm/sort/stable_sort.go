package sort

import (
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/liyue201/gostl/utils/iterator"
)

//Stable sorts the container by using merge sort
func Stable[T any](first, last iterator.RandomAccessIterator[T], cmp comparator.Comparator[T]) {
	n := last.Position() - first.Position()
	if n <= 1 {
		return
	}
	// Snapshot the values up front, so that keys replaced or mutated while
	// the sort is in progress cannot cut in.
	values := make([]T, n)
	iter := first.Clone().(iterator.RandomAccessIterator[T])
	for i := 0; i < n; i++ {
		values[i] = iter.Value()
		iter.Next()
	}
	tempSlice := make([]T, n)
	mergeSortValues(values, tempSlice, cmp)
	iter = first.Clone().(iterator.RandomAccessIterator[T])
	for i := 0; i < n; i++ {
		iter.SetValue(values[i])
		iter.Next()
	}
}

func mergeSortValues[T any](values, tempSlice []T, cmp comparator.Comparator[T]) {
	if len(values) <= 1 {
		return
	}
	mid := len(values) >> 1
	mergeSortValues(values[:mid], tempSlice[:mid], cmp)
	mergeSortValues(values[mid:], tempSlice[mid:], cmp)
	mergeValues(values, mid, tempSlice, cmp)
}

func mergeValues[T any](values []T, mid int, tempSlice []T, cmp comparator.Comparator[T]) {
	left, right, pos := 0, mid, 0
	for left < mid && right < len(values) {
		if cmp(values[left], values[right]) <= 0 {
			tempSlice[pos] = values[left]
			left++
		} else {
			tempSlice[pos] = values[right]
			right++
		}
		pos++
	}
	for ; left < mid; left++ {
		tempSlice[pos] = values[left]
		pos++
	}
	for ; right < len(values); right++ {
		tempSlice[pos] = values[right]
		pos++
	}
	copy(values, tempSlice[:pos])
}
