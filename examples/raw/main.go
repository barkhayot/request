package main

import (
	"context"
	"fmt"
	"log"

	"github.com/barkhayot/request/pkg/request"
)

func main() {
	resp, err := request.RequestRaw(
		context.Background(),
		request.WithEndpoint("https://httpbin.org/status/204"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println("status:", resp.StatusCode)
}
