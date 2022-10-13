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
// or in complicated fallback schemes, more could be done.
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
		return err
	}
	defer inputFile.Close()
	outputFile, err := os.Create(fn + ".huff")
	if err != nil {
		return err
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
			return err
		}
		input := inputBlock[:bytesRead]
		freq := computeFrequencies(input, freqBlock)
		tree := buildHuffTree(freq)
		var encoded []uint8
		// TODO: abstract this away within populateEncDict
		for i := range dict {
			dict[i].width = 0
		}
		if populateEncDict(0, 0, tree, dict) {
			os.Stderr.WriteString(dict.String() + "\n")
			encoded = encodeBlock(dict, input, outputBlock)
		}
		if encoded != nil {
			metaloc := 0
			metaloc = put(metadata, metaloc, 2, uint(len(freq)))
			for _, item := range freq {
				metaloc = put(metadata, metaloc, 1, uint(item.val))
				metaloc = put(metadata, metaloc, 4, uint(item.count))
			}
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			metaloc = put(metadata, metaloc, 4, uint(len(encoded)))
			_, err := outputFile.Write(metadata[:metaloc])
			if err != nil {
				return err
			}
			_, err = outputFile.Write(encoded)
			if err != nil {
				return err
			}
		} else {
			metaloc := 0
			metaloc = put(metadata, metaloc, 2, 0)
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			_, err := outputFile.Write(metadata[:metaloc])
			if err != nil {
				return err
			}
			_, err = outputFile.Write(input)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Process the block and emit bits into the output block.  The last partial byte, if any,
// is filled with zeroes in the high bits.  If the output block fills up we return failure;
// the input should be stored uncompressed.
//
// Returns nil for overflow and output[:N] for N output bytes.

func encodeBlock(dict encDict, input []uint8, output []uint8) []uint8 {
	outptr := 0
	limit := len(output)
	bits := uint64(0)
	width := 0
	for _, b := range input {
		e := dict[b]
		bits = (bits << uint64(e.width)) | e.bits
		width += e.width
		for width >= 8 {
			if outptr == limit {
				return nil
			}
			output[outptr] = uint8(bits & 255)
			outptr++
			bits = bits >> 8
			width -= 8
		}
	}
	if width > 0 {
		if outptr == limit {
			return nil
		}
		output[outptr] = uint8(bits & 255)
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
	if tree.zero == nil {
		if width > 56 {
			return false
		}
		dict[tree.val].bits = bits
		dict[tree.val].width = width
		return true
	}
	return populateEncDict(width+1, bits<<1, tree.zero, dict) &&
		populateEncDict(width+1, (bits<<1)|1, tree.one, dict)
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Decoder
//
// We build a dictionary from the frequency table, the dictionary is a non-full binary tree with
// byte values at the leaves.  We then process the input and emit bytes into the output block.
// By construction, the block will have enough space.

// TODO

func decompress_file(fn string) error {
	if !strings.HasSuffix(fn, ".huff") {
		return encError("File to decompress must be named something.huff")
	}
	inputFile, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer inputFile.Close()
	outputFile, err := os.Create(fn[:len(fn)-5])
	if err != nil {
		return err
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
		// Read frequency table size
		bytesRead, err := inputFile.Read(metadata[0:2])
		if bytesRead == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		metaloc := 0
		freqCount, metaloc := get(metadata, metaloc, 2)
		if freqCount > 0 {
			// Read the frequency table, eof is not allowed
			bytesRead, err = inputFile.Read(metadata[metaloc : metaloc+int(freqCount)*5+8])
			// FIXME: must loop until we have enough bytes
			if err != nil {
				return err
			}
			freq := freqBlock[:int(freqCount)]
			for i := 0; i < int(freqCount); i++ {
				var v uint
				v, metaloc = get(metadata, metaloc, 1)
				freq[i].val = uint8(v)
				v, metaloc = get(metadata, metaloc, 4)
				freq[i].count = uint32(v)
			}
			tree := buildHuffTree(freq)
			bytesEncoded, metaloc := get(metadata, metaloc, 4)
			bytesInEncoding, metaloc := get(metadata, metaloc, 4)
			// Read the input block, eof is not allowed
			bytesRead, err = inputFile.Read(inputBlock)
			if err != nil {
				return err
			}

			decoded := decodeBlock(dict, inputBlock[:bytesInEncoding], outputBlock)
		} else {
			// Literal encoding

		}
		if decoded != nil {
			metaloc := 0
			metaloc = put(metadata, metaloc, 2, uint(len(freq)))
			for _, item := range freq {
				metaloc = put(metadata, metaloc, 1, uint(item.val))
				metaloc = put(metadata, metaloc, 4, uint(item.count))
			}
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			metaloc = put(metadata, metaloc, 4, uint(len(encoded)))
			_, err := outputFile.Write(metadata[:metaloc])
			if err != nil {
				return err
			}
			_, err = outputFile.Write(encoded)
			if err != nil {
				return err
			}
		} else {
			metaloc := 0
			metaloc = put(metadata, metaloc, 2, 0)
			metaloc = put(metadata, metaloc, 4, uint(bytesRead))
			_, err := outputFile.Write(metadata[:metaloc])
			if err != nil {
				return err
			}
			_, err = outputFile.Write(input)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	for nbytes > 0 {
		val = (val << 8) | uint(buf[ptr])
		ptr++
		nbytes--
	}
	newPtr = ptr
	return
}
