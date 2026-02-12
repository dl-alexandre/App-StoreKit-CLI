package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(newConfigInitCommand())
	cmd.AddCommand(newConfigValidateCommand())
	return cmd
}

func newConfigInitCommand() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(path)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "config file path")
	return cmd
}

func newConfigValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, err := config.Load(config.Options{ConfigPath: cfgPath})
			if err != nil {
				return err
			}
			if err := config.Validate(cfg); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "config ok")
			return nil
		},
	}
	return cmd
}

func runConfigInit(path string) error {
	if path == "" {
		defaultPath, err := config.DefaultPath()
		if err != nil {
			return err
		}
		path = defaultPath
	}

	reader := bufio.NewReader(os.Stdin)
	issuerID, err := prompt(reader, "Issuer ID")
	if err != nil {
		return err
	}
	keyID, err := prompt(reader, "Key ID")
	if err != nil {
		return err
	}
	bundleID, err := prompt(reader, "Bundle ID")
	if err != nil {
		return err
	}
	keyPath, err := prompt(reader, "Private key path (.p8)")
	if err != nil {
		return err
	}
	env, err := promptDefault(reader, "Environment (sandbox|production|local-testing)", "production")
	if err != nil {
		return err
	}
	retriesText, err := promptDefault(reader, "Max retries", "3")
	if err != nil {
		return err
	}
	backoffText, err := promptDefault(reader, "Retry backoff (ms)", "500")
	if err != nil {
		return err
	}

	retries, err := strconv.Atoi(retriesText)
	if err != nil {
		return err
	}
	backoffMs, err := strconv.Atoi(backoffText)
	if err != nil {
		return err
	}

	cfg := config.Config{
		IssuerID:       issuerID,
		KeyID:          keyID,
		BundleID:       bundleID,
		PrivateKeyPath: keyPath,
		Environment:    env,
		MaxRetries:     retries,
		RetryBackoffMS: backoffMs,
	}

	if err := config.Save(path, cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Config saved to %s\n", path)
	return nil
}

func prompt(reader *bufio.Reader, label string) (string, error) {
	fmt.Fprintf(os.Stdout, "%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptDefault(reader *bufio.Reader, label, def string) (string, error) {
	fmt.Fprintf(os.Stdout, "%s [%s]: ", label, def)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return def, nil
	}
	return value, nil
}
