package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRefundCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "refund",
		Short:   "Refund operations",
		Aliases: []string{"ref"},
	}

	cmd.AddCommand(newRefundHistoryCommand())
	return cmd
}

func newRefundHistoryCommand() *cobra.Command {
	var transactionID string
	var revision string
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get refund history",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			query := map[string][]string{}
			if revision != "" {
				query["revision"] = []string{revision}
			}
			path := fmt.Sprintf("/inApps/v2/refund/lookup/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, query, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	cmd.Flags().StringVar(&revision, "revision", "", "revision token")
	return cmd
}
