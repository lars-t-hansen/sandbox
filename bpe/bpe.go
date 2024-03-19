// Byte pair encoding.
//
// bpe [-a alphabet-size] input-file output-file

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var alphaSize = flag.Uint("a", 1024, "Alphabet has `num-codes` elements")

func main() {
	flag.Parse()
	// This is sort of wrong, the limit is no more than the number of distinct codes in the input.
	// At the same time, by having all bytes map to themselves it's possible to apply a pre-trained
	// encoding to new input.  So keep it the way it is.
	if *alphaSize < 256 {
		log.Fatal("Alphabet must have at least 256 entries")
	}
	// uint16
	if *alphaSize > 65536 {
		log.Fatal("Alphabet can have at most 65536 entries")
	}
	rest := flag.Args()
	if len(rest) != 2 {
		flag.Usage()
		os.Exit(2)
	}
	inputBytes, err := os.ReadFile(rest[0])
	if err != nil {
		log.Fatal(err)
	}

	lim := len(inputBytes)
	input := make([]uint16, lim)
	for i, c := range inputBytes {
		input[i] = uint16(c)
	}

	table := make(map[uint32]int)
	nextSym := 256
	remaining := *alphaSize - 256
	for remaining > 0 && lim > 1 {
		// For adjacent codes c1, c2 the key here is (c2 << 16) | c1, the value is the number of
		// pairs with that encoding.
		m := make(map[uint32]uint32)

		// I'm not sure whether we would most want to use overlapping or non-overlapping scans here
		// but I'm assuming overlapping is best.
		for i := 0 ; i < lim-1 ; i++ {
			k := uint32(input[i]) | (uint32(input[i+1]) << 16)
			m[k]++
		}

		var maxK, maxV uint32
		for k, v := range m {
			if v > maxV {
				maxK = k
				maxV = v
			}
		}

		// Let's not merge pairs that occur but once.  There's probably a more rigorous approach to
		// this cutoff.  Each dictionary entry takes (often) 6 bytes, so the cutoff may be a little
		// higher than "more than once", I'm guessing "more than twice".
		if maxV < 2 {
			break
		}

		sym := nextSym
		nextSym++
		remaining--
		table[maxK] = sym
		dst := 0
		i := 0
		for i < lim-1 {
			k := uint32(input[i]) | (uint32(input[i+1]) << 16)
			if k == maxK {
				input[dst] = uint16(sym)
				i += 2
			} else {
				input[dst] = input[i]
				i++
			}
			dst++
		}
		if i < lim {
			input[dst] = input[lim-1]
			dst++
		}
		lim = dst
	}

	// TODO: Skip the debug output, write to debug file instead
	// Really maybe we want [-o output-file] and if no argument we write stats or something
	// else on stdout?

	for k, v := range table {
		fmt.Printf("%d %d -> %d\n", int(k & 65535), int(k >> 16), v)
	}
	for i := 0 ; i < lim ; i++ {
		fmt.Printf("%d ", input[i])
	}
	fmt.Println()

	// This overstates the size of the output of course, we could use a leb or variable-length
	// encoding - we could huff it - and it will be much more compact.
	fmt.Printf("%d code units = %d bytes, originally %d\n", lim, 2*lim, len(inputBytes))
	fmt.Printf("%d dictionary entries = %d bytes\n", len(table), 2+(3*len(table)))
}
