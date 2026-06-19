package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Adhnan23/karots-drop/internal/store"
)

func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	return newTestServerWithConfig(t, Config{})
}

func newTestServerWithConfig(t *testing.T, cfg Config) (*Server, *httptest.Server) {
	t.Helper()
	s := store.New()
	t.Cleanup(s.Stop)
	srv := New(cfg, s, nil, nil)
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)
	return srv, ts
}

func url(ts *httptest.Server, path string) string {
	return ts.URL + path
}

func TestStoreGetTextRoundtrip(t *testing.T) {
	_, ts := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("text", "hello world")
	w.Close()

	resp, err := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result storeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if len(result.Code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", result.Code)
	}
	if result.TTL != 1200 {
		t.Fatalf("expected TTL 1200, got %d", result.TTL)
	}
	if result.Encrypted {
		t.Fatal("expected encrypted=false")
	}

	resp2, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", body)
	}
}

func TestStoreGetFileRoundtrip(t *testing.T) {
	_, ts := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", "test.txt")
	part.Write([]byte("file content"))
	w.Close()

	resp, err := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result storeResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Filename != "test.txt" {
		t.Fatalf("expected 'test.txt', got %q", result.Filename)
	}

	resp2, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	cd := resp2.Header.Get("Content-Disposition")
	if !strings.Contains(cd, "test.txt") {
		t.Fatalf("expected Content-Disposition with test.txt, got %q", cd)
	}

	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "file content" {
		t.Fatalf("expected 'file content', got %q", body)
	}
}

func TestStoreRawBody(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Post(url(ts, "/api/store"), "text/plain", strings.NewReader("raw body"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result storeResponse
	json.NewDecoder(resp.Body).Decode(&result)

	resp2, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "raw body" {
		t.Fatalf("expected 'raw body', got %q", body)
	}
}

func TestGetNotFound(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(url(ts, "/api/get/000000"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	var errResp errorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestGetMissingCode(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(url(ts, "/api/get/"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// /api/get/ without a code segment doesn't match the route,
	// falls through to default 404
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestStoreEmptyBody(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Post(url(ts, "/api/store"), "text/plain", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestStoreEmptyForm(t *testing.T) {
	_, ts := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.Close()

	resp, err := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestQRCodeEndpoint(t *testing.T) {
	_, ts := newTestServer(t)

	// First store something
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("text", "qr test")
	w.Close()

	resp, _ := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	var result storeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	// Get QR
	resp2, err := http.Get(url(ts, "/api/qr/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	ct := resp2.Header.Get("Content-Type")
	if ct != "image/png" {
		t.Fatalf("expected 'image/png', got %q", ct)
	}

	png, _ := io.ReadAll(resp2.Body)
	if len(png) == 0 {
		t.Fatal("expected non-empty PNG")
	}

	// Validate PNG header
	expected := []byte{0x89, 0x50, 0x4E, 0x47}
	for i, b := range expected {
		if png[i] != b {
			t.Fatalf("invalid PNG magic at byte %d", i)
		}
	}
}

func TestQRCodeNotFound(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(url(ts, "/api/qr/000000"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCORSPreflight(t *testing.T) {
	_, ts := newTestServer(t)

	req, _ := http.NewRequest("OPTIONS", url(ts, "/api/store"), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected CORS header")
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestStoreEncryptedFlag(t *testing.T) {
	_, ts := newTestServer(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("text", "encrypted data")
	w.WriteField("encrypted", "true")
	w.Close()

	resp, _ := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	var result storeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if !result.Encrypted {
		t.Fatal("expected encrypted=true")
	}

	resp2, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Encrypted") != "true" {
		t.Fatal("expected X-Encrypted header")
	}
}

func TestStoreMaxSize(t *testing.T) {
	_, ts := newTestServer(t)

	// Create data just over 20MB
	data := make([]byte, 21<<20)
	for i := range data {
		data[i] = byte(i)
	}

	resp, err := http.Post(url(ts, "/api/store"), "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized payload, got %d", resp.StatusCode)
	}
}

func TestDeleteOnRetrieve(t *testing.T) {
	_, ts := newTestServerWithConfig(t, Config{DeleteOnRetrieve: true})

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("text", "delete me")
	w.Close()

	resp, err := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	var result storeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	// First retrieve should work
	resp2, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "delete me" {
		t.Fatalf("expected 'delete me', got %q", body)
	}

	// Second retrieve should 404 (already deleted)
	resp3, err := http.Get(url(ts, "/api/get/"+result.Code))
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp3.StatusCode)
	}
}

func TestHealthEndpoint(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(url(ts, "/api/health"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var health struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatal(err)
	}
	if health.Status != "ok" {
		t.Fatalf("expected status 'ok', got %q", health.Status)
	}

	// Health should work even when auth is enabled
	_, ts2 := newTestServerWithConfig(t, Config{Token: "secret"})
	defer ts2.Close()

	resp2, err := http.Get(url(ts2, "/api/health"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with auth, got %d", resp2.StatusCode)
	}
}

func TestConcurrentStore(t *testing.T) {
	_, ts := newTestServer(t)

	errs := make(chan error, 50)
	for i := range 50 {
		go func(i int) {
			var buf bytes.Buffer
			w := multipart.NewWriter(&buf)
			w.WriteField("text", fmt.Sprintf("msg-%d", i))
			w.Close()

			resp, err := http.Post(url(ts, "/api/store"), w.FormDataContentType(), &buf)
			if err != nil {
				errs <- err
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				errs <- fmt.Errorf("got status %d", resp.StatusCode)
				return
			}
			errs <- nil
		}(i)
	}

	for range 50 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}
