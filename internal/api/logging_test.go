package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)
	if rw.status != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rw.status)
	}

	n, err := rw.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes written, got %d", n)
	}
	if rw.written != 5 {
		t.Fatalf("expected written count 5, got %d", rw.written)
	}
	if w.Body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", w.Body.String())
	}
}

func TestLoggingMiddlewareOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, req)

	output := buf.String()
	if !strings.Contains(output, "GET /test 200") {
		t.Fatalf("expected log line with method, path, status; got %q", output)
	}
	if !strings.Contains(output, " 2 ") { // 2 bytes body
		t.Fatalf("expected log line with body size 2; got %q", output)
	}
}

func TestLoggingMiddlewareErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"not found"}`)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/get/000000", nil)
	handler.ServeHTTP(w, req)

	output := buf.String()
	if !strings.Contains(output, "404") {
		t.Fatalf("expected log line with 404; got %q", output)
	}
}
