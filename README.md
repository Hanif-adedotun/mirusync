# mirusync

A production-grade personal devtool for synchronizing folders between two Macs over SSH using rsync.

## Overview

`mirusync` is a CLI tool designed for developers who work across multiple machines and need reliable, safe folder synchronization. It provides:

- **One-way sync** (push/pull) for simple workflows
- **Bidirectional sync** with conflict detection
- **Safety guardrails** to prevent accidental data loss
- **Dry-run preview** before any changes
- **State management** for tracking sync history
- **SSH-based** synchronization (no central server required)

## Features

- ✅ Minimal surface area - uses system SSH and rsync
- ✅ Strong correctness guarantees with validation layers
- ✅ Extensible architecture
- ✅ Production-ready error handling
- ✅ Conflict detection and reporting
- ✅ Configurable per-folder settings

## Getting Started

See [SETUP.md](SETUP.md) for install options (Homebrew, release, Go, or from source).

**Quick setup:**
1. Install mirusync (e.g. `go install github.com/hanif/mirusync@latest` or use the [Homebrew formula](Formula/mirusync.rb)).
2. Run **`mirusync init`** and follow the prompts — you’ll enter the other machine’s address and user, set up SSH, pick a folder, and choose sync direction. No config editing required.
3. Then run `mirusync push <name>`, `mirusync pull <name>`, or `mirusync sync <name>` to sync.

## Commands

### `mirusync init`

Interactive setup: asks for the other machine’s IP/hostname and user, shows your SSH public key (and can run `ssh-copy-id`), verifies the connection, then asks which folder to sync and in which direction. Writes `~/.mirusync/config.yaml` for you.

### `mirusync push <folder>`

Push local folder to remote (one-way: local → remote).

**Options:**
- `--dry-run`: Show what would be synced without actually syncing
- `--verbose-dry-run`: Also print raw rsync dry-run output (useful for debugging)
- `--force`: Override safety checks (not recommended)

**Example:**
```bash
mirusync push projects
mirusync push projects --dry-run
```

### `mirusync pull <folder>`

Pull remote folder to local (one-way: remote → local).

**Options:**
- `--dry-run`: Show what would be synced without actually syncing
- `--verbose-dry-run`: Also print raw rsync dry-run output (useful for debugging)
- `--force`: Override safety checks (not recommended)

**Example:**
```bash
mirusync pull projects
```

### `mirusync sync <folder>`

Bidirectional sync between local and remote. Requires the folder to be configured with `mode: bidirectional`.

**Options:**
- `--dry-run`: Show what would be synced (default: true)
- `--verbose-dry-run`: Also print raw rsync dry-run output (useful for debugging)
- `--force`: Override safety checks (not recommended)

**Example:**
```bash
mirusync sync projects
```

### `mirusync status [folder]`

Show sync status for a folder, or all folders if no folder is specified.

**Example:**
```bash
mirusync status
mirusync status projects
```

### `mirusync doctor`

Diagnose configuration and connectivity issues. Validates:
- Configuration file syntax
- Host definitions
- Folder definitions
- SSH connectivity
- Path accessibility

**Example:**
```bash
mirusync doctor
```

## Configuration

See [SETUP.md](SETUP.md) for detailed configuration instructions including:
- Host and folder configuration
- SSH setup
- Configuration options
- Safety features
- State management

## Example Workflow

### Scenario: Working on Two Macs

**Laptop A (at cafe):**
```bash
# Make changes to your project
cd ~/dev/projects/myapp
# ... edit files ...

# Push changes to Laptop B
mirusync push projects
```

**Laptop B (at home):**
```bash
# Pull changes from Laptop A
mirusync pull projects

# Make more changes
# ... edit files ...

# Push back
mirusync push projects
```

**Laptop A (back at cafe):**
```bash
# Bidirectional sync to merge changes
mirusync sync projects
```

## Troubleshooting

For detailed troubleshooting, see [SETUP.md](SETUP.md).

**Quick diagnostics:**
```bash
# Run comprehensive diagnostics
mirusync doctor

# Test SSH connectivity
ssh -p 22 user@host "echo ok"
```

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation.

## Development

### Project Structure

```
mirusync/
├── cmd/              # CLI commands
├── internal/         # Internal packages
│   ├── config/       # Configuration management
│   ├── engine/       # Sync engine
│   ├── ssh/          # SSH abstraction
│   ├── state/        # State management
│   └── validator/    # Validation layer
├── pkg/              # Public packages
│   └── rsync/        # rsync wrapper
└── main.go           # Entry point
```

### Building

```bash
go build -o mirusync
```

### Testing

```bash
go test ./...
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Why Go?

- Static binary distribution
- Fast execution
- Strong typing
- Excellent CLI tool ecosystem
- Production-ready standard library

## Future Enhancements

Potential future features:
- Watch mode using `fsnotify`
- Compression toggle
- Parallel transfers
- Encrypted folder mode
- Optional cloud relay
- Minimal TUI dashboard

