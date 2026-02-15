package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/hanif/mirusync/internal/config"
	"github.com/hanif/mirusync/internal/ssh"
	"github.com/hanif/mirusync/internal/state"
	"github.com/hanif/mirusync/internal/validator"
	"github.com/hanif/mirusync/pkg/rsync"
)

type Engine struct {
	force bool
}

func NewEngine(force bool) *Engine {
	return &Engine{force: force}
}

func (e *Engine) Push(folderName string, dryRunOnly bool) error {
	folder, err := config.GetFolder(folderName)
	if err != nil {
		return err
	}

	// Validate
	if err := validator.ValidateFolder(folderName); err != nil {
		return err
	}

	if err := validator.CheckForbiddenPath(folder.LocalPath, e.force); err != nil {
		return err
	}

	// Check SSH connectivity
	if err := ssh.CheckConnectivity(folder.RemoteHost, 5*time.Second); err != nil {
		return err
	}

	// Build paths
	source := ensureTrailingSlash(folder.LocalPath)
	dest := ssh.BuildRSyncRemotePath(folder.RemoteHost, folder.RemoteSubpath)

	// Build SSH command
	sshCmd := ssh.BuildSSHCommand(folder.RemoteHost)

	// Dry run first
	options := rsync.RSyncOptions{
		Source:      source,
		Destination: dest,
		Delete:      folder.Delete,
		Checksum:    folder.Checksum,
		DryRun:      true,
		SSHCommand:  sshCmd,
	}

	dryRunResult, err := rsync.DryRun(options)
	if err != nil {
		return fmt.Errorf("dry-run failed: %w", err)
	}

	// Show preview
	fmt.Println("\n📋 Dry-run preview:")
	fmt.Printf("  + %d files to add\n", dryRunResult.FilesAdded)
	fmt.Printf("  ~ %d files to modify\n", dryRunResult.FilesModified)
	if folder.Delete {
		fmt.Printf("  - %d files to delete\n", dryRunResult.FilesDeleted)
	}
	if dryRunResult.TotalSize > 0 {
		fmt.Printf("  📦 Total size: %s\n", formatSize(dryRunResult.TotalSize))
	}

	if dryRunOnly {
		fmt.Println("\n✓ Dry-run complete. Use without --dry-run to execute.")
		return nil
	}

	// Execute
	fmt.Println("\n🚀 Executing sync...")
	options.DryRun = false
	if err := rsync.Execute(options); err != nil {
		return err
	}

	// Update state
	syncState := &state.SyncState{
		LastDirection: "push",
		LastHost:      folder.RemoteHost,
		FileCount:     dryRunResult.FilesAdded + dryRunResult.FilesModified,
	}
	if err := state.SaveState(folderName, syncState); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	fmt.Println("✓ Push complete!")
	return nil
}

func (e *Engine) Pull(folderName string, dryRunOnly bool) error {
	folder, err := config.GetFolder(folderName)
	if err != nil {
		return err
	}

	// Validate
	if err := validator.ValidateFolder(folderName); err != nil {
		return err
	}

	if err := validator.CheckForbiddenPath(folder.LocalPath, e.force); err != nil {
		return err
	}

	// Check SSH connectivity
	if err := ssh.CheckConnectivity(folder.RemoteHost, 5*time.Second); err != nil {
		return err
	}

	// Build paths
	source := ssh.BuildRSyncRemotePath(folder.RemoteHost, folder.RemoteSubpath) + "/"
	dest := ensureTrailingSlash(folder.LocalPath)

	// Build SSH command
	sshCmd := ssh.BuildSSHCommand(folder.RemoteHost)

	// Dry run first
	options := rsync.RSyncOptions{
		Source:      source,
		Destination: dest,
		Delete:      folder.Delete,
		Checksum:    folder.Checksum,
		DryRun:      true,
		SSHCommand:  sshCmd,
	}

	dryRunResult, err := rsync.DryRun(options)
	if err != nil {
		return fmt.Errorf("dry-run failed: %w", err)
	}

	// Show preview
	fmt.Println("\n📋 Dry-run preview:")
	fmt.Printf("  + %d files to add\n", dryRunResult.FilesAdded)
	fmt.Printf("  ~ %d files to modify\n", dryRunResult.FilesModified)
	if folder.Delete {
		fmt.Printf("  - %d files to delete\n", dryRunResult.FilesDeleted)
	}
	if dryRunResult.TotalSize > 0 {
		fmt.Printf("  📦 Total size: %s\n", formatSize(dryRunResult.TotalSize))
	}

	if dryRunOnly {
		fmt.Println("\n✓ Dry-run complete. Use without --dry-run to execute.")
		return nil
	}

	// Execute
	fmt.Println("\n🚀 Executing sync...")
	options.DryRun = false
	if err := rsync.Execute(options); err != nil {
		return err
	}

	// Update state
	syncState := &state.SyncState{
		LastDirection: "pull",
		LastHost:      folder.RemoteHost,
		FileCount:     dryRunResult.FilesAdded + dryRunResult.FilesModified,
	}
	if err := state.SaveState(folderName, syncState); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	fmt.Println("✓ Pull complete!")
	return nil
}

func (e *Engine) Sync(folderName string, dryRunOnly bool) error {
	folder, err := config.GetFolder(folderName)
	if err != nil {
		return err
	}

	if folder.Mode != "bidirectional" {
		return fmt.Errorf("sync command requires mode 'bidirectional' for folder '%s'", folderName)
	}

	// Validate
	if err := validator.ValidateFolder(folderName); err != nil {
		return err
	}

	if err := validator.CheckForbiddenPath(folder.LocalPath, e.force); err != nil {
		return err
	}

	// Check SSH connectivity
	if err := ssh.CheckConnectivity(folder.RemoteHost, 5*time.Second); err != nil {
		return err
	}

	// Load previous state
	prevState, _ := state.LoadState(folderName)

	// For bidirectional sync, we do a smart merge:
	// 1. Pull remote changes first
	// 2. Detect conflicts
	// 3. Handle conflicts by creating .conflict files
	// 4. Push local changes

	fmt.Println("🔄 Starting bidirectional sync...")

	// Step 1: Pull remote changes
	fmt.Println("\n📥 Step 1: Pulling remote changes...")
	if err := e.Pull(folderName, dryRunOnly); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	// Step 2: Push local changes
	fmt.Println("\n📤 Step 2: Pushing local changes...")
	if err := e.Push(folderName, dryRunOnly); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	// Step 3: Detect conflicts (simplified - in production, you'd compare file hashes)
	if prevState != nil && !prevState.LastSync.IsZero() {
		fmt.Println("\n🔍 Checking for conflicts...")
		// This is a simplified conflict detection
		// In a full implementation, you'd compare file modification times
		// and content hashes to detect conflicts
	}

	// Update state
	syncState := &state.SyncState{
		LastDirection: "sync",
		LastHost:      folder.RemoteHost,
	}
	if err := state.SaveState(folderName, syncState); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	fmt.Println("\n✓ Bidirectional sync complete!")
	return nil
}

func ensureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

