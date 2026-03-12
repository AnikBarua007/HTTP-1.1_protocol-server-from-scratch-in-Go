# HTTP/1.1 from Scratch in Go

This project is a small HTTP/1.1 learning implementation written in Go. It focuses on the low-level parts of web servers: TCP listeners, HTTP request parsing, header parsing, response writing, and graceful shutdown.

The codebase currently includes:

- a concurrent TCP-based HTTP server
- a custom HTTP/1.1 request parser
- a custom header parser
- helpers for writing HTTP status lines and headers
- tests for request-line, header, and body parsing behavior
- small command-line programs used while learning sockets and protocols

## Goals

The main goal of this repository is to understand how HTTP works under the hood without relying on Go's built-in `net/http` server stack.

Implemented topics include:

- listening for TCP connections manually
- accepting and closing connections safely
- parsing HTTP request lines
- parsing HTTP headers incrementally from streamed input
- reading request bodies based on `Content-Length`
- writing valid HTTP/1.1 response status lines and headers
- returning a simple text response to clients
- shutting the server down gracefully on `SIGINT` and `SIGTERM`

## What has been implemented

### 1. HTTP server bootstrap

The main entry point for the HTTP server is [cmd/httpserver/main.go](cmd/httpserver/main.go).

Implemented behavior:

- starts the server on port `42069`
- logs startup status
- waits for OS shutdown signals
- closes the listener cleanly before exit

### 2. Concurrent TCP server

The reusable server implementation lives in [internal/server/servers.go](internal/server/servers.go).

Implemented behavior:

- opens a TCP listener with `net.Listen`
- accepts connections in a loop
- handles each client in its own goroutine
- tracks shutdown state with an atomic flag
- stops accepting new connections after `Close()`
- waits for the listener loop to finish before returning from shutdown

Current response behavior:

- always returns `HTTP/1.1 200 OK`
- always sends a plain-text body
- response body is:

	`Hello World!`

- default response headers include:
	- `Content-Length`
	- `Connection: close`
	- `Content-Type: text/plain`

### 3. HTTP request parsing

The request parser lives in [internal/request/request.go](internal/request/request.go).

Implemented behavior:

- parses requests from any `io.Reader`
- supports incremental parsing from small chunks of bytes
- parses the request line into:
	- method
	- request target
	- HTTP version
- validates that:
	- the request line contains exactly 3 parts
	- the version starts with `HTTP/`
	- the method is uppercase
	- the request target starts with `/`
	- the HTTP version is exactly `1.1`
- parses headers after a valid request line
- reads the body only when `Content-Length` is present and greater than zero
- marks the request complete once headers are done for bodyless requests
- returns errors for malformed or incomplete requests

Internal parser states currently implemented:

- initialized
- parsing headers
- parsing body
- done

### 4. Header parsing

The header parser lives in [internal/headers/headers.go](internal/headers/headers.go).

Implemented behavior:

- parses one header line at a time from raw bytes
- detects the blank line (`\r\n`) that ends the header section
- normalizes header names to lowercase for storage
- trims optional spaces and tabs around header values
- validates header field names
- rejects malformed header spacing around the colon
- concatenates repeated headers using commas
- exposes a case-insensitive getter through `Get()`

Header validation rules implemented in code include:

- no whitespace inside the header field name
- only valid token characters are allowed in field names
- malformed colon spacing is rejected

### 5. HTTP response helpers

Response-writing helpers live in [internal/response/response.go](internal/response/response.go).

Implemented behavior:

- writes HTTP/1.1 status lines
- supports these status codes:
	- `200 OK`
	- `400 Bad Request`
	- `500 Internal Server Error`
- writes headers in sorted order
- canonicalizes header key casing when writing to the wire
- provides a helper for generating default plain-text response headers

### 6. Learning utilities / experiments

This repository also includes small utilities used while learning sockets and request parsing.

#### TCP listener demo

The file [cmd/tcplistener/main.go](cmd/tcplistener/main.go) contains a simpler listener that:

- accepts TCP connections on port `42069`
- parses an incoming HTTP request using the custom request parser
- prints the parsed request line, headers, and body to stdout
- sends back a basic text response that includes the request target

#### UDP sender demo

The file [cmd/udpsender/main.go](cmd/udpsender/main.go) contains a small UDP client that:

- connects to `localhost:42069` over UDP
- reads lines from standard input
- sends each line as a UDP datagram

This is separate from the HTTP server and appears to be a socket-learning utility rather than part of the HTTP path.

## Tests currently implemented

Test coverage currently exists for header parsing and request parsing.

### Header tests

See [internal/headers/headers_test.go](internal/headers/headers_test.go).

Covered scenarios:

- valid single header parsing
- valid header parsing with extra whitespace in values
- preserving existing headers while parsing new data
- detecting the blank line that ends the header block
- rejecting invalid spacing in the header name
- rejecting invalid spacing patterns in values
- rejecting invalid characters in header names

### Request tests

See [internal/request/request_test.go](internal/request/request_test.go).

Covered scenarios:

- valid `GET / HTTP/1.1`
- valid `GET /coffee HTTP/1.1`
- rejection of unsupported HTTP version
- incremental header parsing from small chunks
- malformed header handling
- request body parsing using `Content-Length`
- empty body with `Content-Length: 0`
- empty body when `Content-Length` is missing
- detection of incomplete bodies
- ignoring body bytes when no `Content-Length` is provided

The tests use chunked readers to simulate how data arrives over a network connection in small pieces.

## Project structure

```text
.
├── cmd/
│   ├── httpserver/      # runnable HTTP server entrypoint
│   ├── tcplistener/     # request parsing / TCP experiment
│   └── udpsender/       # UDP client experiment
├── internal/
│   ├── headers/         # header parsing logic and tests
│   ├── request/         # request parsing logic and tests
│   ├── response/        # status line and header writing helpers
│   └── server/          # TCP server lifecycle and connection handling
├── tmp/                 # scratch/example files used during development
├── tcp.txt              # notes / sample output
├── go.mod
└── README.md
```

## Requirements

- Go `1.25.7` or newer matching [go.mod](go.mod)

## How to run

### Start the HTTP server

From the project root:

```bash
go run ./cmd/httpserver
```

The server listens on:

- `localhost:42069`

### Example request with curl

```bash
curl -i http://localhost:42069/
```

Expected response:

```http
HTTP/1.1 200 OK
Connection: close
Content-Length: 13
Content-Type: text/plain

Hello World!
```

### Run the TCP listener demo

```bash
go run ./cmd/tcplistener
```

### Run the UDP sender demo

```bash
go run ./cmd/udpsender
```

## How to test

Run all tests:

```bash
go test ./...
```

## Example implementation details

### Request line parsing

The request parser expects lines shaped like:

```text
METHOD /path HTTP/1.1\r\n
```

Example:

```text
GET /coffee HTTP/1.1\r\n
```

### Header storage behavior

Headers are stored in lowercase form. For example:

- `Host` becomes `host`
- `User-Agent` becomes `user-agent`

Repeated headers are merged into a comma-separated value.

### Body parsing behavior

The body is only read when `Content-Length` is present.

That means:

- if `Content-Length` is missing, the parser finishes after headers
- if `Content-Length: 0`, the parser finishes with an empty body
- if fewer bytes arrive than declared, parsing fails with an incomplete request error

## Current limitations

This project is intentionally small and educational. The following are not implemented yet:

- routing
- different handlers per path or method
- persistent connections / keep-alive support
- chunked transfer encoding
- query-string parsing
- HTTP/1.0 support
- HTTPS / TLS
- request timeouts
- structured logging
- full RFC-complete header validation
- middleware, static files, or JSON responses
- automatic `400` or `500` responses based on parse failures in the main server path

## Why this project is useful

This repository is a good reference for learning:

- how HTTP is just bytes over TCP
- how servers parse streamed data incrementally
- how headers and bodies are framed
- how responses are constructed manually
- how graceful shutdown works in a simple network service

## Next logical improvements

If development continues, useful next steps would be:

- connect the request parser to the main concurrent server handler
- return `400 Bad Request` for malformed requests
- add method/path-based routing
- support reading and responding to `POST` payloads
- improve tests around response writing and server behavior
- add integration tests using real TCP connections
- document example raw HTTP requests and responses

## License

See [LICENSE](LICENSE).
