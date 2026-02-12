package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMessagingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messaging",
		Short:   "Retention messaging operations",
		Aliases: []string{"msg"},
	}

	cmd.AddCommand(newMessagingImageCommand())
	cmd.AddCommand(newMessagingMessageCommand())
	cmd.AddCommand(newMessagingDefaultCommand())
	return cmd
}

func newMessagingImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Image operations",
	}
	cmd.AddCommand(newMessagingImageListCommand())
	cmd.AddCommand(newMessagingImageUploadCommand())
	cmd.AddCommand(newMessagingImageDeleteCommand())
	return cmd
}

func newMessagingImageListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			resp, err := app.Client.Do(cmd.Context(), "GET", "/inApps/v1/messaging/image/list", nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	return cmd
}

func newMessagingImageUploadCommand() *cobra.Command {
	var imageID string
	var filePath string
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if imageID == "" {
				return fmt.Errorf("image-id is required")
			}
			if filePath == "" {
				return fmt.Errorf("file is required")
			}
			image, err := readBody(filePath)
			if err != nil {
				return err
			}
			if len(image) == 0 {
				return fmt.Errorf("file is empty")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/messaging/image/%s", imageID)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, image, "image/png")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&imageID, "image-id", "", "image identifier (UUID)")
	cmd.Flags().StringVar(&filePath, "file", "", "png file path")
	return cmd
}

func newMessagingImageDeleteCommand() *cobra.Command {
	var imageID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if imageID == "" {
				return fmt.Errorf("image-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/messaging/image/%s", imageID)
			resp, err := app.Client.Do(cmd.Context(), "DELETE", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&imageID, "image-id", "", "image identifier (UUID)")
	return cmd
}

func newMessagingMessageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "Message operations",
	}
	cmd.AddCommand(newMessagingMessageListCommand())
	cmd.AddCommand(newMessagingMessageUploadCommand())
	cmd.AddCommand(newMessagingMessageDeleteCommand())
	return cmd
}

func newMessagingMessageListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			resp, err := app.Client.Do(cmd.Context(), "GET", "/inApps/v1/messaging/message/list", nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	return cmd
}

func newMessagingMessageUploadCommand() *cobra.Command {
	var messageID string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageID == "" {
				return fmt.Errorf("message-id is required")
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
			path := fmt.Sprintf("/inApps/v1/messaging/message/%s", messageID)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&messageID, "message-id", "", "message identifier (UUID)")
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

func newMessagingMessageDeleteCommand() *cobra.Command {
	var messageID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if messageID == "" {
				return fmt.Errorf("message-id is required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/messaging/message/%s", messageID)
			resp, err := app.Client.Do(cmd.Context(), "DELETE", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&messageID, "message-id", "", "message identifier (UUID)")
	return cmd
}

func newMessagingDefaultCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "Default message configuration",
	}
	cmd.AddCommand(newMessagingDefaultConfigureCommand())
	cmd.AddCommand(newMessagingDefaultDeleteCommand())
	return cmd
}

func newMessagingDefaultConfigureCommand() *cobra.Command {
	var productID string
	var locale string
	var bodyPath string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure default message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID == "" || locale == "" {
				return fmt.Errorf("product-id and locale are required")
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
			path := fmt.Sprintf("/inApps/v1/messaging/default/%s/%s", productID, locale)
			resp, err := app.Client.Do(cmd.Context(), "PUT", path, nil, body, "application/json")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "product id")
	cmd.Flags().StringVar(&locale, "locale", "", "locale")
	cmd.Flags().StringVar(&bodyPath, "body", "", "json body file path (use - for stdin)")
	return cmd
}

func newMessagingDefaultDeleteCommand() *cobra.Command {
	var productID string
	var locale string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete default message",
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID == "" || locale == "" {
				return fmt.Errorf("product-id and locale are required")
			}
			app, err := appOrExit(cmd)
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/inApps/v1/messaging/default/%s/%s", productID, locale)
			resp, err := app.Client.Do(cmd.Context(), "DELETE", path, nil, nil, "")
			if err != nil {
				return err
			}
			return writeResponse(app, responseData(app, resp))
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "product id")
	cmd.Flags().StringVar(&locale, "locale", "", "locale")
	return cmd
}
