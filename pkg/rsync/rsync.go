package rsync

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"unicode"
)

type DryRunResult struct {
	FilesAdded    int
	FilesModified int
	FilesDeleted  int
	TotalSize     int64
	Output        string
}

// TotalChanges returns the sum of add/modify/delete counts from parsed itemize lines.
func (r *DryRunResult) TotalChanges() int {
	if r == nil {
		return 0
	}
	return r.FilesAdded + r.FilesModified + r.FilesDeleted
}

type RSyncOptions struct {
	Source      string
	Destination string
	Delete      bool
	Checksum    bool
	DryRun      bool
	SSHCommand  string
}

func Execute(options RSyncOptions) error {
	args := buildArgs(options)

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync error: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func DryRun(options RSyncOptions) (*DryRunResult, error) {
	options.DryRun = true
	args := buildArgs(options)

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// rsync returns non-zero for dry-run even when successful
		// We need to check if it's actually an error
		if !strings.Contains(string(output), "sending incremental file list") {
			return nil, fmt.Errorf("rsync dry-run error: %w\nOutput: %s", err, string(output))
		}
	}

	result := parseDryRunOutput(string(output))
	result.Output = string(output)
	return result, nil
}

func buildArgs(options RSyncOptions) []string {
	args := []string{
		"-avz", // archive, verbose, compress
	}

	if options.DryRun {
		// Dry-run: do not make changes, and emit a machine-parseable line per change.
		args = append(args, "-n", "--out-format=%i|%n") // dry-run with structured output
	}

	if options.Delete {
		args = append(args, "--delete")
	}

	if options.Checksum {
		args = append(args, "--checksum")
	}

	if options.SSHCommand != "" {
		args = append(args, "-e", options.SSHCommand)
	}

	args = append(args, options.Source, options.Destination)

	return args
}

// itemizeNewMask is true when rsync's 9 update-flag characters are all '+' (brand-new path).
func itemizeNewMask(itemize string) bool {
	return len(itemize) >= 11 && itemize[2:11] == "+++++++++"
}

func applyItemizeChange(result *DryRunResult, itemize string) {
	if strings.HasPrefix(itemize, "*deleting") {
		result.FilesDeleted++
		return
	}

	// Directory created on receiver: "cd+++++++++..." (see rsync --itemize-changes).
	if strings.HasPrefix(itemize, "cd") && len(itemize) >= 3 {
		if itemizeNewMask(itemize) {
			result.FilesAdded++
		} else if len(itemize) >= 11 {
			result.FilesModified++
		}
		return
	}

	if len(itemize) < 2 || itemize[0] != '>' {
		return
	}

	kind := itemize[1]
	switch kind {
	case 'f', 'd', 'L':
		if itemizeNewMask(itemize) {
			result.FilesAdded++
		} else {
			result.FilesModified++
		}
	default:
		// Other transfer types (e.g. devices) ignored for summary counts.
	}
}

func parseTotalSizeBytes(line string) int64 {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "total size is") {
		return 0
	}
	rest := strings.TrimPrefix(line, "total size is")
	rest = strings.TrimSpace(rest)
	// Stop at first space (rsync may append "speedup is ..." on the same line).
	if i := strings.IndexFunc(rest, unicode.IsSpace); i >= 0 {
		rest = rest[:i]
	}
	rest = strings.ReplaceAll(rest, ",", "")
	var n int64
	_, _ = fmt.Sscanf(rest, "%d", &n)
	return n
}

func parseDryRunOutput(output string) *DryRunResult {
	result := &DryRunResult{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "total size is") {
			if sz := parseTotalSizeBytes(line); sz > 0 {
				result.TotalSize = sz
			}
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		change := strings.TrimSpace(parts[0])
		if change == "" {
			continue
		}

		applyItemizeChange(result, change)
	}

	return result
}
