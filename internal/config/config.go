package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	IssuerID         string        `mapstructure:"issuer_id" yaml:"issuer_id"`
	KeyID            string        `mapstructure:"key_id" yaml:"key_id"`
	BundleID         string        `mapstructure:"bundle_id" yaml:"bundle_id"`
	PrivateKeyPath   string        `mapstructure:"private_key_path" yaml:"private_key_path"`
	PrivateKey       string        `mapstructure:"private_key" yaml:"private_key"`
	Environment      string        `mapstructure:"env" yaml:"env"`
	MaxRetries       int           `mapstructure:"max_retries" yaml:"max_retries"`
	RetryBackoff     time.Duration `mapstructure:"retry_backoff" yaml:"retry_backoff"`
	RetryBackoffMS   int           `mapstructure:"retry_backoff_ms" yaml:"retry_backoff_ms"`
	RequestTimeout   time.Duration `mapstructure:"timeout" yaml:"timeout"`
	RequestTimeoutMS int           `mapstructure:"timeout_ms" yaml:"timeout_ms"`
}

type Options struct {
	ConfigPath string
	Overrides  Config
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ask", "config.yaml"), nil
}

func Load(opts Options) (Config, error) {
	v := viper.New()

	cfgPath := opts.ConfigPath
	if cfgPath == "" {
		defaultPath, err := DefaultPath()
		if err != nil {
			return Config{}, err
		}
		cfgPath = defaultPath
	}

	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, err
		}
	}

	v.SetEnvPrefix("ASK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := bindEnvKeys(v); err != nil {
		return Config{}, err
	}

	v.SetDefault("env", "production")
	v.SetDefault("max_retries", 3)
	v.SetDefault("retry_backoff_ms", 500)
	v.SetDefault("timeout_ms", 30000)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	applyOverrides(&cfg, opts.Overrides)
	normalizeDurations(&cfg)

	return cfg, nil
}

func bindEnvKeys(v *viper.Viper) error {
	if err := v.BindEnv("issuer_id"); err != nil {
		return err
	}
	if err := v.BindEnv("key_id"); err != nil {
		return err
	}
	if err := v.BindEnv("bundle_id"); err != nil {
		return err
	}
	if err := v.BindEnv("private_key_path"); err != nil {
		return err
	}
	if err := v.BindEnv("private_key"); err != nil {
		return err
	}
	if err := v.BindEnv("env"); err != nil {
		return err
	}
	if err := v.BindEnv("max_retries"); err != nil {
		return err
	}
	if err := v.BindEnv("retry_backoff_ms"); err != nil {
		return err
	}
	if err := v.BindEnv("timeout_ms"); err != nil {
		return err
	}
	return nil
}

func normalizeDurations(cfg *Config) {
	if cfg.RetryBackoff == 0 && cfg.RetryBackoffMS > 0 {
		cfg.RetryBackoff = time.Duration(cfg.RetryBackoffMS) * time.Millisecond
	}
	if cfg.RequestTimeout == 0 && cfg.RequestTimeoutMS > 0 {
		cfg.RequestTimeout = time.Duration(cfg.RequestTimeoutMS) * time.Millisecond
	}
}

func applyOverrides(cfg *Config, overrides Config) {
	if overrides.IssuerID != "" {
		cfg.IssuerID = overrides.IssuerID
	}
	if overrides.KeyID != "" {
		cfg.KeyID = overrides.KeyID
	}
	if overrides.BundleID != "" {
		cfg.BundleID = overrides.BundleID
	}
	if overrides.PrivateKeyPath != "" {
		cfg.PrivateKeyPath = overrides.PrivateKeyPath
	}
	if overrides.PrivateKey != "" {
		cfg.PrivateKey = overrides.PrivateKey
	}
	if overrides.Environment != "" {
		cfg.Environment = overrides.Environment
	}
	if overrides.MaxRetries != 0 {
		cfg.MaxRetries = overrides.MaxRetries
	}
	if overrides.RetryBackoffMS != 0 {
		cfg.RetryBackoffMS = overrides.RetryBackoffMS
	}
	if overrides.RequestTimeoutMS != 0 {
		cfg.RequestTimeoutMS = overrides.RequestTimeoutMS
	}
	if overrides.RequestTimeout != 0 {
		cfg.RequestTimeout = overrides.RequestTimeout
	}
}
