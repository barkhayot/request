package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/barkhayot/request/pkg/request"
	"github.com/barkhayot/request/pkg/throttler"
	"golang.org/x/time/rate"
)

type Response struct {
	Args map[string]string `json:"args"`
}

func main() {
	// Create a context for the request lifecycle.
	ctx := context.Background()

	// Configure a throttler. We use golang.org/x/time/rate to
	// control requests per second and burst size.
	//
	// Settings guidance:
	// - rate.Limit: average requests per second allowed.
	// - burst: maximum number of requests allowed to burst at once.
	//
	// Example: rate.Every(200*time.Millisecond) == 5 requests/sec.
	// A burst of 2 allows short spikes above the steady rate.
	r := rate.Every(200 * time.Millisecond)
	t := throttler.NewRateThrottler(r, 2)

	// Build a request to httpbin.org (or replace with your endpoint).
	// Response will be decoded into Response struct.
	var resp Response

	// Perform the request using the package helper. We pass the
	// configured throttler so that the request will wait when needed.
	resp, err := request.Request[Response](ctx,
		request.WithEndpoint("https://httpbin.org/get"),
		request.WithMethod("GET"),
		request.WithThrottler(t),
		// Optionally set a timeout shorter than the default:
		request.WithTimeout(3*time.Second),
	)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	fmt.Printf("response args: %+v\n", resp.Args)
}
