package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
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
