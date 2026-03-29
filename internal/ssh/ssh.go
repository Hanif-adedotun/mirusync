package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hanif/mirusync/internal/config"
)

const (
	mirusyncKeyName = "id_mirusync"
)

func CheckConnectivity(hostName string, timeout time.Duration) error {
	host, err := config.GetHost(hostName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	return CheckConnectivityRaw(host.User, host.Host, host.Port, timeout)
}

// CheckConnectivityRaw verifies SSH access without using config (e.g. for init wizard).
func CheckConnectivityRaw(user, host string, port int, timeout time.Duration) error {
	sshCmd := fmt.Sprintf("ssh -p %d -o ConnectTimeout=%d -o StrictHostKeyChecking=accept-new %s@%s 'echo ok'",
		port, int(timeout.Seconds()), user, host)
	cmd := exec.Command("sh", "-c", sshCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot connect to %s@%s:%d - %v (output: %s)", user, host, port, err, string(output))
	}
	if string(output) != "ok\n" {
		return fmt.Errorf("unexpected response from SSH")
	}
	return nil
}

func BuildSSHCommand(hostName string) string {
	host, err := config.GetHost(hostName)
	if err != nil {
		return ""
	}
	// For rsync's -e option we must pass ONLY the remote shell (ssh + options),
	// NOT the user@host target. rsync adds user@host itself via the source/dest.
	// Example:
	//   rsync -e "ssh -p 22 -o StrictHostKeyChecking=no" src user@host:/path
	// If we include user@host here, the remote shell sees "user@host" as the command
	// to run, which leads to `command not found: user@host`.
	return fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no", host.Port)
}

func BuildRemotePath(hostName string, subpath string) string {
	host, err := config.GetHost(hostName)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s@%s:%s/%s", host.User, host.Host, host.BasePath, subpath)
}

func BuildRSyncRemotePath(hostName string, subpath string) string {
	host, err := config.GetHost(hostName)
	if err != nil {
		return ""
	}
	// rsync format: user@host:/path/to/dir
	return fmt.Sprintf("%s@%s:%s/%s", host.User, host.Host, host.BasePath, subpath)
}

// EnsureMirusyncKey returns the path and contents of a dedicated, passphrase-less
// SSH public key for mirusync (~/.ssh/id_mirusync.pub). If it does not exist,
// it is created via ssh-keygen with an empty passphrase.
func EnsureMirusyncKey() (pubPath string, pub []byte, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", nil, err
	}

	privPath := filepath.Join(sshDir, mirusyncKeyName)
	pubPath = privPath + ".pub"

	// If pubkey already exists, just read it.
	if data, readErr := os.ReadFile(pubPath); readErr == nil {
		return pubPath, data, nil
	}

	// Generate a new ed25519 key with NO passphrase for mirusync.
	comment := "mirusync@" + hostnameFallback()
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", privPath, "-C", comment)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("ssh-keygen failed: %w", err)
	}

	data, err := os.ReadFile(pubPath)
	if err != nil {
		return "", nil, err
	}
	return pubPath, data, nil
}

func hostnameFallback() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "mirusync"
	}
	return h
}

// DefaultPublicKey returns the first found public key path (ed25519 then rsa).
func DefaultPublicKey() (path string, content []byte, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	for _, name := range []string{"id_ed25519.pub", "id_rsa.pub"} {
		p := filepath.Join(sshDir, name)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		return p, data, nil
	}
	return "", nil, fmt.Errorf("no SSH public key found in %s (run: ssh-keygen -t ed25519)", sshDir)
}

// CopyID runs ssh-copy-id to install the specified public key on the remote host.
func CopyID(keyPath, user, host string, port int) error {
	cmd := exec.Command("ssh-copy-id", "-i", keyPath, "-p", fmt.Sprintf("%d", port), fmt.Sprintf("%s@%s", user, host))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-copy-id failed: %w", err)
	}
	return nil
}


