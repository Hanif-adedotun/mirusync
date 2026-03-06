package cmd

import (
	"fmt"

	"github.com/hanif/mirusync/internal/engine"
	"github.com/spf13/cobra"
)

var syncDryRun bool
var syncForce bool

var syncCmd = &cobra.Command{
	Use:   "sync <folder>",
	Short: "Bidirectional sync between local and remote",
	Long: `Sync performs a bidirectional synchronization between local and remote.

This command requires the folder to be configured with mode: 'bidirectional'.

The sync process:
  1. Pulls remote changes
  2. Pushes local changes
  3. Detects and reports conflicts

Example:
  mirusync sync projects`,
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", true, "Show what would be synced without actually syncing (default: true)")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Override safety checks (not recommended)")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	folderName := args[0]

	eng := engine.NewEngine(syncForce)
	if err := eng.Sync(folderName, syncDryRun); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}


