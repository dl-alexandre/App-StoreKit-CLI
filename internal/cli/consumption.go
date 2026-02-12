package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConsumptionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consumption",
		Short:   "Consumption information",
		Aliases: []string{"cons"},
	}

	cmd.AddCommand(newConsumptionSendCommand())
	cmd.AddCommand(newConsumptionSendV1Command())
	return cmd
}

func newConsumptionSendCommand() *cobra.Command {
	var transactionID string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send consumption information (v2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
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
			path := fmt.Sprintf("/inApps/v2/transactions/consumption/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

func newConsumptionSendV1Command() *cobra.Command {
	var transactionID string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "send-v1",
		Short: "Send consumption information (v1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transactionID == "" {
				return fmt.Errorf("transaction-id is required")
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
			path := fmt.Sprintf("/inApps/v1/transactions/consumption/%s", transactionID)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction id")
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}
