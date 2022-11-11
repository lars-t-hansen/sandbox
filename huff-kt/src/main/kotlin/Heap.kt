// A heap implementing a priority queue.
// Break equal priorities by always choosing the node first added.

class Heap<T> {
    private class HeapNode<T>(var weight: Int, var serial: Int, var tree: T)

    // Here `size` is the number of active elements, `store` is as least as long as `size`
    // but may be longer, it is grown on demand.  Nodes are not shared outside the Heap and
    // may be reused.
    var size = 0
    private var store = Vector<HeapNode<T>>(0) {throw Exception("Bad")}
    private var serial = 0

    fun extractMax(): Pair<T, Int> {
        assert(size > 0)
        size--
        swap(0, size)
        val max = store[size]
        if (size > 1) {
            heapifyAtZero()
        }
        return Pair(max.tree, max.weight)
    }

    fun insert(weight: Int, tree: T) {
        push(weight, tree)
        var i = size-1
        var p = parent(i)
        while (i > 0 && greater(store[i], store[p])) {
            swap(i, p)
            i = p
            p = parent(i)
        }
    }

    private fun push(weight: Int, tree: T) {
        val loc = size
        size++
        if (loc >= store.size) {
            store.push(HeapNode<T>(weight, serial++, tree))
        } else {
            store[loc].weight = weight
            store[loc].serial = serial++
            store[loc].tree = tree
        }
    }

    private fun heapifyAtZero() {
        var loc = 0
        while (true) {
            var greatest = loc
            val l = left(loc)
            if (l < size && greater(store[l], store[loc])) {
                greatest = l
            }
            val r = right(loc)
            if (r < size && greater(store[r], store[greatest])) {
                greatest = r
            }
            if (greatest == loc) {
                break
            }
            swap(loc, greatest)
            loc = greatest
        }
    }

    private fun swap(i: Int, j: Int) {
        val t = store[i]
        store[i] = store[j]
        store[j] = t
    }

    private fun greater(a: HeapNode<T>, b: HeapNode<T>) = a.weight > b.weight || a.weight == b.weight && a.serial < b.serial

    private fun parent(loc: Int) = (loc - 1) / 2
    private fun left(loc: Int) = (loc * 2) + 1
    private fun right(loc: Int) = (loc + 1) * 2
}
