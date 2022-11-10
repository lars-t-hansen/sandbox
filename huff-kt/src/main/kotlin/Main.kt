// Huffman compressor / decompressor
//
// (Based on the Go and Rust versions in `sandbox/huff` and `sandbox/huffrs`,
// except this is still not multi-threaded)
//
// huffer compress [-o outfile] filename
//   Creates outfile, or if no -o option, filename.huff
//
// huffer decompress [-o outfile] filename.huff
//   Creates outfile, or if no -o option, filename
//
// Since this is a programming exercise, it works by reading 64KB blocks and
// compressing them individually; the output file consists of compressed blocks.
// Also, we don't care about micro-efficiencies in representing the dictionary
// in the file or in complicated fallback schemes, more could be done.
//
// A compressed block is represented as
//   number of dictionary entries: u16 > 0 (max value is really 256)
//   run of dictionary entries sorted descending by frequency with ties
//   broken by lower byte values first:
//     value: u8
//     frequency: u32 (max value is really 65536)
//   number of encoded bytes: u32 (max value is really 65536)
//   number of bytes used for encoded bytes: u32 (max value 65536)
//   bytes, the number of which is encoded by previous field
//
// An uncompressed block can be written under some circumstances, it is represented as
//   0: u16
//   number of bytes: u32 (really max 65536)
//   bytes, the number of which is encoded by previous field

import java.io.InputStream
import java.io.OutputStream
import java.nio.file.Files

fun main(args: Array<String>) {
    // TODO: Command line parsing and file name computation
    // TODO: Decompression
    compressFile("test.txt", "test.txt.huff")
}

fun compressFile(inFilename: String, outFilename: String) {
    val input = Files.newInputStream(java.nio.file.Paths.get(inFilename))
    val output = Files.newOutputStream(java.nio.file.Paths.get(outFilename))
    compressStream(input, output)
    input.close()
    output.close()
}

fun compressStream(input: InputStream, output: OutputStream) {
    while (true) {
        val inBuf = ByteVector(65536)
        val bytesRead = input.read(inBuf.dataref)
        if (bytesRead == -1) {
            break
        }
        inBuf.shrinkTo(bytesRead)
        val freq = computeFrequencies(inBuf)
        val tree = buildHuffmanTree(freq)
        val dict = buildEncodingDictionary(tree)
        val outBuf = if (dict != null) {
            compressBlock(inBuf, dict) ?: inBuf
        } else {
            inBuf
        }
        val metadata = constructMetadata(dict != null, inBuf, freq, outBuf)
        output.write(metadata.dataref, 0, metadata.size)
        output.write(outBuf.dataref, 0, outBuf.size)
    }
}

fun constructMetadata(wasEncoded: Boolean, inBuf: ByteVector, freq: Vector<FreqItem>, outBuf: ByteVector): ByteVector {
    val m = ByteVector()
    if (wasEncoded) {
        put(m, freq.size.toLong(), 2)
        for (f in freq) {
            put(m, f.byte.toLong(), 1)
            put(m, f.count.toLong(), 4)
        }
        put(m, inBuf.size.toLong(), 4)
        put(m, outBuf.size.toLong(), 4)
    } else {
        put(m, 0, 2)
        put(m, inBuf.size.toLong(), 4)
    }
    return m
}

fun put(out: ByteVector, _v: Long, _nbytes: Int) {
    var v = _v
    var nbytes = _nbytes
    while (nbytes > 0) {
        out.push(v.toByte())
        nbytes--
        v = v ushr 8
    }
}

fun compressBlock(input: ByteVector, dict: Vector<DictItem>): ByteVector? {
    val output = ByteVector(65536)
    var outptr = 0
    var window = 0L
    var width = 0
    for ( b in input ) {
        val d = dict[b.toInt()]
        window = (d.bits shl width) or window
        width += d.width
        while (width >= 8) {
            if (outptr == output.size) {
                return null
            }
            output[outptr++] = window.toByte()
            window = window ushr 8
            width -= 8
        }
    }
    if (width > 0) {
       output[outptr++] = window.toByte()
    }
    output.shrinkTo(outptr)
    return output
}

// Build the encoding dictionary.  The builder returns null if the bit strings would be too wide.

class DictItem {
    var bits = 0L
    var width = 0
}

fun buildEncodingDictionary(t: HuffTree) : Vector<DictItem>? {
    val dict = Vector<DictItem>(256) { DictItem() }
    fun build(t: HuffTree, bits: Long, width: Int): Boolean {
        if (t.zero == null) {
            dict[t.byte.toInt()].bits = bits
            dict[t.byte.toInt()].width = width
            return true
        }
        if (width == 56) {
            return false
        }
        build(t.zero!!, bits, width + 1)
        build(t.one!!, (1L shl width) or bits, width + 1)
        return true
    }
    if (!build(t, 0, 0)) {
        return null
    }
    return dict
}

// Build the Huffman tree.  The input array must be sorted by decreasing count, with lower byte values coming before
// higher byte values.

class HuffTree {
    var byte : Byte = 0
    var zero: HuffTree? = null
    var one: HuffTree? = null
    constructor(b:Byte) {
        byte = b
    }
    constructor(z:HuffTree, o:HuffTree) {
        zero=z
        one=o
    }
}

fun buildHuffmanTree(freqItems: Vector<FreqItem>) : HuffTree {
    val priq = Heap<HuffTree>()
    for ( i in 0 until freqItems.size) {
        val f = freqItems[i]
        priq.insert(-f.count, HuffTree(f.byte))
    }
    while (priq.size > 1) {
        val (a, wa) = priq.extractMax()
        val (b, wb) = priq.extractMax()
        priq.insert(wa + wb, HuffTree(a, b))
    }
    return priq.extractMax().t
}

// Compute byte frequencies and produce a sorted array for non-zero byte values.

class FreqItem(b: Byte, c: Int) {
    var byte = b
    var count = c
}

// Kotlin and Java suck (compared to Rust and Go) because there are no slices,
// hence there's going to be a lot of copying.

fun computeFrequencies(bytes: ByteVector): Vector<FreqItem> {
    val freqItems = Vector<FreqItem>(256) {FreqItem(it.toByte(), 0)}
    for ( i in 0 until bytes.size) {
        freqItems[bytes[i].toInt()].count++
    }
    // The sort is stable, so lower byte values are sorted before higher byte values at equal counts,
    // as required.
    freqItems.sortWith { a:FreqItem, b:FreqItem -> if (a.count != b.count) b.count - a.count else a.byte.toInt() - b.byte.toInt()}
    var numFreqItems = 256
    while (numFreqItems > 0 && freqItems[numFreqItems-1].count == 0) {
        numFreqItems--
    }
    freqItems.shrinkTo(numFreqItems)
    return freqItems
}
