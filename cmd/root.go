package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanif/mirusync/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "mirusync",
	Short: "A production-grade folder synchronization tool",
	Long: `mirusync is a CLI tool for synchronizing folders between two Macs over SSH.

It uses rsync for efficient file transfers and provides safety guardrails
for production use.`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Show branding on regular command runs across the CLI.
		if cmd.Name() != "completion" {
			tui.PrintLogo()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mirusync/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".mirusync")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK for some commands (like init)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: Error reading config file: %v\n", err)
		}
	}
}


