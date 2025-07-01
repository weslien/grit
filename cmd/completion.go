package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate completion script for your shell.

To load completions:

Bash:
  $ source <(grit completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ grit completion bash > /etc/bash_completion.d/grit
  # macOS:
  $ grit completion bash > $(brew --prefix)/etc/bash_completion.d/grit

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ grit completion zsh > "${fpath[1]}/_grit"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ grit completion fish | source

  # To load completions for each session, execute once:
  $ grit completion fish > ~/.config/fish/completions/grit.fish

PowerShell:
  PS> grit completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> grit completion powershell > grit.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			fmt.Fprintf(os.Stderr, "Unknown shell: %s\n", args[0])
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}