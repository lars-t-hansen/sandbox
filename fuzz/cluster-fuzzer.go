// Fuzzer for anything with the name "cluster"

package main

import (
	"iter"
)

func clusterFuzzer() iter.Seq[FuzzValue] {
	return func(yield func(FuzzValue) bool) {
		yield(FuzzValue{"fox.educloud.no", true})
	}
}

func init() {
	paramFuzzers["cluster"] = clusterFuzzer()
}
