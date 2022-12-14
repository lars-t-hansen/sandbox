package heaps

import (
	"testing"
)

// Once we start passing these functions around, having a "Heap" type that carries
// the comparison operator becomes really appealing

func gt(x int, y int) bool {
	return x > y
}

func mk(x int) int {
	return x
}

func TestIsHeap(t *testing.T) {
	if v := [...]int{3, 2, 1}; !HasHeapProperty[int](v[:], gt) {
		t.Fatalf(`HasHeapProperty returned false but should return true`)
	}
	if w := [...]int{2, 3, 1}; HasHeapProperty[int](w[:], gt) {
		t.Fatalf(`HasHeapProperty returned true but should return false`)
	}
}

func TestBuildHeap(t *testing.T) {
	for i := 0; i < 100; i++ {
		makeTestHeap(t, i, gt, mk)
	}
}

func TestExtract(t *testing.T) {
	for sz := range []int{1, 2, 3, 5, 10, 24, 100} {
		h := makeTestHeap(t, sz, gt, mk)
		hasPrev, prev := false, 0
		for !IsEmpty(h) {
			elem := ExtractMaximum(h)
			if hasPrev && prev < elem {
				t.Fatalf(`Elements extracted in wrong order`)
			}
			if !HasHeapProperty(h.xs, gt) {
				t.Fatalf(`Bad heap in TestExtract after extraction`)
			}
			hasPrev, prev = true, elem
		}
	}
}

func makeTestHeap[T any](t *testing.T, len int, gt func(T, T) bool, mk func(int) T) *Heap[T] {
	a := []T{}
	xs := a[:]
	for j := 0; j < len; j++ {
		xs = append(xs, mk(j))
	}
	h := New(xs, gt)
	if !HasHeapProperty(h.xs, gt) {
		t.Fatalf(`Produced bad heap`)
	}
	return h
}

func isSortedAscending[T any](xs []T, less func(T, T) bool) bool {
	if len(xs) == 0 {
		return true
	}
	for i := 1; i < len(xs); i++ {
		if less(xs[i], xs[i-1]) {
			return false
		}
	}
	return true
}
