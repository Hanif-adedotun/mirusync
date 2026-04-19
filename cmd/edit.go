package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hanif/mirusync/internal/config"
	"github.com/hanif/mirusync/internal/prompt"
	"github.com/hanif/mirusync/internal/tui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <folder>",
	Short: "Edit an existing folder config interactively",
	Long: `Interactively edit an existing mirusync folder configuration.

The editor shows numbered options, current values, and asks for confirmation
before saving each change.`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	folderName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	folder, ok := cfg.Folders[folderName]
	if !ok {
		return fmt.Errorf("folder '%s' not found in configuration", folderName)
	}

	fmt.Printf("%s%s%s\n", tui.ColorMagenta, "Settings Editor", tui.ColorReset)
	fmt.Printf("%sEditing folder%s '%s'\n", tui.ColorCyan, tui.ColorReset, folderName)

	for {
		host, hostExists := cfg.Hosts[folder.RemoteHost]
		hostLabel := folder.RemoteHost
		if !hostExists {
			hostLabel = folder.RemoteHost + " (missing)"
		}

		options := []string{
			fmt.Sprintf("%sFolder%s local_path (%s)", tui.ColorCyan, tui.ColorReset, folder.LocalPath),
			fmt.Sprintf("%sFolder%s remote_host (%s)", tui.ColorCyan, tui.ColorReset, folder.RemoteHost),
			fmt.Sprintf("%sFolder%s remote_subpath (%s)", tui.ColorCyan, tui.ColorReset, folder.RemoteSubpath),
			fmt.Sprintf("%sFolder%s mode (%s)", tui.ColorCyan, tui.ColorReset, folder.Mode),
			fmt.Sprintf("%sFolder%s delete (%t)", tui.ColorCyan, tui.ColorReset, folder.Delete),
			fmt.Sprintf("%sFolder%s checksum (%t)", tui.ColorCyan, tui.ColorReset, folder.Checksum),
			fmt.Sprintf("%sHost%s user [%s] (%s)", tui.ColorMagenta, tui.ColorReset, hostLabel, host.User),
			fmt.Sprintf("%sHost%s host [%s] (%s)", tui.ColorMagenta, tui.ColorReset, hostLabel, host.Host),
			fmt.Sprintf("%sHost%s port [%s] (%d)", tui.ColorMagenta, tui.ColorReset, hostLabel, host.Port),
			fmt.Sprintf("%sHost%s base_path [%s] (%s)", tui.ColorMagenta, tui.ColorReset, hostLabel, host.BasePath),
			fmt.Sprintf("%sExit editor%s", tui.ColorDim, tui.ColorReset),
		}

		idx, err := prompt.SelectStyled("Select a field to edit", options, 0, prompt.Chevron, true)
		if err != nil {
			fmt.Printf("%sInvalid selection:%s %v\n", tui.ColorMagenta, tui.ColorReset, err)
			continue
		}

		if idx == len(options)-1 {
			fmt.Printf("%sNo further changes.%s\n", tui.ColorDim, tui.ColorReset)
			return nil
		}

		changed, err := applyEditSelection(cfg, folderName, idx)
		if err != nil {
			fmt.Printf("%sEdit cancelled:%s %v\n", tui.ColorMagenta, tui.ColorReset, err)
			continue
		}
		if !changed {
			fmt.Printf("%sNo changes made.%s\n", tui.ColorDim, tui.ColorReset)
			continue
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Refresh local copy from map for next iteration.
		folder = cfg.Folders[folderName]
		fmt.Printf("%sChange saved.%s\n", tui.ColorCyan, tui.ColorReset)
	}
}

func applyEditSelection(cfg *config.Config, folderName string, idx int) (bool, error) {
	folder := cfg.Folders[folderName]

	switch idx {
	case 0:
		newValue, err := prompt.StringStyled("New local_path", folder.LocalPath, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = normalizePath(newValue)
		if newValue == "" {
			return false, fmt.Errorf("local_path cannot be empty")
		}
		return confirmAndApplyString("local_path", folder.LocalPath, newValue, func(v string) {
			folder.LocalPath = v
			cfg.Folders[folderName] = folder
		})
	case 1:
		newValue, err := prompt.StringStyled("New remote_host (must match a hosts key)", folder.RemoteHost, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue == "" {
			return false, fmt.Errorf("remote_host cannot be empty")
		}
		if _, ok := cfg.Hosts[newValue]; !ok {
			return false, fmt.Errorf("host '%s' not found under hosts", newValue)
		}
		return confirmAndApplyString("remote_host", folder.RemoteHost, newValue, func(v string) {
			folder.RemoteHost = v
			cfg.Folders[folderName] = folder
		})
	case 2:
		newValue, err := prompt.StringStyled("New remote_subpath", folder.RemoteSubpath, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = strings.Trim(strings.TrimSpace(newValue), "/")
		if newValue == "" {
			return false, fmt.Errorf("remote_subpath cannot be empty")
		}
		return confirmAndApplyString("remote_subpath", folder.RemoteSubpath, newValue, func(v string) {
			folder.RemoteSubpath = v
			cfg.Folders[folderName] = folder
		})
	case 3:
		modeOptions := []string{"push", "pull", "bidirectional"}
		defaultIdx := 0
		for i, m := range modeOptions {
			if folder.Mode == m {
				defaultIdx = i
				break
			}
		}
		choice, err := prompt.SelectStyled("New mode", modeOptions, defaultIdx, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue := modeOptions[choice]
		return confirmAndApplyString("mode", folder.Mode, newValue, func(v string) {
			folder.Mode = v
			cfg.Folders[folderName] = folder
		})
	case 4:
		newValue, err := promptForBool("New delete", folder.Delete)
		if err != nil {
			return false, err
		}
		return confirmAndApplyBool("delete", folder.Delete, newValue, func(v bool) {
			folder.Delete = v
			cfg.Folders[folderName] = folder
		})
	case 5:
		newValue, err := promptForBool("New checksum", folder.Checksum)
		if err != nil {
			return false, err
		}
		return confirmAndApplyBool("checksum", folder.Checksum, newValue, func(v bool) {
			folder.Checksum = v
			cfg.Folders[folderName] = folder
		})
	case 6, 7, 8, 9:
		return editHostField(cfg, folder.RemoteHost, idx)
	default:
		return false, fmt.Errorf("unsupported menu selection")
	}
}

func editHostField(cfg *config.Config, hostName string, idx int) (bool, error) {
	host, ok := cfg.Hosts[hostName]
	if !ok {
		return false, fmt.Errorf("host '%s' not found; update folder remote_host first", hostName)
	}

	switch idx {
	case 6:
		newValue, err := prompt.StringStyled("New host user", host.User, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue == "" {
			return false, fmt.Errorf("host user cannot be empty")
		}
		return confirmAndApplyString("host.user", host.User, newValue, func(v string) {
			host.User = v
			cfg.Hosts[hostName] = host
		})
	case 7:
		newValue, err := prompt.StringStyled("New host host (IP/hostname)", host.Host, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue == "" {
			return false, fmt.Errorf("host host cannot be empty")
		}
		return confirmAndApplyString("host.host", host.Host, newValue, func(v string) {
			host.Host = v
			cfg.Hosts[hostName] = host
		})
	case 8:
		newValue, err := prompt.IntStyled("New host port", host.Port, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		if newValue <= 0 || newValue > 65535 {
			return false, fmt.Errorf("port must be between 1 and 65535")
		}
		return confirmAndApplyInt("host.port", host.Port, newValue, func(v int) {
			host.Port = v
			cfg.Hosts[hostName] = host
		})
	case 9:
		newValue, err := prompt.StringStyled("New host base_path", host.BasePath, prompt.Chevron, true)
		if err != nil {
			return false, err
		}
		newValue = normalizePath(newValue)
		if newValue == "" {
			return false, fmt.Errorf("host base_path cannot be empty")
		}
		return confirmAndApplyString("host.base_path", host.BasePath, newValue, func(v string) {
			host.BasePath = v
			cfg.Hosts[hostName] = host
		})
	default:
		return false, fmt.Errorf("invalid host field selection")
	}
}

func promptForBool(label string, current bool) (bool, error) {
	currentStr := strconv.FormatBool(current)
	raw, err := prompt.StringStyled(label+" (true/false)", currentStr, prompt.Chevron, true)
	if err != nil {
		return false, err
	}
	value, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(raw)))
	if err != nil {
		return false, fmt.Errorf("expected true or false")
	}
	return value, nil
}

func confirmAndApplyString(field, oldValue, newValue string, apply func(string)) (bool, error) {
	if oldValue == newValue {
		return false, nil
	}
	fmt.Printf("%sCurrent%s %s: %s\n", tui.ColorDim, tui.ColorReset, field, oldValue)
	fmt.Printf("%sNew%s %s: %s\n", tui.ColorCyan, tui.ColorReset, field, newValue)
	ok, err := prompt.ConfirmStyled("Are you sure you want to apply this change?", false, prompt.Chevron, true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	apply(newValue)
	return true, nil
}

func confirmAndApplyBool(field string, oldValue, newValue bool, apply func(bool)) (bool, error) {
	if oldValue == newValue {
		return false, nil
	}
	fmt.Printf("%sCurrent%s %s: %t\n", tui.ColorDim, tui.ColorReset, field, oldValue)
	fmt.Printf("%sNew%s %s: %t\n", tui.ColorCyan, tui.ColorReset, field, newValue)
	ok, err := prompt.ConfirmStyled("Are you sure you want to apply this change?", false, prompt.Chevron, true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	apply(newValue)
	return true, nil
}

func confirmAndApplyInt(field string, oldValue, newValue int, apply func(int)) (bool, error) {
	if oldValue == newValue {
		return false, nil
	}
	fmt.Printf("%sCurrent%s %s: %d\n", tui.ColorDim, tui.ColorReset, field, oldValue)
	fmt.Printf("%sNew%s %s: %d\n", tui.ColorCyan, tui.ColorReset, field, newValue)
	ok, err := prompt.ConfirmStyled("Are you sure you want to apply this change?", false, prompt.Chevron, true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	apply(newValue)
	return true, nil
}

func normalizePath(v string) string {
	cleaned := strings.TrimSpace(v)
	if cleaned == "" {
		return cleaned
	}
	if strings.HasPrefix(cleaned, "~") {
		return cleaned
	}
	return filepath.Clean(cleaned)
}
