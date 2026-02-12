package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("issuer_id: issuer\nkey_id: key\nbundle_id: bundle\nprivate_key: filekey\nenv: production\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("ASK_KEY_ID", "envkey")
	loaded, err := Load(Options{ConfigPath: path})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.KeyID != "envkey" {
		t.Fatalf("expected env override, got %s", loaded.KeyID)
	}

	loaded, err = Load(Options{ConfigPath: path, Overrides: Config{KeyID: "flagkey"}})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.KeyID != "flagkey" {
		t.Fatalf("expected override, got %s", loaded.KeyID)
	}
}

func TestConfigValidateRequired(t *testing.T) {
	cfg := Config{}
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected validation error")
	}
}
