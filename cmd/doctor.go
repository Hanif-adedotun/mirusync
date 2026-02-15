package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hanif/mirusync/internal/config"
	"github.com/hanif/mirusync/internal/ssh"
	"github.com/hanif/mirusync/internal/validator"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration and connectivity issues",
	Long: `Doctor checks your mirusync configuration and tests connectivity to all configured hosts.

It validates:
  - Configuration file syntax
  - Host definitions
  - Folder definitions
  - SSH connectivity
  - Path accessibility

Example:
  mirusync doctor`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Running diagnostics...")
	fmt.Println()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %w\n\nRun 'mirusync init' to create a configuration file.", err)
	}

	fmt.Println("✓ Configuration file loaded")

	// Validate hosts
	fmt.Println("\n📡 Checking hosts...")
	hostErrors := 0
	for hostName := range cfg.Hosts {
		fmt.Printf("  Checking host '%s'...", hostName)
		if err := validator.ValidateHost(hostName); err != nil {
			fmt.Printf(" ❌\n     Error: %v\n", err)
			hostErrors++
			continue
		}
		fmt.Printf(" ✓\n")

		// Test connectivity
		fmt.Printf("    Testing SSH connectivity...")
		if err := ssh.CheckConnectivity(hostName, 5*time.Second); err != nil {
			fmt.Printf(" ❌\n     Error: %v\n", err)
			hostErrors++
		} else {
			fmt.Printf(" ✓\n")
		}
	}

	if hostErrors > 0 {
		fmt.Printf("\n⚠️  Found %d host error(s)\n", hostErrors)
	} else {
		fmt.Println("\n✓ All hosts are valid and reachable")
	}

	// Validate folders
	fmt.Println("\n📁 Checking folders...")
	folderErrors := 0
	for folderName := range cfg.Folders {
		fmt.Printf("  Checking folder '%s'...", folderName)
		if err := validator.ValidateFolder(folderName); err != nil {
			fmt.Printf(" ❌\n     Error: %v\n", err)
			folderErrors++
		} else {
			fmt.Printf(" ✓\n")
		}
	}

	if folderErrors > 0 {
		fmt.Printf("\n⚠️  Found %d folder error(s)\n", folderErrors)
	} else {
		fmt.Println("\n✓ All folders are valid")
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	if hostErrors == 0 && folderErrors == 0 {
		fmt.Println("✅ All checks passed! Your configuration is ready to use.")
		return nil
	}

	fmt.Printf("⚠️  Found %d error(s) total. Please fix the issues above.\n", hostErrors+folderErrors)
	return fmt.Errorf("diagnostics found errors")
}
