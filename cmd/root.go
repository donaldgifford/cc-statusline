// Package cmd is the entrypoint to the cc-statusline cli.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/donaldgifford/cc-statusline/internal/statusline"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "cc-statusline",
	Short: "A statusline for Claude Code",
	Long:  "cc-statusline reads session data from Claude Code via stdin and renders a configurable statusline.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return statusline.Run(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cc-statusline.yaml)")
}
