// Date fuzzer for anything with format:"date"
package main

import (
	"iter"
)

// This is pretty bare-bones but suggests how things might be done

func dateFuzzer() iter.Seq[FuzzValue] {
	return func(yield func(FuzzValue) bool) {
		if (!yield(FuzzValue{"", true})) {
			return
		}
		if (!yield(FuzzValue{"2006-10-12", true})) {
			return
		}
		yield(FuzzValue{"20061210", false})
	}
}

func init() {
	queryFuzzers["date"] = dateFuzzer()
}
