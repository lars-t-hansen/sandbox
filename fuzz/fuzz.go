package main

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"iter"
	"log"
	"net/http"
	"strings"
)

var api = getOpenApi("f.json")
var service = "http://localhost:1234"

func main() {
	reqd := map[string]string{
		"cluster":"fox.educloud.no",
	}
	for path, v := range api.Paths {
		if v.Get.OperationId == "" {
			continue
		}

		// For every `get` operation we want to test it.  If some field is fuzzable we want to run
		// it once for each fuzz value (including none, if optional).  This becomes hard: start-time
		// and start-date are kind of mutually exclusive.  So limit ourselves to "date" fields.

		var names []string
		var valueGens []iter.Seq[value]
		for _, param := range v.Get.Parameters {
			switch param.In {
			case "path":
				path = subst(path, param.Name, reqd[param.Name])
			case "query":
				if param.Schema.Format != "date" {
					continue
				}
				names = append(names, param.Name)
				valueGens = append(valueGens, dateFuzzer())
			case "header":
				// Do nothing
			}
		}
		if strings.HasSuffix(path, "/") {
			path = path[:len(path)-1]
		}
		withAll(valueGens, nil, 0, func (values []string, valid bool) {
			vs := strings.Join(zipem(names, values), "&")
			p := service + path
			if vs != "" {
				p = p + "?" + vs
			}
			//fmt.Println(p, valid)
			c := 200
			if !valid {
				c = 422  // Uprocessable Entity
			}
			_, r, err := doit(p, c)
			if err != nil {
				log.Printf("%s: failed: %v %d", p, err, r.StatusCode)
			} else {
				log.Printf("%s: SUCCESS", p)
			}
		})
	}

	// So here, for every API, look for something with format:"date", then invoke the
	// server with every date fuzzer value (including none) for each of the fields.
	/*
	j, _, err := doit("http://localhost:1234/api/v1/jobs/fox.educloud.no?start_date=20260101", 200)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T\n", j);
	if a, ok := j.([]any); ok {
		fmt.Printf("%v\n", a);
	} else {
		log.Fatalf("Bad form: %v", j)
	}
	*/
}

type sofar struct {
	x value
	next *sofar
}

func withAll(gs []iter.Seq[value], zapp *sofar, n int, do func([]string, bool)) {
	if len(gs) == 0 {
		vals, valid := values(zapp, n)
		do(vals, valid)
		return
	}
	for v := range gs[0] {
		withAll(gs[1:], &sofar{v, zapp}, n+1, do)
	}
}

func values(xs *sofar, n int) ([]string, bool) {
	vs := make([]string, n)
	valid := true
	for n > 0 {
		n--
		vs[n] = xs.x.v
		valid = valid && xs.x.valid
		xs = xs.next
	}
	return vs, valid
}

func zipem(as, bs []string) (result []string) {
	for i := range as {
		if bs[i] != "" {
			result = append(result, as[i] + "=" + bs[i])
		}
	}
	return
}

func subst(s, name, t string) string {
	return strings.Replace(s, "{" + name + "}", t, 1)
}


// The first value is a JSON object.  Object types are maps, array types are slices, not sure yet
// what numbers are, possibly json.Number.  The second is the original response.  The third is an
// error, or badCode or badJson.

var (
	badCode = errors.New("Bad status code")
	badJson = errors.New("Bad JSON")
)

func doit(url string, expect int) (any, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, r, err
	}
	if r.StatusCode != expect {
		return nil, r, errors.Join(badCode, errors.New(r.Status))
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	var x any
	err = dec.Decode(&x)
	if err != nil {
		return nil, r, errors.Join(badJson, err)
	}
	return x, r, nil
}

// Date fuzzer for anything with format:"date"

/*
var (
	goodDates = []string{"yyyy-mm-dd"}
	badDates  = []string{"yyyymmdd", "yyyy-mmdd", "yyyymm-dd", "yyy-mm-dd"}
	years     = rng{1980, 2100}
	month     = rng{1, 12}
	month
)

type FuzzDate struct {
}
*/

type value struct {
	v string
	valid bool
}

func dateFuzzer() iter.Seq[value] {
	return func(yield func(value) bool) {
		if (!yield(value{"",true})) {
			return
		}
		if (!yield(value{"2006-10-12",true})) {
			return
		}
		yield(value{"20061210",false})
	}
}

