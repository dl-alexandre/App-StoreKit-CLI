package cli

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/App-Store-Server-CLI/internal/validate"
	"github.com/spf13/cobra"
)

func newTransactionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transaction",
		Short:   "Transaction operations",
		Aliases: []string{"tx"},
	}

	cmd.AddCommand(newTransactionGetCommand())
	cmd.AddCommand(newTransactionHistoryCommand())
	cmd.AddCommand(newTransactionAppCommand())
	cmd.AddCommand(newTransactionAppAccountTokenCommand())
	return cmd
}

func newTransactionGetCommand() *cobra.Command {
	var transactionID string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get transaction info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/transactions/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	return cmd
}

func newTransactionHistoryCommand() *cobra.Command {
	var transactionID string
	var revision string
	var startDate string
	var endDate string
	var productIDs []string
	var productTypes []string
	var sortOrder string
	var subscriptionGroupIDs []string
	var ownershipType string
	var revoked optionalBool
	var version string
	cmd := &cobra.Command{
		Use:     "history",
		Short:   "Get transaction history",
		Aliases: []string{"hist"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			version, err = validate.NormalizeOne("transaction.history.version", version)
			if err != nil {
				return err
			}
			sortOrder, err = validate.NormalizeOne("transaction.history.sort", sortOrder)
			if err != nil {
				return err
			}
			productTypes, err = validate.NormalizeMany("transaction.history.productType", productTypes)
			if err != nil {
				return err
			}
			ownershipType, err = validate.NormalizeOne("transaction.history.ownershipType", ownershipType)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
			}
			if version == "" {
				version = "v2"
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			query := map[string][]string{}
			if revision != "" {
				query["revision"] = []string{revision}
			}
			if startDate != "" {
				query["startDate"] = []string{startDate}
			}
			if endDate != "" {
				query["endDate"] = []string{endDate}
			}
			if len(productIDs) > 0 {
				query["productId"] = productIDs
			}
			if len(productTypes) > 0 {
				query["productType"] = productTypes
			}
			if sortOrder != "" {
				query["sort"] = []string{sortOrder}
			}
			if len(subscriptionGroupIDs) > 0 {
				query["subscriptionGroupIdentifier"] = subscriptionGroupIDs
			}
			if ownershipType != "" {
				query["inAppOwnershipType"] = []string{ownershipType}
			}
			if revoked.value != nil {
				query["revoked"] = []string{fmt.Sprintf("%t", *revoked.value)}
			}
			path := fmt.Sprintf("/inApps/%s/history/%s", version, transactionID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, query, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	cmd.Flags().StringVar(&revision, "revision", "", "revision token")
	cmd.Flags().StringVar(&startDate, "start-date", "", "start date (timestamp in ms)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "end date (timestamp in ms)")
	cmd.Flags().StringArrayVar(&productIDs, "product-id", nil, "product id (repeatable)")
	cmd.Flags().StringArrayVar(&productTypes, "product-type", nil, "product type (repeatable)")
	cmd.Flags().StringVar(&sortOrder, "sort", "", "sort order")
	cmd.Flags().StringArrayVar(&subscriptionGroupIDs, "subscription-group-id", nil, "subscription group identifier (repeatable)")
	cmd.Flags().StringVar(&ownershipType, "ownership-type", "", "in-app ownership type")
	cmd.Flags().Var(&revoked, "revoked", "filter revoked transactions")
	cmd.Flags().StringVar(&version, "version", "v2", "history endpoint version: v1|v2")
	_ = cmd.RegisterFlagCompletionFunc("version", completeEnum("transaction.history.version"))
	_ = cmd.RegisterFlagCompletionFunc("sort", completeEnum("transaction.history.sort"))
	_ = cmd.RegisterFlagCompletionFunc("product-type", completeEnum("transaction.history.productType"))
	_ = cmd.RegisterFlagCompletionFunc("ownership-type", completeEnum("transaction.history.ownershipType"))
	return cmd
}

func newTransactionAppCommand() *cobra.Command {
	var transactionID string
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Get app transaction info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/transactions/appTransactions/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	return cmd
}

func newTransactionAppAccountTokenCommand() *cobra.Command {
	var originalTransactionID string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "app-account-token",
		Short: "Set app account token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if originalTransactionID == "" {
				return fmt.Errorf("original-transaction-id is required")
			}
			body, err := readBody(bodyPath)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return fmt.Errorf("body is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/transactions/%s/appAccountToken", originalTransactionID)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&originalTransactionID, "original-transaction-id", "", "original transaction id")
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

type optionalBool struct {
	value *bool
}

func (o *optionalBool) String() string {
	if o == nil || o.value == nil {
		return ""
	}
	return fmt.Sprintf("%t", *o.value)
}

func (o *optionalBool) Set(value string) error {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return nil
	}
	boolValue := v == "true" || v == "1" || v == "yes"
	o.value = &boolValue
	return nil
}

func (o *optionalBool) Type() string {
	return "bool"
}
