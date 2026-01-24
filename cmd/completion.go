package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for craft-cli.

To load completions:

Bash:
  $ source <(craft completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ craft completion bash > /etc/bash_completion.d/craft
  # macOS:
  $ craft completion bash > $(brew --prefix)/etc/bash_completion.d/craft

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ craft completion zsh > "${fpath[1]}/_craft"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ craft completion fish | source
  # To load completions for each session, execute once:
  $ craft completion fish > ~/.config/fish/completions/craft.fish

PowerShell:
  PS> craft completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> craft completion powershell > craft.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
