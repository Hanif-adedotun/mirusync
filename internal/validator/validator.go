package validator

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Forbidden-path check is done in CheckForbiddenPath(path, force) so --force is respected.
	// Here we only validate emptiness, resolvability, and existence.

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

// CheckForbiddenPath returns an error if path is one of the forbidden roots (e.g. /, /Users, /System).
// Paths under /Users/username/... are allowed; only the exact roots are forbidden.
// If force is true, the check is skipped.
func CheckForbiddenPath(path string, force bool) error {
	if force {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	absPath = filepath.Clean(absPath)

	for _, forbidden := range forbiddenPaths {
		// Only forbid exact match: syncing the root itself (e.g. /, /Users, /System).
		// Do not forbid /Users/hanif/Documents/... or other user content.
		if absPath == forbidden {
			return fmt.Errorf("safety error: path '%s' is a forbidden directory. Use --force to override (not recommended)", absPath)
		}
	}

	return nil
}


