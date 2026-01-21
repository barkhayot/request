package main

import (
	"context"
	"fmt"
	"log"

	"github.com/barkhayot/request/pkg/request"
)

type Response struct {
	JSON map[string]any `json:"json"`
}

func main() {
	payload := map[string]string{
		"name": "example",
	}

	resp, err := request.Request[Response](
		context.Background(),
		request.WithEndpoint("https://httpbin.org/post"),
		request.WithMethod("POST"),
		request.WithBody(payload),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("response: %+v\n", resp)
}
