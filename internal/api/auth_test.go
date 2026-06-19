package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestAuthNoTokenConfigured(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Post(url(ts, "/api/store"), "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAuthMissingToken(t *testing.T) {
	_, ts := newTestServerWithConfig(t, Config{Token: "secret123"})

	resp, err := http.Post(url(ts, "/api/store"), "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthWrongToken(t *testing.T) {
	_, ts := newTestServerWithConfig(t, Config{Token: "secret123"})

	req, _ := http.NewRequest("POST", url(ts, "/api/store"), strings.NewReader("hello"))
	req.Header.Set("X-Auth-Token", "wrong")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthValidToken(t *testing.T) {
	_, ts := newTestServerWithConfig(t, Config{Token: "secret123"})

	req, _ := http.NewRequest("POST", url(ts, "/api/store"), strings.NewReader("hello"))
	req.Header.Set("X-Auth-Token", "secret123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAuthExemptsRootAndStaticAndHealth(t *testing.T) {
	_, ts := newTestServerWithConfig(t, Config{Token: "secret123"})

	// Root should be accessible without token
	resp, err := http.Get(url(ts, "/"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("root should not require auth")
	}

	// Static files should be accessible without token
	resp2, err := http.Get(url(ts, "/static/script.js"))
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode == http.StatusUnauthorized {
		t.Fatal("static files should not require auth")
	}

	// Health endpoint should be accessible without token
	resp3, err := http.Get(url(ts, "/api/health"))
	if err != nil {
		t.Fatal(err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp3.StatusCode)
	}
}
