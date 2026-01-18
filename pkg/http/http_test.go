package http

import (
	"testing"
)

func TestParseHttpRequest_ValidGETRequest(t *testing.T) {
	rawRequest := "GET /api/users HTTP/1.1\r\nHost: example.com\r\nUser-Agent: test-client\r\nAccept: application/json\r\n\r\n"
	data := []byte(rawRequest)

	hm, err := ParseHttpRequest(len(data), data)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if hm.Method != "GET" {
		t.Errorf("Expected method GET, got: %s", hm.Method)
	}

	if hm.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got: %s", hm.Path)
	}

	if hm.Protocol != "HTTP/1.1" {
		t.Errorf("Expected protocol HTTP/1.1, got: %s", hm.Protocol)
	}

	if hm.Host != "example.com" {
		t.Errorf("Expected host example.com, got: %s", hm.Host)
	}

	if hm.Headers["User-Agent"] != "test-client" {
		t.Errorf("Expected User-Agent test-client, got: %s", hm.Headers["User-Agent"])
	}

	if hm.RawBody != "" {
		t.Errorf("Expected empty body, got: %s", hm.RawBody)
	}
}

func TestParseHttpRequest_ValidPOSTRequestWithBody(t *testing.T) {
	rawRequest := "POST /api/users HTTP/1.1\r\nHost: api.example.com\r\nContent-Type: application/json\r\nContent-Length: 27\r\n\r\n{\"name\":\"John\",\"age\":30}"
	data := []byte(rawRequest)

	hm, err := ParseHttpRequest(len(data), data)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if hm.Method != "POST" {
		t.Errorf("Expected method POST, got: %s", hm.Method)
	}

	if hm.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got: %s", hm.Path)
	}

	if hm.Host != "api.example.com" {
		t.Errorf("Expected host api.example.com, got: %s", hm.Host)
	}

	expectedBody := "{\"name\":\"John\",\"age\":30}"
	if hm.RawBody != expectedBody {
		t.Errorf("Expected body %s, got: %s", expectedBody, hm.RawBody)
	}

	if hm.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got: %s", hm.Headers["Content-Type"])
	}
}

func TestParseHttpRequest_InvalidStartLine(t *testing.T) {
	tests := []struct {
		name    string
		request string
	}{
		{
			name:    "missing protocol",
			request: "GET /api/users\r\nHost: example.com\r\n\r\n",
		},
		{
			name:    "empty startLine",
			request: "\r\nHost: example.com\r\n\r\n",
		},
		{
			name:    "too many parts",
			request: "GET /api/users HTTP/1.1 EXTRA\r\nHost: example.com\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.request)
			_, err := ParseHttpRequest(len(data), data)

			if err == nil {
				t.Errorf("Expected error for invalid startLine, got nil")
			}
		})
	}
}

func TestParseHttpRequest_MissingHost(t *testing.T) {
	rawRequest := "GET /api/users HTTP/1.1\r\nUser-Agent: test-client\r\n\r\n"
	data := []byte(rawRequest)

	_, err := ParseHttpRequest(len(data), data)

	if err == nil {
		t.Errorf("Expected error for missing Host header, got nil")
	}

	if err.Error() != "invalid host" {
		t.Errorf("Expected 'invalid host' error, got: %v", err)
	}
}

func TestParseHttpRequest_InvalidHeaderFormat(t *testing.T) {
	rawRequest := "GET /api/users HTTP/1.1\r\nHost: example.com\r\nInvalidHeader\r\n\r\n"
	data := []byte(rawRequest)

	_, err := ParseHttpRequest(len(data), data)

	if err == nil {
		t.Errorf("Expected error for invalid header format, got nil")
	}
}

func TestParseHttpRequest_EmptyData(t *testing.T) {
	data := []byte("")

	_, err := ParseHttpRequest(len(data), data)

	if err == nil {
		t.Errorf("Expected error for empty data, got nil")
	}
}

func TestParseHttpRequest_HeadersWithSpaces(t *testing.T) {
	rawRequest := "GET /test HTTP/1.1\r\nHost: example.com\r\nX-Custom-Header:   value with spaces   \r\n\r\n"
	data := []byte(rawRequest)

	hm, err := ParseHttpRequest(len(data), data)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if hm.Headers["X-Custom-Header"] != "value with spaces" {
		t.Errorf("Expected trimmed header value 'value with spaces', got: '%s'", hm.Headers["X-Custom-Header"])
	}
}

func TestParseHttpRequest_HeaderWithColonInValue(t *testing.T) {
	rawRequest := "GET /test HTTP/1.1\r\nHost: example.com\r\nX-Time: 12:30:45\r\n\r\n"
	data := []byte(rawRequest)

	hm, err := ParseHttpRequest(len(data), data)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if hm.Headers["X-Time"] != "12:30:45" {
		t.Errorf("Expected header value '12:30:45', got: '%s'", hm.Headers["X-Time"])
	}
}

func TestParseHttpRequest_MultipleHeaders(t *testing.T) {
	rawRequest := "GET /api/data HTTP/1.1\r\nHost: api.example.com\r\nAccept: application/json\r\nAccept-Encoding: gzip, deflate\r\nConnection: keep-alive\r\nAuthorization: Bearer token123\r\n\r\n"
	data := []byte(rawRequest)

	hm, err := ParseHttpRequest(len(data), data)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedHeaders := map[string]string{
		"Host":            "api.example.com",
		"Accept":          "application/json",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
		"Authorization":   "Bearer token123",
	}

	for key, expectedValue := range expectedHeaders {
		if hm.Headers[key] != expectedValue {
			t.Errorf("Expected header %s to be '%s', got: '%s'", key, expectedValue, hm.Headers[key])
		}
	}
}
