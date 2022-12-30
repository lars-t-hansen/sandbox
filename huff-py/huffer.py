import heapq
import io
import sys

# Usage: huffer input-name output-name
# FIXME: Improve that

def main():
    compress_file(sys.argv[1], sys.argv[2])

def compress_file(input_name, output_name):
    # Preallocate storage to avoid reallocating all of it needlessly
    input_buf = bytearray(65536)
    output_buf = bytearray(65536)
    meta_buf = bytearray(2+(5*256)+2*4)
    freq_buf = [FreqItem() for b in range(0,256)]
    dict_buf = [DictItem(0, 0) for b in range (0, 256)]

    infile = io.open(input_name, 'rb')
    outfile = io.open(output_name, 'wb')
    while True:
        # TODO: This is correct, though for ease of testing it would be best
        # always to fill the buffer if possible.
        bytes_read = infile.readinto(input_buf)
        if bytes_read == 0:
            break
        (output, output_len, meta, meta_len) = \
            compress_block(freq_buf,
                           dict_buf,
                           input_buf,
                           bytes_read,
                           output_buf,
                           meta_buf)
        # FIXME: This sucks, because it copies the buffer - pointlessly.
        # FIXME: It also sucks because it can write fewer bytes than we want to write.
        outfile.write(meta[:meta_len])
        outfile.write(output[:output_len])
    infile.close()
    outfile.close()

# Returns (output, output_len, meta, meta_len) always

def compress_block(freq_buf, dict_buf, input, input_len, output, meta):
    meta_loc = 0

    def put8(b):
        nonlocal meta_loc
        meta[meta_loc] = b & 255
        meta_loc += 1

    def put16(b):
        nonlocal meta_loc
        meta[meta_loc] = b & 255
        meta[meta_loc+1] = (b >> 8) & 255
        meta_loc += 2

    def put32(b):
        nonlocal meta_loc
        meta[meta_loc] = b & 255
        meta[meta_loc+1] = (b >> 8) & 255
        meta[meta_loc+2] = (b >> 16) & 255
        meta[meta_loc+3] = (b >> 24) & 255
        meta_loc += 4

    freq_len = compute_frequencies(input, input_len, freq_buf)
    tree = build_huffman_tree(freq_buf, freq_len)
    if build_dictionary(tree, dict_buf) is not None:
        output_len = encode_block(input, input_len, output, dict_buf)
        if output_len is not None:
            put16(freq_len)
            for i in range (0,freq_len):
                put8(freq_buf[i].byte)
                put32(freq_buf[i].count)
            put32(input_len)
            put32(output_len)
            return (output, output_len, meta, meta_loc)
    put16(0)
    put32(input_len)
    return (input, input_len, meta, meta_loc)

# Encode a block.
#
# input is a bytes or bytearray.  output is a bytearray.  dictionary is the encoding dictionary.
#
# Returns the length of the output block, or None if the encoding failed (output length exceeded)

def encode_block(input, inputlen, output, dictionary):
    inptr = 0
    outptr = 0
    bits = 0
    width = 0
    outputlen = len(output)
    while inptr < inputlen:
        dix = dictionary[input[inptr]]
        inptr += 1
        bits = bits | (dix.bits << width)
        width = width + dix.width
        while width >= 8:
            if outptr == outputlen:
                return None
            output[outptr] = bits & 255
            outptr += 1
            bits >>= 8
            width -= 8
    if width > 0:
        if outptr == outputlen:
            return None
        output[outptr] = bits & 255
        outptr += 1
    return outptr


# Build an encoding dictionary.  This is a table of length 265 where the entries are meaningful for
# the bytes that appear in the tree.  Each dictionary node is an object with `bits` and `width`
# fields.
#
# Input: the tree and a populated but otherwise garbage dictionary buffer (length 256)
#
# Output: the dictionary, or None if no encoding was found

class DictItem:
    def __init__(self, bits, width):
        self.bits = bits
        self.width = width

    def __str__(self):
        return f"({self.bits} {self.width})"

def build_dictionary(tree, dictionary):
    def descend(t, bits, width):
        if t.left is None:
            if width > 56:
                return False    # Can't encode
            dictionary[t.byte].bits = bits
            dictionary[t.byte].width = width
            return True
        return descend(t.left, bits, width + 1) and descend(t.right, (1 << width) | bits, width + 1)
    
    if not descend(tree, 0, 0):
        return None

    return dictionary

# Build a huffman tree.
#
# Input is a frequency table as produced by compute_frequencies.
#
# Returns a huffman-ordered binary tree.  `left` and `right` are None if this is a leaf node with
# the given byte value, otherwise left and right are both non-None and the byte value is immaterial.

class HuffNode:
    def __init__(self, byte, left, right):
        self.byte = byte
        self.left = left
        self.right = right

    def __str__(self):
        return f"({self.byte} {self.left} {self.right})"

def build_huffman_tree(freq, freq_len):

    class PqNode:
        def __init__(self, weight, serial, tree):
            self.weight = weight
            self.serial = serial
            self.tree = tree

        def __lt__(self, other):
            return (self.weight < other.weight or
                    (self.weight == other.weight and self.serial < other.serial))
            
    pq = []
    serial = 0
    for i in range (0,freq_len):
        it = freq[i]
        heapq.heappush(pq, PqNode(it.count, serial, HuffNode(it.byte, None, None)))
        serial += 1
    while len(pq) > 1:
        a = heapq.heappop(pq)
        b = heapq.heappop(pq)
        heapq.heappush(pq, PqNode(a.weight + b.weight, serial, HuffNode(0, a.tree, b.tree)))
        serial += 1
    return heapq.heappop(pq).tree

# Build table of character frequencies.
#
# Input is a bytearray or bytes and a buffer of length 256 of frequency items,
# all garbage.
#
# Initializes the table, sorts it descending per spec, and returns the number
# of nonzero entries (the effective length of the frequency table).

class FreqItem:
    def __init__(self):
        self.byte = 0
        self.count = 0

    def __lt__(self, other):
        return (self.count > other.count or
                (self.count == other.count and self.byte < other.byte))

    def __str__(self):
        return f"({self.byte} {self.count})"

def compute_frequencies(input, input_len, freq):
    for b in range (0, 256):
        it = freq[b]
        it.byte = b
        it.count = 0
    for i in range (0, input_len):
        freq[input[i]].count += 1
    freq.sort()
    i = 256
    while freq[i-1].count == 0:
        i -= 1
    return i

main()
