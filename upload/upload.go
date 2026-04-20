package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type UploadOutput struct {
	Body string
}

func main() {
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("upload", "1"))

	huma.Post(
		api,
		"/upload",
		func(
			ctx context.Context,
			input *struct {
				Body struct {
					Zappa string `json:"zappa"`
					Zuppa string `json:"zuppa"`
				}
			},
		) (*UploadOutput, error) {
			payload, err := json.Marshal(input.Body)
			if err != nil {
				panic(err);
			}
			fmt.Println("Uploaded ", string(payload))
			return &UploadOutput{
				Body: "hi there " + input.Body.Zappa,
			}, nil
		},
	)
	http.ListenAndServe("127.0.0.1:5555", router)
}

