package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hanif/mirusync/internal/config"
)

const (
	mirusyncKeyName = "id_mirusync"
)

// MirusyncPrivateKeyPath returns the path to ~/.ssh/id_mirusync (private key).
func MirusyncPrivateKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", mirusyncKeyName), nil
}

// shellSingleQuote wraps s in single quotes for safe use inside sh -c / rsync -e.
func shellSingleQuote(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `'"'"'`) + `'`
}

func sshCommonOptions(strictHostKey string, connectTimeoutSec int) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, err
	}

	privPath, err := MirusyncPrivateKeyPath()
	if err != nil {
		return nil, err
	}

	opts := []string{
		"-o", fmt.Sprintf("ConnectTimeout=%d", connectTimeoutSec),
		"-o", "StrictHostKeyChecking=" + strictHostKey,
	}

	if st, err := os.Stat(privPath); err == nil && !st.IsDir() {
		opts = append(opts,
			"-i", privPath,
			"-o", "IdentitiesOnly=yes",
			"-o", "BatchMode=yes",
			"-o", "ControlMaster=auto",
			"-o", "ControlPath=" + filepath.Join(sshDir, "mirusync-cm-%r@%h:%p"),
			"-o", "ControlPersist=60",
		)
	}

	return opts, nil
}

// buildSSHExecArgs returns argv for `ssh` (no remote target or command).
func buildSSHExecArgs(port int, strictHostKey string, connectTimeoutSec int) ([]string, error) {
	opts, err := sshCommonOptions(strictHostKey, connectTimeoutSec)
	if err != nil {
		return nil, err
	}
	args := []string{"ssh", "-p", fmt.Sprintf("%d", port)}
	args = append(args, opts...)
	return args, nil
}

func CheckConnectivity(hostName string, timeout time.Duration) error {
	host, err := config.GetHost(hostName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	return CheckConnectivityRaw(host.User, host.Host, host.Port, timeout)
}

// CheckConnectivityRaw verifies SSH access without using config (e.g. for init wizard).
func CheckConnectivityRaw(user, host string, port int, timeout time.Duration) error {
	args, err := buildSSHExecArgs(port, "accept-new", int(timeout.Seconds()))
	if err != nil {
		return err
	}
	args = append(args, fmt.Sprintf("%s@%s", user, host), "echo ok")
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot connect to %s@%s:%d - %v (output: %s)", user, host, port, err, string(output))
	}
	if string(output) != "ok\n" {
		return fmt.Errorf("unexpected response from SSH")
	}
	return nil
}

// BuildSSHCommand returns the shell command string for rsync's -e option (ssh + options only).
func BuildSSHCommand(hostName string) string {
	host, err := config.GetHost(hostName)
	if err != nil {
		return ""
	}
	s, err := buildSSHCommandForPort(host.Port, "no", 30)
	if err != nil {
		return ""
	}
	return s
}

// buildSSHCommandForPort builds `ssh ...` for rsync -e; strictHostKey is `no` or `accept-new`.
func buildSSHCommandForPort(port int, strictHostKey string, connectTimeoutSec int) (string, error) {
	opts, err := sshCommonOptions(strictHostKey, connectTimeoutSec)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("ssh ")
	b.WriteString("-p ")
	b.WriteString(fmt.Sprintf("%d", port))
	for i := 0; i < len(opts); i += 2 {
		if i+1 >= len(opts) {
			break
		}
		b.WriteString(" ")
		b.WriteString(opts[i])
		b.WriteString(" ")
		// Quote values that may contain spaces or special chars (e.g. ControlPath).
		val := opts[i+1]
		if strings.ContainsAny(val, " \t\n\"'\\") || strings.HasPrefix(val, "-") {
			b.WriteString(shellSingleQuote(val))
		} else {
			b.WriteString(val)
		}
	}
	return b.String(), nil
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
