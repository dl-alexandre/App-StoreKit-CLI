package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
	}

	cmd.AddCommand(newCompletionShellCommand("bash"))
	cmd.AddCommand(newCompletionShellCommand("zsh"))
	cmd.AddCommand(newCompletionShellCommand("fish"))
	cmd.AddCommand(newCompletionShellCommand("powershell"))
	return cmd
}

func newCompletionShellCommand(shell string) *cobra.Command {
	return &cobra.Command{
		Use:   shell,
		Short: "Generate " + shell + " completions",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return nil
			}
		},
	}
}
