package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExternalPurchaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external",
		Short: "External purchase reporting",
	}

	cmd.AddCommand(newExternalSendCommand())
	cmd.AddCommand(newExternalGetCommand())
	return cmd
}

func newExternalSendCommand() *cobra.Command {
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send external purchase report",
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
			if app.External.BaseURL == "" {
				return fmt.Errorf("external purchase is unsupported for env: %s", app.Config.Environment)
			}
			path := "/externalPurchase/v1/reports"
			resp, err := app.External.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

func newExternalGetCommand() *cobra.Command {
	var requestID string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve external purchase report",
		RunE: func(cmd *cobra.Command, args []string) error {
			if requestID == "" {
				return fmt.Errorf("request-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			if app.External.BaseURL == "" {
				return fmt.Errorf("external purchase is unsupported for env: %s", app.Config.Environment)
			}
			path := fmt.Sprintf("/externalPurchase/v1/reports/%s", requestID)
			resp, err := app.External.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&requestID, "request-id", "", "request identifier (UUID)")
	return cmd
}
