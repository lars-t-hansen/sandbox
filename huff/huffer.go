// Huffman compressor / decompressor
//
// huffer compress [-o outfile] filename
// huff [-o outfile] filename
//   Creates outfile, or if no -o option, filename.huff
//
// huffer decompress [-o outfile] filename.huff
// puff [-o outfile] filename.huff
//   Creates outfile, or if no -o option, filename
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
// An uncompressed block can be written under some circumstances, it is represented as
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

type huffError string

func (e huffError) Error() string {
	return string(e)
}

var usage string = "Usage: huffer [compress|decompress] [-o outfilename] infilename"

func main() {
	var err error
	var isCompress, isDecompress bool
	var inFilename, outFilename string

	args := os.Args

	{
		components := strings.Split(args[0], "/")
		progname := components[len(components)-1]
		if progname == "huff" {
			isCompress = true
		} else if progname == "puff" {
			isDecompress = true
		}
		args = args[1:]
	}

	if !isCompress && !isDecompress {
		if args[0] == "compress" {
			isCompress = true
			args = args[1:]
		} else if args[0] == "decompress" {
			isDecompress = true
			args = args[1:]
		} else {
			err = huffError(usage)
		}
	}

	if err == nil {
		if len(args) == 3 {
			if args[0] == "-o" {
				outFilename = args[1]
				inFilename = args[2]
			} else {
				err = huffError(usage)
			}
		} else if len(args) == 1 {
			inFilename = args[0]
			if isCompress {
				outFilename = inFilename + ".huff"
			} else {
				if !strings.HasSuffix(inFilename, ".huff") {
					err = huffError("File to decompress must be named something.huff")
				} else {
					outFilename = inFilename[:len(inFilename)-5]
				}
			}
		}
	}

	if err == nil {
		if isCompress {
			err = compressFile(inFilename, outFilename)
		} else {
			err = decompressFile(inFilename, outFilename)
		}
	}

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Compressor

func compressFile(inFilename, outFilename string) error {
	inputFile, err := os.Open(inFilename)
	if err != nil {
		return huffError("Opening " + inFilename + " for reading: " + err.Error())
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outFilename)
	if err != nil {
		return huffError("Opening " + outFilename + " for writing: " + err.Error())
	}
	defer outputFile.Close()
	return compressStream(inputFile, outputFile, inFilename, outFilename)
}

func compressStream(inputFile, outputFile *os.File, inputName, outputName string) error {
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
			return huffError("Reading input from " + inputName + ": " + err.Error())
		}
		input := inputBlock[:bytesRead]
		freq := computeFrequencies(input, freqBlock)
		tree := buildHuffTree(freq)
		var encoded []uint8
		if populateEncDict(0, 0, tree, dict) {
			encoded = compressBlock(dict, input, outputBlock)
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
			return huffError("Writing metadata to " + outputName + ": " + err.Error())
		}
		_, err = outputFile.Write(encoded)
		if err != nil {
			return huffError("Writing output to " + outputName + ": " + err.Error())
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

func compressBlock(dict encDict, input []uint8, output []uint8) []uint8 {
	outptr := 0
	limit := len(output)
	window := uint64(0)
	width := 0
	for _, b := range input {
		e := dict[b]
		window = window | (e.bits << width)
		width += e.width
		for width >= 8 {
			if outptr == limit {
				return nil
			}
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

// The bit string in an item is encoded with bits higher in the tree toward the
// least significant bits, because that is how the decoder wants to use them:
// it masks off the low bit to branch left or right, then shifts in the higher bits.

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
	return doPopulateEncDict(width+1, bits, tree.zero, dict) &&
		doPopulateEncDict(width+1, (1<<width)|bits, tree.one, dict)
}

/////////////////////////////////////////////////////////////////////////////////////
//
// Decompressor

func decompressFile(inFilename, outFilename string) error {
	inputFile, err := os.Open(inFilename)
	if err != nil {
		return huffError("Opening " + inFilename + " for reading: " + err.Error())
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outFilename)
	if err != nil {
		return huffError("Opening " + outFilename + " for writing: " + err.Error())
	}
	defer outputFile.Close()
	return decompressStream(inputFile, outputFile, inFilename, outFilename)
}

func decompressStream(inputFile, outputFile *os.File, inputName, outputName string) error {
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
			return huffError("Reading metadata from " + inputName + ": " + err.Error())
		}
		if bytesRead < 2 {
			return huffError("Reading metadata from " + inputName + ": premature EOF")
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
			return huffError("Reading metadata from " + inputName + ": premature EOF")
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
			return huffError("Reading data from " + inputName + ": premature EOF")
		}
		var decoded []uint8
		if freqCount > 0 {
			tree := buildHuffTree(freq)
			decoded = decompressBlock(tree, bytesEncoded, input, outputBlock)
		} else {
			decoded = input
		}
		_, err = outputFile.Write(decoded)
		if err != nil {
			return huffError("Writing data to " + outputName + ": " + err.Error())
		}
	}
	return nil
}

func decompressBlock(tree *huffTree, bytesEncoded uint, input []uint8, output []uint8) []uint8 {
	outPtr := 0
	inPtr := 0
	inbyte := uint8(0)
	inwidth := 0
	t := tree
	for {
		// If we get to a leaf, emit the leaf.  If we've emitted as many as we should, exit.
		if t.zero == nil {
			output[outPtr] = t.val
			outPtr++
			if uint(outPtr) == bytesEncoded {
				break
			}
			t = tree
			continue
		}
		// Backfill input if we've run out.  We can't run out of input here, we should have
		// exited above.
		if inwidth == 0 {
			inbyte = input[inPtr]
			inwidth = 8
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
// Create tree representing the Huffman encoding according to the frequency table.

// The branches are either both nil or both not nil.  If not nil then this is an interior
// node and val is invalid, otherwise it's a leaf.

type huffTree struct {
	zero, one *huffTree
	val       uint8
}

// Build a tree from a frequency table sorted in descending order by frequency, for
// non-zero frequencies only.

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

// Return a table of (byteValue, frequency) sorted in descending order by frequency,
// for non-zero frequencies.

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
// Buffer utilities

// Encode `val` of size `nbytes` little-endian into `buf` at `ptr` and return
// `ptr+nbytes`.

func put(buf []uint8, ptr int, nbytes int, val uint) int {
	for nbytes > 0 {
		buf[ptr] = uint8(val & 255)
		val >>= 8
		ptr++
		nbytes--
	}
	return ptr
}

// Decode `val` of size `nbytes` little-endian from `buf` at `ptr` and return
// `val` and `ptr+nbytes`.

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
