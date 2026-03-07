package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"time"

	"github.com/hanif/mirusync/internal/config"
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
	return fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no %s@%s", host.Port, host.User, host.Host)
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

// CopyID runs ssh-copy-id to install the default public key on the remote host.
func CopyID(user, host string, port int) error {
	keyPath, _, err := DefaultPublicKey()
	if err != nil {
		return err
	}
	cmd := exec.Command("ssh-copy-id", "-i", keyPath, "-p", fmt.Sprintf("%d", port), fmt.Sprintf("%s@%s", user, host))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-copy-id failed: %w", err)
	}
	return nil
}


