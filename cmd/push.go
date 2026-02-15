package cmd

import (
	"fmt"

	"github.com/hanif/mirusync/internal/engine"
	"github.com/spf13/cobra"
)

var pushDryRun bool
var pushForce bool

var pushCmd = &cobra.Command{
	Use:   "push <folder>",
	Short: "Push local folder to remote",
	Long: `Push synchronizes files from your local machine to the remote host.

This is a one-way sync: local → remote.

Example:
  mirusync push projects`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func init() {
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Show what would be synced without actually syncing")
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Override safety checks (not recommended)")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	folderName := args[0]

	eng := engine.NewEngine(pushForce)
	if err := eng.Push(folderName, pushDryRun); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

