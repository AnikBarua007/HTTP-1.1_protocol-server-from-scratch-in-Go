package request

import (
	"fmt"
	"io"
	"strings"
)

type parserState int

const (
	stateInitialized parserState = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
	State       parserState
}
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{State: stateInitialized}
	buffer := make([]byte, 8)
	buffered := 0
	bytesRead := 0
	bytesParsed := 0

	for req.State != stateDone {
		if buffered == len(buffer) {
			grown := make([]byte, len(buffer)*2)
			copy(grown, buffer[:buffered])
			buffer = grown
		}

		n, readErr := reader.Read(buffer[buffered:])
		if n > 0 {
			buffered += n
			bytesRead += n
		}

		for buffered > 0 {
			consumed, parseErr := req.parse(buffer[:buffered])
			if parseErr != nil {
				return nil, parseErr
			}
			if consumed == 0 {
				break
			}
			bytesParsed += consumed
			copy(buffer, buffer[consumed:buffered])
			buffered -= consumed

			if bytesRead-bytesParsed != buffered {
				return nil, fmt.Errorf("internal parser accounting error")
			}
			if req.State == stateDone {
				break
			}
		}

		if readErr == io.EOF {
			if req.State != stateDone {
				return nil, fmt.Errorf("incomplete request line")
			}
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}

	return req, nil
}

func parseRequestLine(line []byte) (int, []string, error) {
	idx := strings.Index(string(line), "\r\n")
	if idx == -1 {
		// Need more data before a full request line can be parsed.
		return 0, nil, nil
	}

	requestLine := string(line[:idx])
	parts := strings.Fields(requestLine)
	if len(parts) != 3 {
		return idx + 2, nil, fmt.Errorf("invalid request line")
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return idx + 2, nil, fmt.Errorf("invalid request line")
	}

	parts[2] = strings.TrimPrefix(parts[2], "HTTP/")
	return idx + 2, parts, nil
}
func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case stateInitialized:
		consumed, parts, err := parseRequestLine(data)
		if err != nil {
			return consumed, err
		}
		if consumed == 0 {
			return 0, nil
		}
		if len(parts) != 3 {
			return consumed, fmt.Errorf("invalid request line")
		}

		method := parts[0]
		target := parts[1]
		version := parts[2]

		if method != strings.ToUpper(method) {
			return consumed, fmt.Errorf("invalid method: %s", method)
		}
		if version != "1.1" {
			return consumed, fmt.Errorf("unsupported http version: %s", version)
		}
		if !strings.HasPrefix(target, "/") {
			return consumed, fmt.Errorf("invalid request target: %s", target)
		}

		r.RequestLine = RequestLine{
			HttpVersion:   version,
			RequestTarget: target,
			Method:        method,
		}
		r.State = stateDone
		return consumed, nil
	case stateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}
