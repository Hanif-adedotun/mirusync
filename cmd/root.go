package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanif/mirusync/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var cfgFile string
var version = "dev"
var noLogo bool

var rootCmd = &cobra.Command{
	Use:   "mirusync",
	Short: "A production-grade folder synchronization tool",
	Long: `mirusync is a CLI tool for synchronizing folders between two Macs over SSH.

It uses rsync for efficient file transfers and provides safety guardrails
for production use.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := maybePrintLogo(cmd); err != nil {
			return err
		}
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mirusync/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noLogo, "no-logo", false, "omit the banner (useful for scripts and non-interactive use)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Bare `mirusync` uses RunE (logo there); avoid printing twice for subcommands only when cmd is root.
		if cmd == rootCmd {
			return nil
		}
		return maybePrintLogo(cmd)
	}
}

func maybePrintLogo(cmd *cobra.Command) error {
	if noLogo {
		return nil
	}
	// Shell completion output must stay machine-parseable.
	if cmd.Name() == "completion" {
		return nil
	}
	if cmd.Flag("version") != nil && cmd.Flag("version").Changed {
		return nil
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil
	}
	tui.PrintLogo()
	return nil
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
