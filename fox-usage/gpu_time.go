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
		job string
		acct string
	}
	var gpuTime = make(map[point]uint64)
	var counts = make(map[string]int)
	var jobs = make(map[string]bool)
	var accts = make(map[string]bool)
	ixs := findHeader(scanner)
	avgIx := ixs["avg"]
	acctIx := ixs["account"]
	jobIx := ixs["job_name"]
	jobIdIx := ixs["job_id"]
	startIx := ixs["start_time"]
	endIx := ixs["end_time"]
	stateIx := ixs["job_state"]
	var errs int
	for {
		fs := nextLine(scanner, ixs)
		if fs == nil {
			break
		}
		if fs[jobIdIx] != "" {
			job := mungeJobNameAggressive(fs[jobIx])
			jobs[job] = true
			accts[fs[acctIx]] = true
			p := point{
				job: job,
				acct: fs[acctIx],
			}
			t, err := computeTimeS(fs[avgIx], fs[startIx], fs[endIx], fs[stateIx])
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				errs++
				if errs > 10 {
					return
				}
				continue
			}
			gpuTime[p] += t
			counts[job]++
		}
	}
	var sortedJobs = slices.Collect(maps.Keys(jobs))
	slices.Sort(sortedJobs)
	var sortedAccts = slices.Collect(maps.Keys(accts))
	slices.SortFunc(sortedAccts, func (a, b string) int {
		return cmp.Compare(atoi(a[2:]), atoi(b[2:]))
	})

	fmt.Print("Unit: GPU minutes,count,sum")
	for _, acct := range sortedAccts {
		fmt.Print(",")
		fmt.Print(projects[acct])
	}
	fmt.Println()

	for _, job := range sortedJobs {
		fmt.Print(job)
		fmt.Print(",", counts[job])
		var sum uint64
		for _, acct := range sortedAccts {
			sum += gpuTime[point{job, acct}]
		}
		fmt.Print(",", uint64(math.Round(float64(sum)/60)))
		for _, acct := range sortedAccts {
			fmt.Print(",")
			x := gpuTime[point{job, acct}]
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
	// Scale by 100 b/c the utilization is in percentage points
	t = uint64(math.Round(a*float64(e.Unix()-s.Unix())/100))
	return
}
