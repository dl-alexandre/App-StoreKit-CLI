package cli

import (
	"os"
)

func newCompletionCommand() *Command {
	cmd := &Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
	}

	cmd.AddCommand(newCompletionShellCommand("bash"))
	cmd.AddCommand(newCompletionShellCommand("zsh"))
	cmd.AddCommand(newCompletionShellCommand("fish"))
	cmd.AddCommand(newCompletionShellCommand("powershell"))
	return cmd
}

func newCompletionShellCommand(shell string) *Command {
	return &Command{
		Use:   shell,
		Short: "Generate " + shell + " completions",
		RunE: func(cmd *Command, args []string) error {
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
