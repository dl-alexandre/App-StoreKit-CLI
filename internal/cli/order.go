package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newOrderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "order",
		Short:   "Order operations",
		Aliases: []string{"ord"},
	}

	cmd.AddCommand(newOrderLookupCommand())
	return cmd
}

func newOrderLookupCommand() *cobra.Command {
	var orderID string
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up order id",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orderID == "" {
				return fmt.Errorf("order-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/lookup/%s", orderID)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&orderID, "order-id", "", "order id")
	return cmd
}
