package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanif/mirusync/internal/config"
)

var forbiddenPaths = []string{
	"/",
	"/Users",
	"/System",
	"/usr",
	"/bin",
	"/sbin",
	"/etc",
	"/var",
	"/tmp",
	"/opt",
	"/private",
}

func ValidateFolder(folderName string) error {
	folder, err := config.GetFolder(folderName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Validate local path
	if err := validatePath(folder.LocalPath, "local_path"); err != nil {
		return err
	}

	// Validate remote host exists
	_, err = config.GetHost(folder.RemoteHost)
	if err != nil {
		return fmt.Errorf("configuration error: remote_host '%s' not found: %w", folder.RemoteHost, err)
	}

	// Validate mode
	validModes := map[string]bool{
		"push":          true,
		"pull":          true,
		"bidirectional": true,
	}
	if !validModes[folder.Mode] {
		return fmt.Errorf("configuration error: invalid mode '%s'. Must be 'push', 'pull', or 'bidirectional'", folder.Mode)
	}

	return nil
}

func validatePath(path, label string) error {
	if path == "" {
		return fmt.Errorf("configuration error: %s is empty", label)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("configuration error: %s '%s' is invalid: %w", label, path, err)
	}

	// Check if path is forbidden
	for _, forbidden := range forbiddenPaths {
		if absPath == forbidden || strings.HasPrefix(absPath, forbidden+"/") {
			return fmt.Errorf("safety error: %s '%s' is in a forbidden directory. Use --force to override (not recommended)", label, absPath)
		}
	}

	// Check if local path exists (for local_path only)
	if label == "local_path" {
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("validation error: %s '%s' does not exist", label, absPath)
		}
	}

	return nil
}

func ValidateHost(hostName string) error {
	host, err := config.GetHost(hostName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if host.User == "" {
		return fmt.Errorf("configuration error: host '%s' has no user specified", hostName)
	}

	if host.Host == "" {
		return fmt.Errorf("configuration error: host '%s' has no host specified", hostName)
	}

	if host.Port <= 0 || host.Port > 65535 {
		return fmt.Errorf("configuration error: host '%s' has invalid port %d", hostName, host.Port)
	}

	if host.BasePath == "" {
		return fmt.Errorf("configuration error: host '%s' has no base_path specified", hostName)
	}

	return nil
}

func CheckForbiddenPath(path string, force bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, forbidden := range forbiddenPaths {
		if absPath == forbidden || strings.HasPrefix(absPath, forbidden+"/") {
			if !force {
				return fmt.Errorf("safety error: path '%s' is in a forbidden directory. Use --force to override (not recommended)", absPath)
			}
		}
	}

	return nil
}


