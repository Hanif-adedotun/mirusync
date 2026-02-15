package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mirusync configuration",
	Long: `Initialize mirusync by creating the configuration directory and a sample config file.

This will create ~/.mirusync/config.yaml with a template configuration.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".mirusync")
	stateDir := filepath.Join(configDir, "state")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("Configuration file already exists at %s\n", configFile)
		fmt.Println("Skipping template creation. Edit the file to customize your setup.")
		return nil
	}

	// Create sample config
	sampleConfig := `# mirusync configuration
# This file defines your hosts and folders to sync

hosts:
  laptopB:
    user: hanif
    host: 192.168.1.23
    port: 22
    base_path: /Users/hanif/dev

folders:
  projects:
    local_path: /Users/hanif/dev/projects
    remote_host: laptopB
    remote_subpath: projects
    mode: bidirectional
    delete: false
    checksum: false  # Use checksums for comparison (slower but safer)
`

	if err := os.WriteFile(configFile, []byte(sampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Created configuration directory: %s\n", configDir)
	fmt.Printf("✓ Created state directory: %s\n", stateDir)
	fmt.Printf("✓ Created sample configuration: %s\n", configFile)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the config file to match your setup")
	fmt.Println("2. Run 'mirusync doctor' to verify your configuration")
	fmt.Println("3. Run 'mirusync push <folder>' or 'mirusync pull <folder>' to sync")

	return nil
}

