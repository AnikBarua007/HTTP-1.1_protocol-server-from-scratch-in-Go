package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"_http_protocol_1.1/internal/headers"
)

type parserState int

const (
	requestStateInitialized parserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       parserState
	Body        []byte
	bodyBytes   int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		State:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	buffer := make([]byte, 8)
	readToIndex := 0
	bytesRead := 0
	bytesParsed := 0

	for req.State != requestStateDone {
		if readToIndex == len(buffer) {
			grown := make([]byte, len(buffer)*2)
			copy(grown, buffer[:readToIndex])
			buffer = grown
		}

		n, readErr := reader.Read(buffer[readToIndex:])
		if n > 0 {
			readToIndex += n
			bytesRead += n
		}

		consumed, parseErr := req.parse(buffer[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}
		if consumed > 0 {
			bytesParsed += consumed
			copy(buffer, buffer[consumed:readToIndex])
			readToIndex -= consumed
			if bytesRead-bytesParsed != readToIndex {
				return nil, fmt.Errorf("internal parser accounting error")
			}
		}

		if readErr == io.EOF {
			if req.State != requestStateDone {
				return nil, fmt.Errorf("incomplete request")
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
	totalBytesParsed := 0

	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			return totalBytesParsed, nil
		}

		totalBytesParsed += n
		if totalBytesParsed > len(data) {
			return totalBytesParsed, fmt.Errorf("parsed beyond buffer")
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
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
		r.State = requestStateParsingHeaders
		return consumed, nil

	case requestStateParsingHeaders:
		consumed, done, err := r.Headers.Parse(data)
		if err != nil {
			return consumed, err
		}
		if done {
			contentLength := r.Headers["content-length"]
			if contentLength == "" {
				r.State = requestStateDone
				return consumed, nil
			}

			n, convErr := strconv.Atoi(contentLength)
			if convErr != nil || n < 0 {
				return consumed, fmt.Errorf("invalid content-length: %s", contentLength)
			}
			r.bodyBytes = n
			if n == 0 {
				r.State = requestStateDone
				return consumed, nil
			}
			r.State = requestStateParsingBody
		}
		return consumed, nil

	case requestStateParsingBody:
		if r.bodyBytes == 0 {
			r.State = requestStateDone
			return 0, nil
		}
		toConsume := len(data)
		if toConsume > r.bodyBytes {
			toConsume = r.bodyBytes
		}
		if toConsume == 0 {
			return 0, nil
		}

		r.Body = append(r.Body, data[:toConsume]...)
		r.bodyBytes -= toConsume
		if r.bodyBytes == 0 {
			r.State = requestStateDone
		}
		return toConsume, nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}
