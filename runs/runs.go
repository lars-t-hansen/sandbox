// Usage: `runs [options]`
//   Reads stdin and writes stdout
//   Computes frequencies of overlapping two-byte runs in input and prints
//     them in decreasing frequency order.
//
// Options:
//   -r    Consider runs to be non-overlapping
//   -t n  Print only the top n.  Default=100.  Zero means 'all'
//   -b    Print the bottom n instead of the top n
//   -c m  Cutoff for consideration.  Default=0.
//
// TODO: Generalize this to n-byte runs for n up to at least eight.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

type stat struct {
	val   uint
	count uint
}

type stats []stat

func (ft stats) Len() int           { return len(ft) }
func (ft stats) Less(i, j int) bool { return ft[i].count > ft[j].count }
func (ft stats) Swap(i, j int)      { ft[i], ft[j] = ft[j], ft[i] }

func main() {
	var tFlag, cFlag uint
	var rFlag, bFlag bool
	flag.UintVar(&tFlag, "t", 100, "Number of entries to list")
	flag.UintVar(&cFlag, "c", 0, "Cutoff for useless values")
	flag.BoolVar(&rFlag, "r", false, "Consider entries not to overlap")
	flag.BoolVar(&bFlag, "b", false, "Print the bottom entries instead")
	flag.Parse()
	if tFlag == 0 {
		tFlag = 65536
	}
	buf := make([]uint8, 65536)
	ss := make(stats, 65536)
	for i := range ss {
		ss[i].val = uint(i)
	}
	for {
		bytesRead, err := os.Stdin.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		inc := 1
		if rFlag {
			inc = 2
		}
		for i := 0; i < bytesRead-1; i += inc {
			n := (uint(buf[i]) << 8) | uint(buf[i+1])
			ss[n].count++
		}
	}
	sort.Sort(ss)
	var numUseful int
	for numUseful = len(ss); numUseful > 0 && ss[numUseful-1].count <= cFlag; numUseful-- {
	}
	if tFlag > uint(numUseful) {
		tFlag = uint(numUseful)
	}
	toPrint := ss[:tFlag]
	if bFlag {
		toPrint = ss[uint(numUseful)-tFlag : numUseful]
	}
	for _, v := range toPrint {
		fmt.Fprintf(os.Stdout, "%04x\t%d\n", v.val, v.count)
	}
	overlap := ""
	if rFlag {
		overlap = "non-"
	}
	fmt.Fprintf(os.Stdout, "Number of useful %soverlapping entries: %d\n", overlap, numUseful)
}
