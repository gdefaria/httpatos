package httpatos

import (
	"encoding/json"
	"fmt"
	"strings"
)

type httpRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Params  map[string]string
	Body    []byte
}

// Parser de JSON
func (r *httpRequest) JSON(expectedPayload any) (success bool) {
	contentType := r.Headers["content-type"]

	if contentType != "application/json" {
		// TODO: handle bad request
		fmt.Println("Bad content-type. Expected application/json")
		return false
	}

	err := json.Unmarshal(r.Body, expectedPayload)

	if err != nil {
		// TODO: handle bad request
		fmt.Println("Invalid JSON body", err)
		return false
	}

	json.Unmarshal(r.Body, expectedPayload)
	return true
}

func decodeHttpRequest(request string) (*httpRequest, error) {
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
		key := strings.ToLower(strings.TrimSpace(parts[0])) // campos do header são case-insensitive (RFC 9110 §5.1)
		value := strings.TrimSpace(parts[1])

		headers[key] = value
	}

	// parsing do body
	var body []byte

	if len(sections) > 1 {
		body = []byte(sections[1])
	}

	return &httpRequest{
		Method:  method,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    body,
	}, nil
}
