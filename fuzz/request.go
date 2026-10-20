package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var (
	badCode = errors.New("Bad status code")
	badJson = errors.New("Bad JSON")
)

// The first value is a JSON object.  Object types are maps, array types are slices, not sure yet
// what numbers are, possibly json.Number.  The second is the original response.  The third is an
// error, or something wrapping badCode or badJson.

func requestAndResponse(url string, names, values []string, allValid bool, ifInvalid int) (any, *http.Response, error) {
	vs := strings.Join(zipem(names, values), "&")
	if vs != "" {
		url = url + "?" + vs
	}
	//fmt.Println(url, valid)
	expect := 200
	if !allValid {
		expect = ifInvalid
	}
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

func cleanUrl(url string) string {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	return url
}

func zipem(as, bs []string) (result []string) {
	for i := range as {
		if bs[i] != "" {
			result = append(result, as[i]+"="+bs[i])
		}
	}
	return
}

/*
	// So here, for every API, look for something with format:"date", then invoke the
	// server with every date fuzzer value (including none) for each of the fields.
	j, _, err := requestAndResponse("http://localhost:1234/api/v1/jobs/fox.educloud.no?start_date=20260101", 200)
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
