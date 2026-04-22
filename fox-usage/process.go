package main

import (
	"bufio"
	"fmt"
	"os"
	"maps"
	"slices"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: process input-file")
	}
	inp, err := os.Open(os.Args[1])
	if err != nil {
		panic("Can't open input file " + err.Error())
	}
	scanner := bufio.NewScanner(inp)
	var header []string
	var jobIdIx = -1
	var cmdIx = -1
	var acctIx = -1
	var cpuTimeIx = -1
	type point struct {
		cmd string
		acct string
	}
	var grid = make(map[point]uint64)
	var cmds = make(map[string]bool)
	var accts = make(map[string]bool)
	for scanner.Scan() {
		l := scanner.Text()
		if header == nil {
			header = fields(l)
			for i, h := range header {
				switch h {
				case "job_id":
					jobIdIx = i
				case "account":
					acctIx = i
				case "cmd":
					cmdIx = i
				case "cpu_time":
					cpuTimeIx = i
				}
			}
		} else if strings.HasPrefix(l, "--") {
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
				t, err := strconv.ParseUint(fs[cpuTimeIx], 10, 64)
				if err != nil {
					continue
				}
				grid[p] += t
			}
		}
	}
	var sortedCmds = slices.Collect(maps.Keys(cmds))
	slices.Sort(sortedCmds)
	var sortedAccts = slices.Collect(maps.Keys(accts))
	slices.Sort(sortedAccts)

	for _, acct := range sortedAccts {
		fmt.Print(",")
		fmt.Print(acct)
	}
	fmt.Println()

	for _, cmd := range sortedCmds {
		fmt.Print(cmd)
		for _, acct := range sortedAccts {
			fmt.Print(",")
			x := grid[point{cmd, acct}]
			if x > 0 {
				fmt.Print(x)
			}
		}
		fmt.Println()
	}
}

func fields(s string) []string {
	fs := strings.Split(s, "|")
	for i := range fs {
		fs[i] = strings.TrimSpace(fs[i])
	}
	return fs
}
