package config

import (
	"fmt"
	"os"
)

var allowedEnvironments = map[string]struct{}{
	"production":    {},
	"sandbox":       {},
	"local-testing": {},
}

func Validate(cfg Config) error {
	if cfg.IssuerID == "" {
		return fmt.Errorf("issuer_id is required")
	}
	if cfg.KeyID == "" {
		return fmt.Errorf("key_id is required")
	}
	if cfg.BundleID == "" {
		return fmt.Errorf("bundle_id is required")
	}
	if cfg.Environment == "" {
		return fmt.Errorf("env is required")
	}
	if _, ok := allowedEnvironments[cfg.Environment]; !ok {
		return fmt.Errorf("env must be one of: production, sandbox, local-testing")
	}
	if cfg.PrivateKey == "" && cfg.PrivateKeyPath == "" {
		return fmt.Errorf("private_key or private_key_path is required")
	}
	if cfg.PrivateKeyPath != "" {
		if _, err := os.Stat(cfg.PrivateKeyPath); err != nil {
			return fmt.Errorf("private_key_path is invalid: %w", err)
		}
	}
	return nil
}
