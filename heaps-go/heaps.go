package heaps

// A heap is a priority queue: a mutable set with an extractable greatest
// element.
//
// In a heap h, the elements of h.xs have the heap property: for element n,
//
//     left(n) >= len(h) || h[n] >= h[left(n)]
// and right(n) >= len(h) || h[n] >= h[right(n)]
//
// where left() and right() find the children of n and are defined below.

type Heap[T any] struct {
	xs      []T
	greater func(T, T) bool
}

// Test whether the `xs` have the heap property according to `greater`.

func HasHeapProperty[T any](xs []T, greater func(T, T) bool) bool {
	if len(xs) == 0 {
		return true
	}
	return isHeap(xs, greater, 0)
}

// The values in xs are unsorted.  Reorder them in-place according to the
// predicate so that the slice has the heap property, and return a new Heap
// containing xs and the predicate; the heap acquires ownership of xs.

func New[T any](xs []T, greater func(T, T) bool) *Heap[T] {
	buildHeap(xs, greater)
	return &Heap[T]{xs, greater}
}

// Return the number of elements in the heap.

func Size[T any](h *Heap[T]) int {
	return len(h.xs)
}

// Is the heap empty?

func IsEmpty[T any](h *Heap[T]) bool {
	return len(h.xs) == 0
}

// Return the maximum element of a nonempty heap

func Maximum[T any](h *Heap[T]) T {
	if len(h.xs) == 0 {
		panic("Can't extract maximum from empty heap")
	}
	return h.xs[0]
}

// Return and remove the maximum element of a nonempty heap

func ExtractMaximum[T any](h *Heap[T]) T {
	l := len(h.xs)
	if l == 0 {
		panic("Can't extract maximum from empty heap")
	}
	max := h.xs[0]
	h.xs[0] = h.xs[l-1]
	h.xs = h.xs[0 : l-1]
	if l > 2 {
		heapify(h.xs, h.greater, 0)
	}
	return max
}

// Insert a new element

func Insert[T any](h *Heap[T], x T) {
	// extend slice; the stored value doesn't matter
	h.xs = append(h.xs, x)

	// ascend the tree, moving too-small elements out of the way
	i := len(h.xs)
	for i > 0 && h.greater(x, h.xs[parent(i)]) {
		h.xs[i] = h.xs[parent(i)]
		i = parent(i)
	}
	h.xs[i] = x
}

// The `xs` are unsorted.  Sort them ascending according to `greater`.

func HeapSortAscending[T any](xs []T, greater func(T, T) bool) {
	buildHeap(xs, greater)
	for i := len(xs) - 1; i > 0; i-- {
		xs[0], xs[i] = xs[i], xs[0]
		heapify(xs[0:len(xs)-1], greater, 0)
	}
}

// Test whether the `xs` rooted at `root` have the heap property according to
// `greater`.

func isHeap[T any](xs []T, greater func(T, T) bool, root int) bool {
	if l := left(root); l < len(xs) {
		if greater(xs[l], xs[root]) || !isHeap(xs, greater, l) {
			return false
		}
	}
	if r := right(root); r < len(xs) {
		if greater(xs[r], xs[root]) || !isHeap(xs, greater, r) {
			return false
		}
	}
	return true
}

// Reorder the `xs`` so that they have the heap property according to `greater`.

func buildHeap[T any](xs []T, greater func(T, T) bool) {
	// Elements from len(xs)/2 .. len(xs)-1 are all leaves and are
	// proper heaps already.
	for i := len(xs)/2 - 1; i >= 0; i-- {
		heapify(xs, greater, i)
	}
}

// The children of `xs[loc]` have the heap property, but the element `xs[loc]` may be
// smaller than one of its children.  Readjust the heap starting at `loc` so that
// `h[loc]` also has the heap property.

func heapify[T any](xs []T, greater func(T, T) bool, loc int) {
	if loc >= len(xs) {
		panic("Bad root location to heapify")
	}
	for {
		largest := loc
		if l := left(loc); l < len(xs) && greater(xs[l], xs[loc]) {
			largest = l
		}
		if r := right(loc); r < len(xs) && greater(xs[r], xs[largest]) {
			largest = r
		}
		if largest == loc {
			break
		}
		xs[loc], xs[largest] = xs[largest], xs[loc]
		loc = largest
	}
}

func parent(loc int) int {
	return (loc - 1) / 2
}

func left(loc int) int {
	return loc*2 + 1
}

func right(loc int) int {
	return (loc + 1) * 2
}
