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
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: cpu_time input-file")
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
	var cpuTime = make(map[point]uint64)
	var jobNames = make(map[string]map[string]bool)
	var cmds = make(map[string]bool)
	var accts = make(map[string]bool)
	ixs := findHeader(scanner)
	jobIdIx := ixs["job_id"]
	jobNameIx := ixs["job_name"]
	cmdIx := ixs["cmd"]
	acctIx := ixs["account"]
	cpuTimeIx := ixs["cpuTime"]
	for {
		fs := nextLine(scanner, ixs)
		if fs == nil {
			break
		}
		if fs[jobIdIx] != "" {
			cmd := mungeCmd(fs[cmdIx])
			cmds[cmd] = true
			accts[fs[acctIx]] = true
			p := point{
				cmd: cmd,
				acct: fs[acctIx],
			}
			t, err := strconv.ParseUint(fs[cpuTimeIx], 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Bad: %s\n", fs[cpuTimeIx])
				continue
			}
			cpuTime[p] += t
			jobName := mungeJobName(fs[jobNameIx])
			if jobNames[cmd] == nil {
				jobNames[cmd] = make(map[string]bool)
			}
			jobNames[cmd][jobName] = true
		}
	}
	var sortedCmds = slices.Collect(maps.Keys(cmds))
	slices.Sort(sortedCmds)
	var sortedAccts = slices.Collect(maps.Keys(accts))
	slices.SortFunc(sortedAccts, func (a, b string) int {
		return cmp.Compare(atoi(a[2:]), atoi(b[2:]))
	})

	fmt.Print("Unit: CPU minutes,sum,jobs")
	for _, acct := range sortedAccts {
		fmt.Print(",")
		fmt.Print(projects[acct])
	}
	fmt.Println()

	for _, cmd := range sortedCmds {
		fmt.Print(cmd)
		var sum uint64
		for _, acct := range sortedAccts {
			sum += cpuTime[point{cmd, acct}]
		}
		fmt.Print(",", uint64(math.Round(float64(sum)/60)))
		names := slices.Collect(maps.Keys(jobNames[cmd]))
		if len(names) > 2 {
			fmt.Printf(",%s ...", strings.Join(names[:2], " "))
		} else {
			fmt.Printf(",%s", strings.Join(names, " "))
		}
		for _, acct := range sortedAccts {
			fmt.Print(",")
			x := cpuTime[point{cmd, acct}]
			if x > 0 {
				fmt.Print(uint64(math.Round(float64(x)/60)))
			}
		}
		fmt.Println()
	}
}
