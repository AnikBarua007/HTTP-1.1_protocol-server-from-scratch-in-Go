package response

import (
	"fmt"
	"io"
	"net/textproto"
	"sort"
	"strconv"

	"_http_protocol_1.1/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reason := ""
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	}

	if reason == "" {
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", statusCode)
		return err
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reason)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", textproto.CanonicalMIMEHeaderKey(k), headers[k]); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, "\r\n")
	return err
}
