package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/auth"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/retry"
)

func TestClientRetryAfter(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errorCode":"RATE_LIMIT","errorMessage":"slow down"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	signer := newTestSigner(t)
	client := Client{
		BaseURL: server.URL,
		HTTP:    server.Client(),
		Signer:  signer,
		Retry: retry.Config{
			MaxRetries: 1,
			Backoff:    10 * time.Millisecond,
		},
	}

	start := time.Now()
	resp, err := client.Do(context.Background(), "GET", "/inApps/v1/notifications/test", nil, nil, "")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Status)
	}
	if time.Since(start) < time.Second {
		t.Fatalf("expected retry-after wait")
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorCode":"INVALID","errorMessage":"bad request"}`))
	}))
	defer server.Close()

	signer := newTestSigner(t)
	client := Client{BaseURL: server.URL, HTTP: server.Client(), Signer: signer}

	_, err := client.Do(context.Background(), "GET", "/inApps/v1/notifications/test", nil, nil, "")
	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != "INVALID" {
		t.Fatalf("expected code INVALID, got %s", apiErr.Code)
	}
	if apiErr.Message != "bad request" {
		t.Fatalf("expected message, got %s", apiErr.Message)
	}
}

func TestClientBuildsAuthenticatedRequestAndEncodesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/inApps/v1/test" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("revision") != "r 1" {
			t.Errorf("revision = %q", r.URL.Query().Get("revision"))
		}
		if got := r.Header.Get("X-Request-ID"); got != "request-123" {
			t.Errorf("X-Request-ID = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-client" {
			t.Errorf("User-Agent = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		token := r.Header.Get("Authorization")
		if len(token) < len("Bearer ") || token[:len("Bearer ")] != "Bearer " {
			t.Errorf("Authorization = %q", token)
		}
		var body map[string]bool
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !body["enabled"] {
			t.Errorf("body = %#v, decode error = %v", body, err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, HTTP: server.Client(), Signer: newTestSigner(t), RequestID: "request-123", UserAgent: "test-client"}
	response, err := client.Do(context.Background(), http.MethodPost, "/inApps/v1/test", map[string][]string{"revision": {"r 1"}}, []byte(`{"enabled":true}`), "")
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if response.Status != http.StatusCreated || string(response.Body) != `{"ok":true}` {
		t.Errorf("response = %+v", response)
	}
}

func TestClientDoesNotRetryClientError(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errorCode":"FORBIDDEN"}`))
	}))
	defer server.Close()

	client := Client{BaseURL: server.URL, HTTP: server.Client(), Signer: newTestSigner(t), Retry: retry.Config{MaxRetries: 3, Backoff: time.Millisecond}}
	_, err := client.Do(context.Background(), http.MethodGet, "/inApps/v1/test", nil, nil, "")
	if err == nil || requests != 1 {
		t.Fatalf("Do() error = %v, requests = %d; want one 403 request", err, requests)
	}
}

func newTestSigner(t *testing.T) auth.Signer {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	pemKey := string(pem.EncodeToMemory(block))

	signer, err := auth.NewSigner("issuer", "key", "com.example.app", "", pemKey)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	return signer
}
