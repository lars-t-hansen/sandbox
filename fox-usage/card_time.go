package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"maps"
	"slices"
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
	var cardTime = make(map[string]uint64)
	var jobs = make(map[string]int)
	ixs := findHeader(scanner)
	avgIx := ixs["avg"]
	jobIdIx := ixs["job_id"]
	startIx := ixs["start_time"]
	endIx := ixs["end_time"]
	manuIx := ixs["manufacturer"]
	modelIx := ixs["model"]
	memoryIx := ixs["memory"]
	for scanner.Scan() {
		fs := nextLine(scanner, ixs)
		if fs == nil {
			break
		}
		if fs[jobIdIx] != "" {
			card := fs[manuIx] + "/" + fs[modelIx] + "/" + fs[memoryIx]
			t, err := computeTimeS(fs[avgIx], fs[startIx], fs[endIx])
			if err != nil {
				fail(err.Error())
				continue
			}
			cardTime[card] += t
			jobs[card]++
		}
	}
	var sortedCards = slices.Collect(maps.Keys(cardTime))
	slices.Sort(sortedCards)

	fmt.Println("Card,GPU minutes,numjobs")

	for _, card := range sortedCards {
		fmt.Println(card, ",", uint64(math.Round(float64(cardTime[card])/60)), ",", jobs[card])
	}
}
