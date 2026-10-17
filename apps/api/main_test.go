package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloRoute(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		status int
		body   string
	}{
		{"greets the username", http.MethodGet, "/hello/Max", http.StatusOK, "Hello Max"},
		{"decodes the username", http.MethodGet, "/hello/J%C3%B6rg%20M%C3%BCller", http.StatusOK, "Hello Jörg Müller"},
		{"requires a username", http.MethodGet, "/hello/", http.StatusNotFound, ""},
		{"rejects nested paths", http.MethodGet, "/hello/Max/extra", http.StatusNotFound, ""},
		{"rejects POST requests", http.MethodPost, "/hello/Max", http.StatusMethodNotAllowed, ""},
	}

	handler := newHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
			if tt.status == http.StatusOK {
				if body := response.Body.String(); body != tt.body {
					t.Errorf("body = %q, want %q", body, tt.body)
				}
				if contentType := response.Header().Get("Content-Type"); contentType != "text/plain; charset=utf-8" {
					t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", contentType)
				}
			}
		})
	}
}
