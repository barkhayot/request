package proxy

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
	proxyURL := "http://example-proxy.com:8080"

	resp, err := request.Request[Response](
		ctx,
		request.WithEndpoint("https://httpbin.org/get"),
		request.WithProxy(proxyURL),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("response: %+v\n", resp)
}
