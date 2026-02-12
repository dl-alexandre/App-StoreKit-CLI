package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignerClaims(t *testing.T) {
	privateKey, pemKey := generateTestKey(t)

	signer, err := NewSigner("issuer", "key", "com.example.app", "", pemKey)
	if err != nil {
		t.Fatalf("expected signer, got error: %v", err)
	}

	token, err := signer.Token()
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("expected token parse ok, got error: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected map claims")
	}

	if claims["iss"] != "issuer" {
		t.Fatalf("expected iss issuer, got %v", claims["iss"])
	}
	if claims["bid"] != "com.example.app" {
		t.Fatalf("expected bid com.example.app, got %v", claims["bid"])
	}
	if claims["aud"] != "appstoreconnect-v1" {
		t.Fatalf("expected aud appstoreconnect-v1, got %v", claims["aud"])
	}

	issuedAt := int64(claims["iat"].(float64))
	expiresAt := int64(claims["exp"].(float64))
	if expiresAt <= issuedAt {
		t.Fatalf("expected exp after iat")
	}
	if expiresAt-issuedAt != int64((5 * time.Minute).Seconds()) {
		t.Fatalf("expected exp-iat 300 seconds, got %d", expiresAt-issuedAt)
	}
}

func generateTestKey(t *testing.T) (*ecdsa.PrivateKey, string) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	return key, string(pem.EncodeToMemory(block))
}
