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
go get github.com/barkhayot/request
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

#### QUERY request

`QUERY` is safe and idempotent like `GET`, but the request body carries the query
itself — useful for searches too large or too structured for the URL.

```go
resp, err := request.Request[Response](
	ctx,
	request.WithEndpoint("https://api.example.com/search"),
	request.WithMethod(request.MethodQuery),
	request.WithBody(map[string]any{"select": "name"}),
)
if err != nil {
    log.Fatal(err)
}
```

Methods are available as typed constants — `request.MethodGet`, `MethodPost`,
`MethodPut`, `MethodPatch`, `MethodDelete`, `MethodQuery`. Plain string literals
such as `request.WithMethod("GET")` still work.

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
-   For safe body-carrying methods (`QUERY`), `301`/`302` responses are returned
    to the caller rather than auto-followed, because `net/http` would rewrite
    them to a bodiless `GET` and silently change the request's meaning. `303`
    (which means "`GET` this other resource") and `307`/`308` are still followed

### Examples
Runnable examples are available in the `examples/` directory

## License
MIT License. See `LICENSE` file for details.
