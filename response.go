package httpatos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

var statusText = map[int]string{
	100: "Continue",
	101: "Switching Protocols",

	200: "OK",
	201: "Created",
	202: "Accepted",
	204: "No Content",
	206: "Partial Content",

	301: "Moved Permanently",
	302: "Found",
	304: "Not Modified",
	307: "Temporary Redirect",
	308: "Permanent Redirect",

	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	413: "Content Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	422: "Unprocessable Content",
	429: "Too Many Requests",

	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

type httpResponse struct {
	Version    string
	Status     int
	StatusText string
	Headers    map[string]string
	Body       []byte

	// listener espera até que Ready == true
	Ready bool
}

func (r *httpResponse) Json(payload any) *httpResponse {
	body, _ := json.Marshal(payload)

	r.Headers["Content-Type"] = "application/json"
	r.Body = body

	return r
}

func (r *httpResponse) Text(text string) *httpResponse {
	body := []byte(text)

	r.Headers["Content-Type"] = "text/plain"
	r.Body = body

	return r
}

func (r *httpResponse) Send(statusCode int) {
	var message string

	if text, ok := statusText[statusCode]; ok {
		message = text
	} else {
		message = "Unknown"
	}

	r.StatusText = message
	r.Status = statusCode

	r.Headers["Content-Length"] = strconv.Itoa(len(r.Body))

	r.Ready = true
}

func (r *httpResponse) serialize() []byte {
	var buf bytes.Buffer

	// status line
	fmt.Fprintf(&buf, "HTTP/1.1 %d %s\r\n", r.Status, r.StatusText)

	// headers
	for key, value := range r.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
	}

	// divisão entre header e body
	buf.WriteString("\r\n")

	// body
	buf.Write(r.Body)

	return buf.Bytes()
}

func NewHTTPResponse() *httpResponse {
	return &httpResponse{
		Headers: make(map[string]string),
	}
}
