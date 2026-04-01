
# mirusync

https://github.com/user-attachments/assets/6aa6edf2-de7e-4bef-a8c8-41f7996f4a08

Simple folder sync between two machines over SSH.

For most people, the easiest path is:
1. Download a release from [GitHub Releases](https://github.com/hanif/mirusync/releases)
2. Run `mirusync init`
3. Use `mirusync push`, `mirusync pull`, or `mirusync sync`

## Install (Recommended)

See [INSTALL.md](INSTALL.md) for step-by-step instructions with:
- Mac (Apple Silicon and Intel) release downloads
- Windows release downloads
- Homebrew as an optional convenience on macOS

If you are technical and prefer Go tooling, see [SETUP.md](SETUP.md).
To publish new versions, follow [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md).

## What mirusync does

`mirusync` helps you keep project folders in sync across two machines with clear safety checks.

- One-way sync (`push` / `pull`)
- Bidirectional sync with conflict detection (`sync`)
- Dry-run preview before changes
- SSH-based sync (no central server)
- Config/state files managed under `~/.mirusync`

## Quick Start

After installing:

```bash
mirusync init
```

Then:

```bash
mirusync push <name>
mirusync pull <name>
mirusync sync <name>
```

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

### `mirusync edit <folder>`

Interactively edit an existing saved folder config with numbered options.
For each selected field, mirusync shows the current value, asks for a new value,
then asks for confirmation before saving.

You can edit:
- Folder fields: `local_path`, `remote_host`, `remote_subpath`, `mode`, `delete`, `checksum`
- Linked host fields: `user`, `host`, `port`, `base_path`

**Example:**
```bash
mirusync edit grad-school
```

## Configuration

See [SETUP.md](SETUP.md) for advanced configuration details including:
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

**Find your internal IP (LAN) on macOS:**
```bash
# Wi-Fi
ipconfig getifaddr en0

# Ethernet (if used)
ipconfig getifaddr en1
```

Use the IP shown on the *other* machine as `host` during `mirusync init`.

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

For advanced install/build options (`go install`, source build, Homebrew formula workflow), see [SETUP.md](SETUP.md).

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

