package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove cc-statusline from Claude Code settings",
	Long:  "Removes the statusLine entry from Claude Code's settings.json.",
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(_ *cobra.Command, _ []string) error {
	settingsPath := claudeSettingsPath()

	settings, err := readSettingsFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No Claude Code settings found. Nothing to uninstall.")
			return nil
		}
		return err
	}

	if _, ok := settings["statusLine"]; !ok {
		fmt.Println("No statusLine entry found. Nothing to uninstall.")
		return nil
	}

	delete(settings, "statusLine")
	if err := writeSettingsFile(settingsPath, settings); err != nil {
		return err
	}
	fmt.Println("Uninstalled cc-statusline.")
	return nil
}
