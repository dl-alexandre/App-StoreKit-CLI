package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNotificationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notification",
		Short:   "Notification operations",
		Aliases: []string{"notif"},
	}

	cmd.AddCommand(newNotificationHistoryCommand())
	cmd.AddCommand(newNotificationTestCommand())
	cmd.AddCommand(newNotificationTestStatusCommand())
	return cmd
}

func newNotificationHistoryCommand() *cobra.Command {
	var paginationToken string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get notification history",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			body, err := readBody(bodyPath)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return fmt.Errorf("body is required")
			}
			query := map[string][]string{}
			if paginationToken != "" {
				query["paginationToken"] = []string{paginationToken}
			}
			path := "/inApps/v1/notifications/history"
			resp, err := app.Client.Do(cmd.Context(), "POST", path, query, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	cmd.Flags().StringVar(&paginationToken, "pagination-token", "", "pagination token")
	return cmd
}

func newNotificationTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Request test notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := "/inApps/v1/notifications/test"
			resp, err := app.Client.Do(cmd.Context(), "POST", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	return cmd
}

func newNotificationTestStatusCommand() *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "test-status",
		Short: "Get test notification status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("test-notification-token is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/notifications/test/%s", token)
			resp, err := app.Client.Do(cmd.Context(), "GET", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&token, "test-notification-token", "", "test notification token")
	return cmd
}
