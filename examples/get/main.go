package main

import (
	"context"
	"fmt"
	"log"

	"github.com/barkhayot/request/pkg/request"
)

type Response struct {
	Args map[string]string `json:"args"`
}

func main() {
	ctx := context.Background()

	resp, err := request.Request[Response](
		ctx,
		request.WithEndpoint("https://httpbin.org/get"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("response: %+v\n", resp)
}
