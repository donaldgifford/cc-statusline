// Package cmd is the entrypoint to the cc-statusline cli.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/config"
	"github.com/donaldgifford/cc-statusline/internal/statusline"
)

var (
	cfgFile              string
	noColor              bool
	experimentalJSONL    bool
	experimentalUsageAPI bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "cc-statusline",
	Short: "A statusline for Claude Code",
	Long:  "cc-statusline reads session data from Claude Code via stdin and renders a configurable statusline.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfgPath := cfgFile
		if cfgPath == "" {
			cfgPath = config.DefaultPath()
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// CLI flags override config file values.
		if noColor {
			f := false
			cfg.Color = &f
		}
		if experimentalJSONL {
			cfg.Experimental.JSONL = true
		}
		if experimentalUsageAPI {
			cfg.Experimental.UsageAPI = true
		}

		// Set color state once before rendering.
		enabled := cfg.ColorEnabled()
		color.SetEnabled(&enabled)

		return statusline.RunWithConfig(cmd.InOrStdin(), cmd.OutOrStdout(), cfg)
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
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&experimentalJSONL, "experimental-jsonl", false, "enable experimental JSONL transcript segments")
	rootCmd.PersistentFlags().BoolVar(&experimentalUsageAPI, "experimental-usage-api", false, "enable experimental usage API segments")
}
