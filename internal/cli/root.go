package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/api"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/auth"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/config"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/output"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/retry"
	"github.com/spf13/cobra"
)

type App struct {
	Config    config.Config
	Client    api.Client
	External  api.Client
	Format    output.Format
	JQ        string
	RequestID string
	Timeout   time.Duration
	Columns   []string
	Debug     bool
}

type rootOptions struct {
	ConfigPath string
	Format     string
	JQ         string
	RequestID  string
	Timeout    time.Duration
	Columns    []string
	Debug      bool

	IssuerID       string
	KeyID          string
	BundleID       string
	PrivateKeyPath string
	PrivateKey     string
	Environment    string
	MaxRetries     int
	RetryBackoffMS int
}

func Execute() int {
	// Perform automatic update check on startup (non-blocking)
	AutoUpdateCheck()

	root := newRootCommand()
	if err := root.Execute(); err != nil {
		writeError(err)
		return exitCode(err)
	}
	return exitSuccess
}

func newRootCommand() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:           "ask",
		Short:         "CLI for the App Store Server API",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if shouldSkipAppInit(cmd) {
				return nil
			}
			app, err := buildApp(cmd.Context(), opts)
			if err != nil {
				return err
			}
			cmd.SetContext(withApp(cmd.Context(), app))
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "json", "output format: json|table|raw")
	cmd.PersistentFlags().StringVar(&opts.JQ, "jq", "", "jq expression")
	cmd.PersistentFlags().StringSliceVar(&opts.Columns, "table-columns", nil, "table columns (comma-separated)")
	cmd.PersistentFlags().StringVar(&opts.RequestID, "request-id", "", "request id header")
	cmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "request timeout")
	cmd.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "enable debug output")

	cmd.PersistentFlags().StringVar(&opts.IssuerID, "issuer-id", "", "issuer id")
	cmd.PersistentFlags().StringVar(&opts.KeyID, "key-id", "", "key id")
	cmd.PersistentFlags().StringVar(&opts.BundleID, "bundle-id", "", "bundle id")
	cmd.PersistentFlags().StringVar(&opts.PrivateKeyPath, "private-key-path", "", "path to .p8 file")
	cmd.PersistentFlags().StringVar(&opts.PrivateKey, "private-key", "", "private key contents")
	cmd.PersistentFlags().StringVar(&opts.Environment, "env", "", "environment: sandbox|production|local-testing")
	cmd.PersistentFlags().IntVar(&opts.MaxRetries, "max-retries", 0, "max retries")
	cmd.PersistentFlags().IntVar(&opts.RetryBackoffMS, "retry-backoff-ms", 0, "retry backoff (ms)")

	cmd.AddCommand(newConfigCommand())
	cmd.AddCommand(newTransactionCommand())
	cmd.AddCommand(newNotificationCommand())
	cmd.AddCommand(newRefundCommand())
	cmd.AddCommand(newSubscriptionCommand())
	cmd.AddCommand(newOrderCommand())
	cmd.AddCommand(newMessagingCommand())
	cmd.AddCommand(newConsumptionCommand())
	cmd.AddCommand(newExternalPurchaseCommand())
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newCheckUpdateCommand())

	return cmd
}

func buildApp(ctx context.Context, opts *rootOptions) (*App, error) {
	cfg, err := config.Load(config.Options{
		ConfigPath: opts.ConfigPath,
		Overrides: config.Config{
			IssuerID:       opts.IssuerID,
			KeyID:          opts.KeyID,
			BundleID:       opts.BundleID,
			PrivateKeyPath: opts.PrivateKeyPath,
			PrivateKey:     opts.PrivateKey,
			Environment:    opts.Environment,
			MaxRetries:     opts.MaxRetries,
			RetryBackoffMS: opts.RetryBackoffMS,
			RequestTimeout: opts.Timeout,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	signer, err := auth.NewSigner(cfg.IssuerID, cfg.KeyID, cfg.BundleID, cfg.PrivateKeyPath, cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	baseURL, err := resolveBaseURL(cfg.Environment)
	if err != nil {
		return nil, err
	}

	timeout := cfg.RequestTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}

	app := &App{
		Config:    cfg,
		Format:    output.Format(opts.Format),
		JQ:        opts.JQ,
		RequestID: opts.RequestID,
		Timeout:   timeout,
		Columns:   opts.Columns,
		Debug:     opts.Debug,
	}

	app.Client = api.Client{
		BaseURL: baseURL,
		HTTP:    httpClient,
		Signer:  signer,
		Retry: retry.Config{
			MaxRetries: cfg.MaxRetries,
			Backoff:    cfg.RetryBackoff,
		},
		RequestID: opts.RequestID,
		UserAgent: "ask/0.1.0",
		Debug:     opts.Debug,
		DebugOut:  os.Stderr,
	}

	externalBaseURL, err := resolveExternalPurchaseBaseURL(cfg.Environment)
	if err != nil {
		return nil, err
	}
	app.External = api.Client{
		BaseURL: externalBaseURL,
		HTTP:    httpClient,
		Signer:  signer,
		Retry: retry.Config{
			MaxRetries: cfg.MaxRetries,
			Backoff:    cfg.RetryBackoff,
		},
		RequestID: opts.RequestID,
		UserAgent: "ask/0.1.0",
		Debug:     opts.Debug,
		DebugOut:  os.Stderr,
	}

	return app, nil
}

func resolveBaseURL(env string) (string, error) {
	switch env {
	case "sandbox":
		return "https://api.storekit-sandbox.itunes.apple.com", nil
	case "local-testing":
		return "https://local-testing-base-url", nil
	case "production", "":
		return "https://api.storekit.itunes.apple.com", nil
	default:
		return "", fmt.Errorf("unsupported environment: %s", env)
	}
}

func resolveExternalPurchaseBaseURL(env string) (string, error) {
	switch env {
	case "sandbox":
		return "https://api.storekit-sandbox.apple.com", nil
	case "production", "":
		return "https://api.storekit.apple.com", nil
	case "local-testing":
		return "", nil
	default:
		return "", fmt.Errorf("unsupported environment: %s", env)
	}
}

func appOrExit(cmd *cobra.Command) (*App, error) {
	app := appFromContext(cmd.Context())
	if app == nil {
		return nil, errors.New("app context not initialized")
	}
	return app, nil
}

func shouldSkipAppInit(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	return strings.HasPrefix(cmd.CommandPath(), "ask config")
}

func writeError(err error) {
	message := formatError(err)
	if message == "" {
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, message)
}
