package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hanif/mirusync/internal/config"
	"github.com/hanif/mirusync/internal/prompt"
	"github.com/hanif/mirusync/internal/ssh"
	"github.com/hanif/mirusync/internal/tui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up mirusync (interactive)",
	Long: `Run an interactive setup to connect this machine to another and choose a folder to sync.

mirusync will ask for the other machine's address and user, help you set up SSH access,
verify the connection, then ask which folder to sync. No config editing required.`,
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

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	if _, err := os.Stat(configFile); err == nil {
		fmt.Println()
		overwrite, err := prompt.ConfirmStyled("Configuration already exists. Start fresh and overwrite?", false, prompt.Chevron, true)
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println()
			fmt.Println("  Keeping existing config. Edit ~/.mirusync/config.yaml to change settings.")
			return nil
		}
	}

	fmt.Println()
	tui.PrintLogo()

	// --- Other machine ---
	remoteHost, err := prompt.StringStyled("Other machine IP or hostname", "", prompt.Chevron, true)
	if err != nil {
		return err
	}
	if remoteHost == "" {
		return fmt.Errorf("host is required")
	}

	remoteUser, err := prompt.StringStyled("Username on the other machine", "", prompt.Chevron, true)
	if err != nil {
		return err
	}
	if remoteUser == "" {
		return fmt.Errorf("username is required")
	}

	sshPort, err := prompt.IntStyled("SSH port on the other machine", 22, prompt.Chevron, true)
	if err != nil {
		return err
	}

	// --- SSH key ---
	// Use a dedicated, passphrase-less key for mirusync so we don't prompt for a key
	// passphrase every time. The user will still authenticate to the remote machine
	// with their laptop password when ssh-copy-id runs.
	keyPath, keyContent, err := ssh.EnsureMirusyncKey()
	if err != nil {
		fmt.Println()
		fmt.Println("  Failed to prepare mirusync SSH key.")
		fmt.Println("  You can create one manually with:")
		fmt.Println("    ssh-keygen -t ed25519 -N \"\" -f ~/.ssh/id_mirusync")
		fmt.Println()
		return err
	}

	fmt.Println()
	fmt.Println("  Your SSH public key (add this to the other machine if you haven't already):")
	fmt.Println()
	fmt.Println("  " + strings.TrimSpace(string(keyContent)))
	fmt.Println("  (from " + keyPath + ")")
	fmt.Println()

	tryCopyID, err := prompt.ConfirmStyled("Try to install this key on the other machine now (ssh-copy-id)?", true, prompt.Chevron, true)
	if err != nil {
		return err
	}
	if tryCopyID {
		fmt.Println("  Running ssh-copy-id (you may need to enter the other machine's password)...")
		if err := ssh.CopyID(keyPath, remoteUser, remoteHost, sshPort); err != nil {
			fmt.Printf("  Warning: %v\n", err)
			prompt.Pause("  Add the key manually to the other machine, then continue here.")
		} else {
			fmt.Println("  Key installed successfully.")
		}
	} else {
		prompt.Pause("  Add the key to the other machine (e.g. append to ~/.ssh/authorized_keys), then continue here.")
	}

	fmt.Println("  Checking connection to the other machine...")
	if err := ssh.CheckConnectivityRaw(remoteUser, remoteHost, sshPort, 10*time.Second); err != nil {
		return fmt.Errorf("could not connect: %w\n  Fix SSH access and run 'mirusync init' again", err)
	}
	fmt.Println("  Connection OK.")
	fmt.Println()

	// --- Host name ---
	hostName, err := prompt.StringStyled("Name for this host in config (e.g. laptopB)", "remote", prompt.Chevron, true)
	if err != nil {
		return err
	}
	if hostName == "" {
		hostName = "remote"
	}

	remoteBasePath, err := prompt.StringStyled("Base path on the other machine (e.g. /Users/you/dev)", "", prompt.Chevron, true)
	if err != nil {
		return err
	}
	if remoteBasePath == "" {
		return fmt.Errorf("remote base path is required")
	}
	remoteBasePath = strings.TrimSuffix(remoteBasePath, "/")

	// --- Folder to sync ---
	defaultLocal := filepath.Join(home, "dev", "projects")
	localPath, err := prompt.StringStyled("Local folder to sync (this machine)", defaultLocal, prompt.Chevron, true)
	if err != nil {
		return err
	}
	if localPath == "" {
		localPath = defaultLocal
	}
	if strings.HasPrefix(localPath, "~") {
		localPath = filepath.Join(home, strings.TrimPrefix(localPath, "~"))
	}
	localPath, err = filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	// Ensure local dir exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		create, err := prompt.ConfirmStyled("Local folder doesn't exist. Create it?", true, prompt.Chevron, true)
		if err != nil {
			return err
		}
		if create {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return fmt.Errorf("failed to create folder: %w", err)
			}
		}
	}

	defaultSubpath := filepath.Base(localPath)
	remoteSubpath, err := prompt.StringStyled("Path on the other machine (under base path)", defaultSubpath, prompt.Chevron, true)
	if err != nil {
		return err
	}
	if remoteSubpath == "" {
		remoteSubpath = defaultSubpath
	}
	remoteSubpath = strings.Trim(remoteSubpath, "/")

	modeOptions := []string{"Push only (this → other)", "Pull only (other → this)", "Both (sync both ways)"}
	modeIdx, err := prompt.SelectStyled("Sync direction", modeOptions, 2, prompt.Chevron, true)
	if err != nil {
		return err
	}
	var mode string
	switch modeIdx {
	case 0:
		mode = "push"
	case 1:
		mode = "pull"
	default:
		mode = "bidirectional"
	}

	folderName := filepath.Base(localPath)
	if folderName == "." || folderName == "/" {
		folderName = "projects"
	}
	folderName, err = prompt.StringStyled("Name for this folder in mirusync", folderName, prompt.Chevron, true)
	if err != nil {
		return err
	}
	if folderName == "" {
		folderName = "projects"
	}

	// Write config
	cfg := &config.Config{
		Hosts: map[string]config.Host{
			hostName: {
				User:     remoteUser,
				Host:     remoteHost,
				Port:     sshPort,
				BasePath: remoteBasePath,
			},
		},
		Folders: map[string]config.Folder{
			folderName: {
				LocalPath:     localPath,
				RemoteHost:    hostName,
				RemoteSubpath: remoteSubpath,
				Mode:          mode,
				Delete:        false,
				Checksum:      false,
			},
		},
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("  ✓ Setup complete. Config saved to " + configFile)
	fmt.Println()
	fmt.Println("  Next:")
	fmt.Printf("    mirusync push %s   — send your folder to the other machine\n", folderName)
	fmt.Printf("    mirusync pull %s   — get the folder from the other machine\n", folderName)
	if mode == "bidirectional" {
		fmt.Printf("    mirusync sync %s   — sync both ways\n", folderName)
	}
	fmt.Println()

	return nil
}
