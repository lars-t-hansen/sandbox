// Resizable generic array.

class Vector<T>(sz: Int, init: (Int) -> T) {
    // This is sort of bad but a known Kotlin problem with Arrays, can't use Array<T> here
    private var store : Array<Any?> = Array<Any?>(sz, init)
    private var _size = sz
    private fun reserve(n: Int) {
        val free = store.size - _size;
        if (free < n) {
            val required = n - free
            val reserved = if (required < store.size*2) store.size*2 else required+256
            val new = Array<Any?>(reserved, {null})
            for ( i in 0 until _size ) {
                new[i] = store[i]
            }
            store = new
        }
    }
    operator fun get(i: Int): T = store[i] as T
    operator fun set(i: Int, v: T) { store[i] = v }
    operator fun iterator(): Iterator<T> {
        return object : Iterator<T> {
            var i = 0
            override fun next() : T { return store[i++] as T }
            override fun hasNext(): Boolean { return i < _size }
        }
    }
    val size get() = _size
    fun push(b: T) {
        reserve(1)
        store[_size++] = b
    }
    fun sortWith(cmp: (T,T) -> Int) {
        // This is nuts but OK for now.  The alternative is that I somehow apply Array::sortWith
        // to a subarray, and wrap cmp in a second closure.
        // This sort must be stable but is not, stability must be ensured by the predicate
        for ( i in 0 .. _size-2 ) {
            for ( j in i+1 .. size-1 ) {
                val res = cmp(store[i] as T, store[j] as T)
                if (res > 0) {
                    val tmp = store[i]
                    store[i] = store[j]
                    store[j] = tmp
                }
            }
        }
    }
    fun shrinkTo(target: Int) {
        assert(target >= 0 && target <= _size)
        _size = target
    }
}
