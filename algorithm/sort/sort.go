package sort

import (
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/liyue201/gostl/utils/iterator"
)

//Sort sorts the container by using stable sort
func Sort[T any](first, last iterator.RandomAccessIterator[T], cmp comparator.Comparator[T]) {
	Stable(first, last, cmp)
}

func doPivot[T any](first, mid, last iterator.RandomAccessIterator[T], cmp comparator.Comparator[T]) {
	if cmp(first.Value(), mid.Value()) > 0 {
		swapValue(first, mid)
	}
	if cmp(first.Value(), last.Value()) > 0 {
		swapValue(first, last)
	}
	if cmp(mid.Value(), last.Value()) > 0 {
		swapValue(mid, last)
	}
}

func swapValue[T any](a, b iterator.RandomAccessIterator[T]) {
	valA := a.Value()
	valB := b.Value()
	a.SetValue(valB)
	b.SetValue(valA)
}
