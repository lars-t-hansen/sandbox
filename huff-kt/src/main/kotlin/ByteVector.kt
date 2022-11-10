// Resizable byte array.

class ByteVector() {
    private var store = ByteArray(256)
    private var _size = 0
    private fun reserve(n: Int) {
        val free = store.size - _size;
        if (free < n) {
            val required = n - free
            val reserved = if (required < store.size*2) store.size*2 else required+256
            val new = ByteArray(reserved)
            for ( i in 0 until _size ) {
                new[i] = store[i]
            }
            store = new
        }
    }
    constructor(ba: ByteArray) : this() {
        store = ba
        _size = ba.size
    }
    constructor(sz: Int) : this() {
        store = ByteArray(sz)
        _size = sz
    }
    operator fun get(i: Int): Byte = store[i]
    operator fun set(i: Int, v: Byte) { store[i] = v }
    operator fun iterator(): Iterator<Byte> {
        return object : Iterator<Byte> {
            var i = 0
            override fun next() : Byte { return store[i++] }
            override fun hasNext(): Boolean { return i < _size }
        }
    }
    val size get() = _size
    val dataref get() = store
    fun push(b: Byte) {
        reserve(1)
        store[_size++] = b
    }
    fun shrinkTo(target: Int) {
        assert(target >= 0 && target <= _size)
        _size = target
    }
    fun growTo(target: Int) {
        assert(target >= 0)
        if (target > _size) {
            reserve(target - _size)
            _size = target
        }
    }
}
