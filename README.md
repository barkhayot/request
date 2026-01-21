# request

`request` is a small, opinionated HTTP client wrapper for Go, inspired by Python’s `requests` library.

It provides a simpler API on top of `net/http` while keeping Go’s explicitness and control.

## Features

- Context-aware requests
- Functional options (`WithX(...)`)
- Automatic JSON encoding/decoding
- Strong defaults (timeouts, headers)
- Fully testable (uses `httptest`)

## Installation

```bash
go get github.com/yourusername/request
```

## Usage

#### GET request
```go 
ctx := context.Background()

type Response struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

resp, err := request.Request[Response](
	ctx,
	request.WithEndpoint("https://api.example.com/resource"),
	request.WithMethod("GET"),
    request.WithTimeout(10*time.Second),
)
if err != nil {
	log.Fatal(err)
}
```

#### POST request with JSON body
```go
payload := map[string]string{
	"name": "example",
}

resp, err := request.Request[Response](
	ctx,
	request.WithEndpoint("https://api.example.com/resource"),
	request.WithMethod("POST"),
	request.WithBody(payload),
)
if err != nil {
    log.Fatal(err)
}
```

#### Headers and Query Parameters
```go
headers := http.Header{}
headers.Set("Authorization", "Bearer token")

params := url.Values{}
params.Set("page", "1")

resp, err := request.Request[Response](
	ctx,
	request.WithEndpoint("https://api.example.com/resource"),
	request.WithMethod("GET"),
	request.WithHeaders(headers),
	request.WithQueryParams(params),
)
if err != nil {
    log.Fatal(err)
}
```

#### Raw Requests
```go
resp, err := request.RequestRaw(
	ctx,
	request.WithEndpoint("https://api.example.com/resource"),
)
defer resp.Body.Close()

if err != nil {
    log.Fatal(err)
}
```

### Error Handling
- HTTP status codes ≥ 400 return an error
- Context cancellation is respected
- JSON decoding errors are returned directly

### Design Notes
-   The library avoids hidden magic and keeps behavior explicit
-   net/http primitives are preserved where possible
-   Internals are unexported to keep the public API small

### Examples

Runnable examples are available in the `examples/` directory:

```bash
go run ./examples/get
go run ./examples/post_json
go run ./examples/raw
```

## License
MIT License. See `LICENSE` file for details.
