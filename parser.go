package main

import (
	"fmt"
	"strings"
)

type HTTPRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Params  map[string]string
	Body    []byte
}

func DecodeHTTPRequest(request string) (*HTTPRequest, error) {
	// separa entre headers e (possivelmente) body
	sections := strings.Split(request, "\r\n\r\n")

	// separa a request line + cada linha dos headers
	lines := strings.Split(sections[0], "\r\n")

	// parsing da request line
	requestLine := strings.Fields(lines[0])
	if len(requestLine) != 3 {
		return nil, fmt.Errorf("Invalid request line: %s", lines[0])
	}

	method := requestLine[0]
	path := requestLine[1]
	version := requestLine[2]

	// parsing dos headers
	headers := make(map[string]string)
	for _, pair := range lines[1:] {
		// divide em duas partes, isso é, para na primeira ocorrência de `:`
		parts := strings.SplitN(pair, ":", 2)

		if len(parts) != 2 {
			return nil, fmt.Errorf("Invalid header line: %s", pair)
		}

		// removendo possíveis espaços
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		headers[key] = value
	}

	return &HTTPRequest{
		Method:  method,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    []byte(sections[1]),
	}, nil
}
