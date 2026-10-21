package main

import (
	"flag"
	"iter"
	"log"
	"strings"
)

var api = getOpenApi("f.json")
var service = "http://localhost:1234"
var queryFuzzers = make(map[string]iter.Seq[FuzzValue])
var paramFuzzers = make(map[string]iter.Seq[FuzzValue])
var warn = flag.Bool("w", false, "Warn about unfuzzed formats")

type FuzzValue struct {
	V     string
	Valid bool
}

func main() {
	flag.Parse()

	// For every `get` operation we want to test it.  If some field is fuzzable we want to run
	// it once for each fuzz value (including none, if optional).  Known-invalid values will
	// lead to a known-invalid flag being set and an error code being expected.
	//
	// TODO: Here the error code is pinned to a specific value and that may be OK for what we're
	// testing, but in general the fuzz values might carry the expected error code and the allValid
	// flag should be replaced by a set of possible error codes.

	warned := make(map[string]bool)
	for path, v := range api.Paths {
		if v.Get.OperationId == "" {
			continue
		}

		var pathNames []string
		var pathValueGens []iter.Seq[FuzzValue]
		var queryNames []string
		var queryValueGens []iter.Seq[FuzzValue]
		for _, param := range v.Get.Parameters {
			switch param.In {
			case "path":
				g := paramFuzzers[param.Name]
				if g == nil {
					log.Fatalf("Path parameter fuzzer must exist for name \"%s\"", param.Name)
				}
				pathNames = append(pathNames, param.Name)
				pathValueGens = append(pathValueGens, g)
			case "query":
				// TODO: This is too primitive.  Consider cases where we don't have a format but may
				// have a type, and we may have min/max values or other constraints: minLength is
				// obvious.  So the map is not just queryFuzzers, it's queryFormatFuzzers; for other
				// cases there may be more.  And for the format, there may be additional
				// constraints, b/c int64 shows up as a format for the "integer" type but there can
				// be an additional minvalue constraint.
				if param.Schema.Format == "" {
					continue
				}
				g := queryFuzzers[param.Schema.Format]
				if g == nil {
					if *warn && !warned[param.Schema.Format] {
						warned[param.Schema.Format] = true
						log.Printf("Unfuzzed format \"%s\"", param.Schema.Format)
					}
					continue
				}
				queryNames = append(queryNames, param.Name)
				queryValueGens = append(queryValueGens, g)
			case "header":
				// Do nothing
			}
		}
		withAllValues(pathValueGens, func(pathValues []string, allValid bool) {
			newPath := path
			for i := range pathValues {
				newPath = subst(newPath, pathNames[i], pathValues[i])
			}
			url := cleanUrl(service + newPath)
			withAllValues(queryValueGens, func(queryValues []string, allValid bool) {
				// 422 = Unprocessable Entity
				_, r, err := requestAndResponse(url, queryNames, queryValues, allValid, 422)
				if err != nil {
					log.Printf("%s: failed: %v %d", url, err, r.StatusCode)
				} else {
					log.Printf("%s: SUCCESS", url)
				}
			})
		})
	}
}

func subst(s, name, t string) string {
	return strings.Replace(s, "{"+name+"}", t, 1)
}

func withAllValues(gs []iter.Seq[FuzzValue], apply func([]string, bool)) {
	withAll(gs, make([]string, len(gs)), 0, true, apply)
}

func withAll(gs []iter.Seq[FuzzValue], values []string, index int, allValid bool, apply func([]string, bool)) {
	if index == len(gs) {
		apply(values, allValid)
	} else {
		for v := range gs[index] {
			values[index] = v.V
			withAll(gs, values, index+1, allValid && v.Valid, apply)
		}
	}
}
