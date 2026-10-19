package main

import (
	"encoding/json"
	_ "fmt"
	"log"
	"os"
)

type Parameter struct {
	In          string  `json:"in"`
	Name        string  `json:"name"`
	Schema      PSchema `json:"schema"`
	Description string  `json:"string"`
	Explode     bool    `json:"explode"`
}

type PSchema struct {
	Description string `json:"description"`
	Format      string `json:"format"`
	Type        string `json:"type"`
}

type OpenApi struct {
	Components struct {
		Schemas map[string]any `json:"schemas"`
	} `json:"components"`
	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	OpenApi string `json:"openapi"`
	Paths   map[string]struct {
		Get struct {
			OperationId string         `json:"operationId"`
			Parameters  []Parameter    `json:"parameters"`
			Responses   map[string]any `json:"responses"`
			Summary     string         `json:"summary"`
		} `json:"get"`
		Post struct {
			OperationId string         `json:"operationId"`
			Parameters  []Parameter    `json:"parameters"`
			RequestBody any            `json:"requestBody"`
			Required    bool           `json:"required"`
			Responses   map[string]any `json:"responses"`
			Summary     string         `json:"summary"`
		} `json:"post"`
	} `json:"paths"`
}

func getOpenApi(fn string) OpenApi {
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(f)
	var api OpenApi
	err = dec.Decode(&api)
	if err != nil {
		log.Fatal(err)
	}
	return api
}

/*
func main() {
	api := getOpenApi("f.json")
	fmt.Printf("%v\n", api.Info)
	for k, v := range api.Paths {
		fmt.Println(k)
		if v.Get.OperationId != "" {
			fmt.Println("  get")
			for _, p := range v.Get.Parameters {
				fmt.Printf("   %s %s %s %s\n", p.In, p.Name, p.Schema.Type, p.Schema.Format)
			}
		} else {
			fmt.Println("  post")
		}
	}
}
*/
