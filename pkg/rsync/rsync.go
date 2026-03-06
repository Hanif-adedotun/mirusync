package rsync

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type DryRunResult struct {
	FilesAdded    int
	FilesModified int
	FilesDeleted  int
	TotalSize     int64
	Output        string
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
		args = append(args, "-n") // dry-run
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

func parseDryRunOutput(output string) *DryRunResult {
	result := &DryRunResult{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Count files
		if strings.HasPrefix(line, ">f") || strings.HasPrefix(line, ">f+") {
			result.FilesAdded++
		} else if strings.HasPrefix(line, ">f.st") {
			result.FilesModified++
		} else if strings.HasPrefix(line, "*deleting") {
			result.FilesDeleted++
		}

		// Try to extract size (format: "total size is 12345")
		if strings.Contains(line, "total size is") {
			var size int64
			fmt.Sscanf(line, "total size is %d", &size)
			result.TotalSize = size
		}
	}

	return result
}


