package main

import (
	"bufio"
	"cmp"
	"fmt"
	"math"
	"os"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: gpu_time input-file")
	}
	inp, err := os.Open(os.Args[1])
	if err != nil {
		panic("Can't open input file " + err.Error())
	}
	scanner := bufio.NewScanner(inp)
	type point struct {
		cmd string
		acct string
	}
	var grid = make(map[point]uint64)
	var cmds = make(map[string]bool)
	var accts = make(map[string]bool)
	if !scanner.Scan() {
		panic("No input")
	}
	ixs := make(map[string]int)
	for i, h := range fields(scanner.Text()) {
		ixs[h] = i
	}
	avgIx := ixs["avg"]
	acctIx := ixs["account"]
	cmdIx := ixs["job_name"]
	jobIdIx := ixs["job_id"]
	startIx := ixs["start_time"]
	endIx := ixs["end_time"]
	stateIx := ixs["job_state"]
	var errs int
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "--") {
			// nothing
		} else {
			fs := fields(l)
			if fs[jobIdIx] != "" {
				cmds[fs[cmdIx]] = true
				accts[fs[acctIx]] = true
				p := point{
					cmd: fs[cmdIx],
					acct: fs[acctIx],
				}
				t, err := computeTimeS(fs[avgIx], fs[startIx], fs[endIx], fs[stateIx])
				if err != nil {
					fmt.Println(err.Error())
					errs++
					if errs > 10 {
						return
					}
					continue
				}
				grid[p] += t
			}
		}
	}
	var sortedCmds = slices.Collect(maps.Keys(cmds))
	slices.Sort(sortedCmds)
	var sortedAccts = slices.Collect(maps.Keys(accts))
	slices.SortFunc(sortedAccts, func (a, b string) int {
		return cmp.Compare(atoi(a[2:]), atoi(b[2:]))
	})

	fmt.Print("Unit: GPU minutes,sum")
	for _, acct := range sortedAccts {
		fmt.Print(",")
		fmt.Print(acct)
	}
	fmt.Println()

	for _, cmd := range sortedCmds {
		fmt.Print(cmd)
		var sum uint64
		for _, acct := range sortedAccts {
			sum += grid[point{cmd, acct}]
		}
		fmt.Print(",", uint64(math.Round(float64(sum)/60)))
		for _, acct := range sortedAccts {
			fmt.Print(",")
			x := grid[point{cmd, acct}]
			if x > 0 {
				fmt.Print(uint64(math.Round(float64(x)/60)))
			}
		}
		fmt.Println()
	}
}

func computeTimeS(avg, start, end, state string) (t uint64, err error) {
	a, err := strconv.ParseFloat(avg, 64)
	if err != nil {
		return
	}
	s, err := time.Parse("2006-01-02 15:04:05-07", start)
	if err != nil {
		return
	}
	var e time.Time
	if end == "" {
		e = time.Now()
	} else {
		e, err = time.Parse("2006-01-02 15:04:05-07", end)
		if err != nil {
			return
		}
	}
	t = uint64(math.Round(a*float64(e.Unix()-s.Unix())))
	return
}

func fields(s string) []string {
	fs := strings.Split(s, "|")
	for i := range fs {
		fs[i] = strings.TrimSpace(fs[i])
	}
	return fs
}

func atoi(s string) int {
	n, _ := strconv.ParseInt(s, 10, 32)
	return int(n)
}
