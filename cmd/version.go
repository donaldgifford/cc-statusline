package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build info set by main via SetVersionInfo.
var (
	buildVersion = "dev"
	buildCommit  = "none"
)

// SetVersionInfo sets the version and commit hash for the version command.
func SetVersionInfo(version, commit string) {
	buildVersion = version
	buildCommit = commit
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and build information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("cc-statusline version %s (%s)\n", buildVersion, buildCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
