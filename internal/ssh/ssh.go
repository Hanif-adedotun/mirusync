package ssh

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/hanif/mirusync/internal/config"
)

func CheckConnectivity(hostName string, timeout time.Duration) error {
	host, err := config.GetHost(hostName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Build SSH command
	sshCmd := fmt.Sprintf("ssh -p %d -o ConnectTimeout=%d -o StrictHostKeyChecking=no %s@%s 'echo ok'",
		host.Port, int(timeout.Seconds()), host.User, host.Host)

	cmd := exec.Command("sh", "-c", sshCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("network error: cannot connect to %s@%s:%d - %v (output: %s)",
			host.User, host.Host, host.Port, err, string(output))
	}

	if string(output) != "ok\n" {
		return fmt.Errorf("network error: unexpected response from SSH connection")
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


