package cli

import (
	"fmt"

	"github.com/dl-alexandre/App-Store-Server-CLI/internal/validate"
	"github.com/spf13/cobra"
)

func newSubscriptionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription",
		Short:   "Subscription operations",
		Aliases: []string{"sub"},
	}

	cmd.AddCommand(newSubscriptionStatusCommand())
	cmd.AddCommand(newSubscriptionExtendCommand())
	cmd.AddCommand(newSubscriptionExtendMassCommand())
	cmd.AddCommand(newSubscriptionExtendStatusCommand())
	return cmd
}

func newSubscriptionStatusCommand() *cobra.Command {
	var transactionID string
	var statuses []string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get all subscription statuses",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			statuses, err = validate.NormalizeMany("subscription.status", statuses)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			query := map[string][]string{}
			if len(statuses) > 0 {
				query["status"] = statuses
			}
			path := fmt.Sprintf("/inApps/v1/subscriptions/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, query, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "subscription status filter (repeatable)")
	_ = cmd.RegisterFlagCompletionFunc("status", completeEnum("subscription.status"))
	return cmd
}

func newSubscriptionExtendCommand() *cobra.Command {
	var originalTransactionID string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "extend",
		Short: "Extend subscription renewal date",
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
			path := fmt.Sprintf("/inApps/v1/subscriptions/extend/%s", originalTransactionID)
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

func newSubscriptionExtendMassCommand() *cobra.Command {
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "extend-mass",
		Short: "Extend renewal date for all active subscribers",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			path := "/inApps/v1/subscriptions/extend/mass"
			resp, err := app.Client.Do(cmd.Context(), "POST", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

func newSubscriptionExtendStatusCommand() *cobra.Command {
	var requestID string
	var productID string
	cmd := &cobra.Command{
		Use:   "extend-status",
		Short: "Get status of renewal date extension request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if requestID == "" || productID == "" {
				return fmt.Errorf("request-id and product-id are required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/subscriptions/extend/mass/%s/%s", productID, requestID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&requestID, "request-id", "", "request identifier (UUID)")
	cmd.Flags().StringVar(&productID, "product-id", "", "product id")
	return cmd
}
