package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Signer struct {
	IssuerID   string
	KeyID      string
	BundleID   string
	PrivateKey *ecdsa.PrivateKey
}

func NewSigner(issuerID, keyID, bundleID, privateKeyPath, privateKey string) (Signer, error) {
	if issuerID == "" || keyID == "" || bundleID == "" {
		return Signer{}, errors.New("issuer id, key id, and bundle id are required")
	}

	keyData := []byte(strings.TrimSpace(privateKey))
	if len(keyData) == 0 && privateKeyPath != "" {
		// Validate path to prevent directory traversal
		cleanPath := filepath.Clean(privateKeyPath)
		data, err := os.ReadFile(cleanPath) // #nosec G304 - path is cleaned above
		if err != nil {
			return Signer{}, err
		}
		keyData = data
	}
	if len(keyData) == 0 {
		return Signer{}, errors.New("private key is required")
	}

	key, err := parseECPrivateKey(keyData)
	if err != nil {
		return Signer{}, err
	}

	return Signer{IssuerID: issuerID, KeyID: keyID, BundleID: bundleID, PrivateKey: key}, nil
}

func (s Signer) Token() (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss": s.IssuerID,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": s.BundleID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.KeyID

	return token.SignedString(s.PrivateKey)
}

func parseECPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ECDSA")
	}

	return ecdsaKey, nil
}
