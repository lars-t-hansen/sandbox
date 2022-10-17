// Huffman encoder / decoder
//
// huff encode filename
//   Creates filename.huff
//
// huff decode filename.huff
//   Creates filename
//
// Since this is a programming exercise, it works by reading 64KB blocks and
// compressing them individually; the output file consists of compressed blocks.
// Also, we don't care about micro-efficiencies in representing the dictionary
// in the file or in complicated fallback schemes, more could be done.
//
// A compressed block is represented as
//   number of dictionary entries: u16 > 0 (max value is really 256)
//   run of dictionary entries sorted descending by frequency:
//     value: u8
//     frequency: u32 (max value is really 65536)
//   number of encoded bytes: u32 (max value is really 65536)
//   number of bytes used for encoded bytes: u32 (max value 65536)
//   bytes, the number of which is encoded by previous field
//
// An uncompressed block is written under some circumstances, it is represented as
//   0: u16
//   number of bytes: u32 (really max 65536)
//   bytes, the number of which is encoded by previous field

package main

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	args := os.Args[1:]
	var err error
	if args[0] == "encode" && len(args) == 2 {
		err = compress_file(args[1])
	} else if args[0] == "decode" && len(args) == 2 {
		err = decompress_file(args[1])
	} else {
		panic(fmt.Sprintf("Bad command %v", args))
	}
	if err != nil {
		panic(fmt.Sprintf("Failed: %v", err))
	}
}

type encError string

func (e encError) Error() string {
	return string(e)
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Encoder

func compress_file(fn string) error {
	inputFile, err := os.Open(fn)
	if err != nil {
		return encError("Opening " + fn + " for reading: " + err.Error())
	}
	defer inputFile.Close()

	outFn := fn + ".huff"
	outputFile, err := os.Create(outFn)
	if err != nil {
		return encError("Opening " + outFn + " for writing: " + err.Error())
	}
	defer outputFile.Close()

	inputBlock := make([]uint8, 65536)
	outputBlock := make([]uint8, 65536)
	freqBlock := make([]freqEntry, 256)
	dict := make(encDict, 256)
	metasize := 2 /* freq table size */ +
		256*5 /* freq table max size */ +
		4 /* number of input bytes encoded */ +
		4 /* number of bytes in encoding */
	metadata := make([]uint8, metasize)
	for {
		bytesRead, err := inputFile.Read(inputBlock)
		if bytesRead == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return encError("Reading input from " + fn + ": " + err.Error())
		}
		input := inputBlock[:bytesRead]
		freq := computeFrequencies(input, freqBlock)
		tree := buildHuffTree(freq)
		var encoded []uint8
		if populateEncDict(0, 0, tree, dict) {
			os.Stderr.WriteString(dict.String() + "\n")
			encoded = encodeBlock(dict, input, outputBlock)
		}
		metaloc := 0
		if encoded != nil {
			metaloc = put(metadata, metaloc, 2, uint(len(freq)))
			for _, item := range freq {
				metaloc = put(metadata, metaloc, 1, uint(item.val))
				metaloc = put(metadata, metaloc, 4, uint(item.count))
			}
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			metaloc = put(metadata, metaloc, 4, uint(len(encoded)))
		} else {
			metaloc = put(metadata, metaloc, 2, 0)
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			encoded = input
		}
		_, err = outputFile.Write(metadata[:metaloc])
		if err != nil {
			return encError("Writing metadata to " + outFn + ": " + err.Error())
		}
		_, err = outputFile.Write(encoded)
		if err != nil {
			return encError("Writing output to " + outFn + ": " + err.Error())
		}
	}
	return nil
}

// Process the block and emit bits into the output block.  The bits are output by inserting them
// into a sliding window above the bits previously output and then writing eight bits at a time
// to the output.  There are always zeroes in the window so the the last partial byte, if any,
// is filled with zeroes in the high bits.  If the output block fills up we return failure;
// the input should be stored uncompressed.
//
// Returns nil for overflow and output[:N] for N output bytes.

func encodeBlock(dict encDict, input []uint8, output []uint8) []uint8 {
	outptr := 0
	limit := len(output)
	window := uint64(0)
	width := 0
	for _, b := range input {
		e := dict[b]
		// FIXME: This is a hack.  The reversed bits should be stored in the
		// dictionary that way and the dictionary should print it both ways,
		// tagged appropriately.
		//
		// FIXME: Update comments here and there to reflect that.
		bs := uint64(0)
		x := e.bits
		for i := 0; i < e.width; i++ {
			bs = (bs << 1) | (x & 1)
			x >>= 1
		}
		window = window | (bs << width)
		width += e.width
		for width >= 8 {
			if outptr == limit {
				return nil
			}
			os.Stderr.WriteString(fmt.Sprintf("%b ", window&255))
			output[outptr] = uint8(window & 255)
			outptr++
			window >>= 8
			width -= 8
		}
	}
	if width > 0 {
		if outptr == limit {
			return nil
		}
		os.Stderr.WriteString(fmt.Sprintf("%b ", window&255))
		output[outptr] = uint8(window & 255)
		outptr++
	}
	return output[:outptr]
}

// The encoding dictionary is an array mapping byte values to bit strings; only the
// entries representing values that have been found to be in the input are represented
// in the dictionary.  The dictionary always has length 256 though.

type encDict []encDictItem

func (d encDict) String() string {
	s := ""
	for i, e := range d {
		if e.width > 0 {
			bits := fmt.Sprintf("%b", e.bits+(1<<56))[57-e.width : 57]
			s = s + fmt.Sprintf("('%s' %s) ", string(rune(i)), bits)
		}
	}
	return s
}

type encDictItem struct {
	width int
	bits  uint64
}

func populateEncDict(width int, bits uint64, tree *huffTree, dict encDict) bool {
	for i := range dict {
		dict[i].width = 0
	}
	return doPopulateEncDict(width, bits, tree, dict)
}

func doPopulateEncDict(width int, bits uint64, tree *huffTree, dict encDict) bool {
	if tree.zero == nil {
		if width > 56 {
			return false
		}
		dict[tree.val].bits = bits
		dict[tree.val].width = width
		return true
	}
	return doPopulateEncDict(width+1, bits<<1, tree.zero, dict) &&
		doPopulateEncDict(width+1, (bits<<1)|1, tree.one, dict)
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Decoder

func decompress_file(fn string) error {
	if !strings.HasSuffix(fn, ".huff") {
		return encError("File to decompress must be named something.huff")
	}
	inputFile, err := os.Open(fn)
	if err != nil {
		return encError("Opening " + fn + " for reading: " + err.Error())
	}
	defer inputFile.Close()

	outFn := fn[:len(fn)-5]
	outputFile, err := os.Create(outFn)
	if err != nil {
		return encError("Opening " + outFn + " for writing: " + err.Error())
	}
	defer outputFile.Close()

	inputBlock := make([]uint8, 65536)
	outputBlock := make([]uint8, 65536)
	freqBlock := make([]freqEntry, 256)
	metasize := 2 /* freq table size */ +
		256*5 /* freq table max size */ +
		4 /* number of input bytes encoded */ +
		4 /* number of bytes in encoding */
	metadata := make([]uint8, metasize)
	for {
		bytesRead, err := inputFile.Read(metadata[0:2])
		if bytesRead == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return encError("Reading metadata from " + fn + ": " + err.Error())
		}
		if bytesRead < 2 {
			return encError("Reading metadata from " + fn + ": premature EOF")
		}
		metaloc := 0
		freqCount := uint(0)
		freqCount, metaloc = get(metadata, metaloc, 2)
		numMetaBytes := 0
		if freqCount > 0 {
			numMetaBytes = int(freqCount)*5 + 4 + 4
		} else {
			numMetaBytes = 4
		}
		bytesRead, err = inputFile.Read(metadata[metaloc : metaloc+numMetaBytes])
		if err != nil {
			return err
		}
		if bytesRead < numMetaBytes {
			return encError("Reading metadata from " + fn + ": premature EOF")
		}
		var bytesEncoded, bytesInEncoding uint
		var freq []freqEntry
		if freqCount > 0 {
			freq = freqBlock[:int(freqCount)]
			for i := 0; i < int(freqCount); i++ {
				var v uint
				v, metaloc = get(metadata, metaloc, 1)
				freq[i].val = uint8(v)
				v, metaloc = get(metadata, metaloc, 4)
				freq[i].count = uint32(v)
			}
			bytesEncoded, metaloc = get(metadata, metaloc, 4)
			bytesInEncoding, metaloc = get(metadata, metaloc, 4)
		} else {
			bytesEncoded, metaloc = get(metadata, metaloc, 4)
			bytesInEncoding = bytesEncoded
		}
		input := inputBlock[:bytesInEncoding]
		bytesRead, err = inputFile.Read(input)
		if err != nil {
			return err
		}
		if bytesRead < len(input) {
			return encError("Reading data from " + fn + ": premature EOF")
		}
		var decoded []uint8
		if freqCount > 0 {
			tree := buildHuffTree(freq)
			decoded = decodeBlock(tree, bytesEncoded, input, outputBlock)
		} else {
			decoded = input
		}
		_, err = outputFile.Write(decoded)
		if err != nil {
			return encError("Writing data to " + outFn + ": " + err.Error())
		}
	}
	return nil
}

func decodeBlock(tree *huffTree, bytesEncoded uint, input []uint8, output []uint8) []uint8 {
	outPtr := 0
	inPtr := 0
	inbyte := uint8(0)
	inwidth := 0
	t := tree
	for {
		// If we get to a leaf, emit the leaf
		if t.zero == nil {
			output[outPtr] = t.val
			outPtr++
			if uint(outPtr) == bytesEncoded {
				break
			}
			t = tree
			continue
		}
		// We need a bit, but if there isn't one then get one.  If there still isn't one
		// then we're done.
		if inwidth == 0 {
			if inPtr == len(input) {
				// TODO: It's probably an error here if t != tree
				break
			}
			inbyte = input[inPtr]
			inPtr++
		}
		bit := inbyte & 1
		inbyte >>= 1
		inwidth--
		if bit == 0 {
			t = t.zero
		} else {
			t = t.one
		}
	}
	return output[:outPtr]
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Create tree representing the huffman encoding according to the frequency table.

// The branches are either both nil or both not nil.  If not nil then this is an interior
// node and val is invalid, otherwise it's a leaf.

type huffTree struct {
	zero, one *huffTree
	val       uint8
}

func buildHuffTree(ft freqTable) *huffTree {
	h := newHuffHeap(ft)
	for h.Len() > 1 {
		a := heap.Pop(&h).(huffItem)
		b := heap.Pop(&h).(huffItem)
		heap.Push(&h, huffItem{
			weight: a.weight + b.weight,
			tree:   &huffTree{zero: a.tree, one: b.tree},
		})
	}
	return heap.Pop(&h).(huffItem).tree
}

// Heap of tree nodes, a priority queue used during tree building.

type huffItem struct {
	weight uint32
	tree   *huffTree
}

type huffHeap []huffItem

func newHuffHeap(ft freqTable) huffHeap {
	h := make(huffHeap, len(ft))
	for i, v := range ft {
		h[i] = huffItem{
			weight: v.count,
			tree:   &huffTree{val: v.val},
		}
	}
	heap.Init(&h)
	return h
}

func (h huffHeap) Len() int           { return len(h) }
func (h huffHeap) Less(i, j int) bool { return h[i].weight < h[j].weight }
func (h huffHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (ft *huffHeap) Push(x any) {
	*ft = append(*ft, x.(huffItem))
}

func (ft *huffHeap) Pop() any {
	old := *ft
	n := len(old)
	x := old[n-1]
	*ft = old[0 : n-1]
	return x
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Compute byte frequencies

type freqEntry struct {
	val   uint8
	count uint32
}

type freqTable []freqEntry

func (ft freqTable) Len() int           { return len(ft) }
func (ft freqTable) Less(i, j int) bool { return ft[i].count > ft[j].count }
func (ft freqTable) Swap(i, j int)      { ft[i], ft[j] = ft[j], ft[i] }

func computeFrequencies(input []uint8, ft freqTable) freqTable {
	for i := range ft {
		ft[i].val = uint8(i)
		ft[i].count = 0
	}
	for _, b := range input {
		ft[b].count++
	}
	sort.Sort(ft)
	i := 0
	for i < len(ft) && ft[i].count > 0 {
		i++
	}
	return ft[:i]
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Random utilities

func put(buf []uint8, ptr int, nbytes int, val uint) int {
	for nbytes > 0 {
		buf[ptr] = uint8(val & 255)
		val >>= 8
		ptr++
		nbytes--
	}
	return ptr
}

func get(buf []uint8, ptr int, nbytes int) (val uint, newPtr int) {
	shift := 0
	for nbytes > 0 {
		val = val | (uint(buf[ptr]) << shift)
		shift += 8
		ptr++
		nbytes--
	}
	newPtr = ptr
	return
}
