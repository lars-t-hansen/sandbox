// Usage: `runs [options]`
//   Reads stdin and writes stdout
//   Computes frequencies of two-byte runs in input and prints them in decreasing
//     frequency order.
//
// Options:
//   -t n  Print only the top n.  Default=100.  Zero means 'all'
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
	var tFlag uint
	flag.UintVar(&tFlag, "t", 100, "Number of entries to list")
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
		for i := 0; i < bytesRead-1; i++ {
			n := (uint(buf[i]) << 8) | uint(buf[i+1])
			ss[n].count++
		}
	}
	sort.Sort(ss)
	var numNonblank int
	for numNonblank = len(ss); numNonblank > 0 && ss[numNonblank-1].count == 0; numNonblank-- {
	}
	if tFlag > uint(numNonblank) {
		tFlag = uint(numNonblank)
	}
	for _, v := range ss[:tFlag] {
		fmt.Fprintf(os.Stdout, "%04x\t%d\n", v.val, v.count)
	}
	fmt.Fprintf(os.Stdout, "Number of nonzero entries: %d\n", numNonblank)
}
