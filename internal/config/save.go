package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg) // #nosec G117 - intentionally serializing config including private_key
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
