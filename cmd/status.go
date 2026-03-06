package cmd

import (
	"fmt"
	"time"

	"github.com/hanif/mirusync/internal/config"
	"github.com/hanif/mirusync/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [folder]",
	Short: "Show sync status for folder(s)",
	Long: `Show the last sync status for a folder, or all folders if no folder is specified.

Example:
  mirusync status projects
  mirusync status`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(args) == 0 {
		// Show status for all folders
		fmt.Println("📊 Sync Status for all folders:\n")
		for folderName := range cfg.Folders {
			if err := showFolderStatus(folderName); err != nil {
				fmt.Printf("Error showing status for %s: %v\n", folderName, err)
			}
			fmt.Println()
		}
		return nil
	}

	// Show status for specific folder
	folderName := args[0]
	return showFolderStatus(folderName)
}

func showFolderStatus(folderName string) error {
	folder, err := config.GetFolder(folderName)
	if err != nil {
		return err
	}

	fmt.Printf("📁 Folder: %s\n", folderName)
	fmt.Printf("   Local path: %s\n", folder.LocalPath)
	fmt.Printf("   Remote: %s\n", folder.RemoteHost)
	fmt.Printf("   Mode: %s\n", folder.Mode)
	fmt.Printf("   Delete enabled: %v\n", folder.Delete)
	fmt.Printf("   Checksum enabled: %v\n", folder.Checksum)

	state, err := state.LoadState(folderName)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.LastSync.IsZero() {
		fmt.Println("   Last sync: Never")
	} else {
		timeSince := time.Since(state.LastSync)
		fmt.Printf("   Last sync: %s (%s ago)\n",
			state.LastSync.Format("2006-01-02 15:04:05"),
			formatDuration(timeSince))
		fmt.Printf("   Last direction: %s\n", state.LastDirection)
		fmt.Printf("   Last host: %s\n", state.LastHost)
		if state.FileCount > 0 {
			fmt.Printf("   Files synced: %d\n", state.FileCount)
		}
		if len(state.Conflicts) > 0 {
			fmt.Printf("   ⚠️  Conflicts: %d\n", len(state.Conflicts))
			for _, conflict := range state.Conflicts {
				fmt.Printf("      - %s\n", conflict)
			}
		}
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}


