package rsync

import (
	"testing"
)

// Sample snippets shaped like rsync 2.6.x / macOS system rsync dry-run output
// (--out-format=%i|%n, -n). Itemize strings follow rsync --itemize-changes.

func TestParseDryRunOutput_newFileAndModified(t *testing.T) {
	out := `sending incremental file list
>f+++++++++|newfile.txt
>f.st......|changed.txt

sent 123 bytes  received 45 bytes  total 168
total size is 1310720  speedup is 1234.5
`
	r := parseDryRunOutput(out)
	if r.FilesAdded != 1 || r.FilesModified != 1 {
		t.Fatalf("got added=%d modified=%d, want 1 and 1", r.FilesAdded, r.FilesModified)
	}
	if r.TotalSize != 1310720 {
		t.Fatalf("TotalSize: got %d want 1310720", r.TotalSize)
	}
}

func TestParseDryRunOutput_newDirectory(t *testing.T) {
	out := `sending incremental file list
cd+++++++++|subdir/

sent 100 bytes  received 10 bytes  total 110
total size is 0  speedup is 0.00
`
	r := parseDryRunOutput(out)
	if r.FilesAdded != 1 {
		t.Fatalf("got added=%d, want 1 for new directory", r.FilesAdded)
	}
}

func TestParseDryRunOutput_newDirTransfer(t *testing.T) {
	out := `sending incremental file list
>d+++++++++|otherdir/

sent 50 bytes  received 5 bytes  total 55
total size is 1024  speedup is 1.00
`
	r := parseDryRunOutput(out)
	if r.FilesAdded != 1 {
		t.Fatalf("got added=%d, want 1 for >d new", r.FilesAdded)
	}
	if r.TotalSize != 1024 {
		t.Fatalf("TotalSize: got %d", r.TotalSize)
	}
}

func TestParseDryRunOutput_delete(t *testing.T) {
	out := `sending incremental file list
*deleting|gone.txt

sent 10 bytes  received 0 bytes  total 10
total size is 0  speedup is 0.00
`
	r := parseDryRunOutput(out)
	if r.FilesDeleted != 1 {
		t.Fatalf("got deleted=%d, want 1", r.FilesDeleted)
	}
}

func TestParseTotalSizeBytes_commas(t *testing.T) {
	line := "total size is 1,234,567  speedup is 1.00"
	n := parseTotalSizeBytes(line)
	if n != 1234567 {
		t.Fatalf("got %d, want 1234567", n)
	}
}

func TestDryRunResult_TotalChanges(t *testing.T) {
	r := &DryRunResult{FilesAdded: 2, FilesModified: 1, FilesDeleted: 1}
	if r.TotalChanges() != 4 {
		t.Fatalf("TotalChanges: got %d", r.TotalChanges())
	}
}
