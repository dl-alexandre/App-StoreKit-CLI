package cli

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestKongParsesAndDispatchesCommandFlags(t *testing.T) {
	root := newCommand()
	root.Use = "ask"
	var verbose bool
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")

	var endpoint string
	var products []string
	var timeout time.Duration
	dispatched := false
	leaf := &Command{
		Use:     "lookup",
		Aliases: []string{"get"},
		RunE: func(_ *Command, _ []string) error {
			dispatched = true
			if endpoint != "/reports/42" {
				return fmt.Errorf("endpoint = %q", endpoint)
			}
			if len(products) != 2 || products[0] != "one,two" || products[1] != "three" {
				return fmt.Errorf("products = %#v", products)
			}
			if timeout != 5*time.Second {
				return fmt.Errorf("timeout = %s", timeout)
			}
			return nil
		},
	}
	leaf.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	leaf.Flags().StringArrayVar(&products, "product-id", nil, "repeatable product identifier")
	leaf.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "request timeout")
	root.AddCommand(leaf)
	root.SetArgs([]string{"--verbose", "get", "--endpoint", "/reports/42", "--product-id", "one,two", "--product-id", "three"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !verbose {
		t.Fatal("global --verbose flag was not parsed")
	}
	if !dispatched {
		t.Fatal("selected command handler was not dispatched")
	}
}

func TestKongRejectsFlagFromAnotherCommand(t *testing.T) {
	root := newCommand()
	root.Use = "ask"
	first := &Command{Use: "first", RunE: func(*Command, []string) error { return nil }}
	var firstOnly string
	first.Flags().StringVar(&firstOnly, "first-only", "", "first command flag")
	second := &Command{Use: "second", RunE: func(*Command, []string) error { return nil }}
	root.AddCommand(first, second)
	root.SetArgs([]string{"second", "--first-only", "no"})

	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "first-only") {
		t.Fatalf("Execute() error = %v, want unknown first-only flag", err)
	}
}

func TestKongPreservesLocalShadowOfGlobalFlag(t *testing.T) {
	root := newCommand()
	root.Use = "ask"
	var globalRequestID string
	root.PersistentFlags().StringVar(&globalRequestID, "request-id", "", "request id header")
	var reportRequestID string
	dispatched := false
	get := &Command{
		Use: "get",
		RunE: func(*Command, []string) error {
			dispatched = reportRequestID == "report-123"
			return nil
		},
	}
	get.Flags().StringVar(&reportRequestID, "request-id", "", "report id")
	root.AddCommand(get)
	root.SetArgs([]string{"get", "--request-id", "report-123"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !dispatched || globalRequestID != "" {
		t.Fatalf("local request-id shadow not preserved: local=%q global=%q", reportRequestID, globalRequestID)
	}
}

func TestKongKeepsGlobalFlagForSiblingCommand(t *testing.T) {
	root := newCommand()
	root.Use = "ask"
	var globalRequestID string
	root.PersistentFlags().StringVar(&globalRequestID, "request-id", "", "request id header")
	shadow := &Command{Use: "get", RunE: func(*Command, []string) error { return nil }}
	var localRequestID string
	shadow.Flags().StringVar(&localRequestID, "request-id", "", "report id")
	dispatched := false
	sibling := &Command{Use: "history", RunE: func(*Command, []string) error {
		dispatched = globalRequestID == "header-123"
		return nil
	}}
	root.AddCommand(shadow, sibling)
	root.SetArgs([]string{"history", "--request-id", "header-123"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !dispatched {
		t.Fatalf("global request-id was not available: got %q", globalRequestID)
	}
}
