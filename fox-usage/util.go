package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FoxProject struct {
	ProjNum string `json:"project_number"`
	ProjNam string `json:"project_name"`
	ProjLongNam string `json:"project_long_name"`
}

var projects = make(map[string]string)

func init() {
	bytes, err := os.ReadFile("fox-projects.json")
	if err != nil {
		panic(err)
	}
	type envelope struct {
		CpuProjects []FoxProject `json:"cpu_projects"`
	}
	var e envelope
	err = json.Unmarshal(bytes, &e)
	if err != nil {
		panic(err)
	}
	for _, p := range e.CpuProjects {
		projects[p.ProjNum] = p.ProjNam
	}
	// for k, v := range projects {
	// 	fmt.Fprintf(os.Stderr, "%s => %s\n", k, v)
	// }
}

func findHeader(scanner *bufio.Scanner) map[string]int {
	for {
		if !scanner.Scan() {
			panic("No input")
		}
		l := scanner.Text()
		fs := fields(l)
		// heuristic
		if len(fs) > 2 {
			ixs := make(map[string]int)
			for i, h := range fs {
				ixs[h] = i
			}
			return ixs
		}
		fmt.Fprintln(os.Stderr, "Skipping: " + l)
	}
}

func nextLine(scanner *bufio.Scanner, ixs map[string]int) []string {
	for {
		if !scanner.Scan() {
			return nil
		}
		l := scanner.Text()
		fs := fields(l)
		if len(fs) != len(ixs) {
			if !strings.HasPrefix(l, "---") {
				fmt.Fprintln(os.Stderr, "Skipping: " + l)
			}
			continue
		}
		return fs
	}
}

func mungeJobName(s string) string {
	s, _, _ = strings.Cut(s, "=")
	if len(s) > 20 {
		s = s[:20] + "*"
	}
	return escape(s)
}

func mungeJobNameAggressive(s string) string {
	s = mungeJobName(s)
	s, _, a := strings.Cut(s, "-")
	s, _, b := strings.Cut(s, "_")
	if a || b {
		s = s + "*"
	}
	return s
}

func mungeCmd(s string) string {
	if strings.HasPrefix(s, "python") {
		return "python"
	}
	if strings.HasPrefix(s, "ipython") {
		return "python"
	}
	if matched, _ := regexp.MatchString("^model_[a-zA-Z0-9]+$", s); matched {
		return "model_XXX"
	}
	return escape(strings.Replace(s, " <defunct>", "", -1))
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

func escape(s string) string {
	return strings.Replace(strings.Replace(s, ",", "_", -1), "\"", "_", -1)
}

var errs int

func fail(s string) {
	fmt.Fprintf(os.Stderr, "ERROR: " + s)
	errs++
	if errs > 10 {
		panic("Too many errors")
	}
}

// Given an average utilization in percentage points and a time span (all strings), compute the
// total time.
func computeTimeS(avg, start, end string) (t uint64, err error) {
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
