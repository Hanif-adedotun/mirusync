package cmd

import (
	"fmt"

	"github.com/hanif/mirusync/internal/engine"
	"github.com/spf13/cobra"
)

var pullDryRun bool
var pullForce bool
var pullVerboseDryRun bool

var pullCmd = &cobra.Command{
	Use:   "pull <folder>",
	Short: "Pull remote folder to local",
	Long: `Pull synchronizes files from the remote host to your local machine.

This is a one-way sync: remote → local.

Example:
  mirusync pull projects`,
	Args: cobra.ExactArgs(1),
	RunE: runPull,
}

func init() {
	pullCmd.Flags().BoolVar(&pullDryRun, "dry-run", false, "Show what would be synced without actually syncing")
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "Override safety checks (not recommended)")
	pullCmd.Flags().BoolVar(&pullVerboseDryRun, "verbose-dry-run", false, "Show raw rsync dry-run output")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	folderName := args[0]

	eng := engine.NewEngine(pullForce, pullVerboseDryRun)
	if err := eng.Pull(folderName, pullDryRun, false); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}


