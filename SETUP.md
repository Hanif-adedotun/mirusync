# mirusync Setup Guide

Complete setup instructions for getting started with mirusync.

## Prerequisites

- macOS (tested on macOS, but should work on Linux)
- Go 1.21 or later (for building from source)
- SSH access configured between machines
- rsync installed (comes with macOS)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/hanif/mirusync.git
cd mirusync

# Build
go build -o mirusync

# Install to your PATH (optional)
sudo mv mirusync /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/hanif/mirusync@latest
```

### Using Make

```bash
# Build
make build

# Install to /usr/local/bin
make install
```

## Quick Start

### 1. Initialize Configuration

```bash
mirusync init
```

This creates `~/.mirusync/config.yaml` with a sample configuration.

### 2. Edit Configuration

Edit `~/.mirusync/config.yaml` to match your setup:

```yaml
hosts:
  laptopB:
    user: your-username
    host: 192.168.1.23  # or hostname
    port: 22
    base_path: /Users/your-username/dev

folders:
  projects:
    local_path: /Users/your-username/dev/projects
    remote_host: laptopB
    remote_subpath: projects
    mode: bidirectional  # or "push" or "pull"
    delete: false        # Set to true to enable delete sync
    checksum: false      # Use checksums (slower but safer)
```

### 3. Verify Configuration

```bash
mirusync doctor
```

This checks your configuration and tests SSH connectivity.

### 4. Sync Your Folders

```bash
# Push local changes to remote
mirusync push projects

# Pull remote changes to local
mirusync pull projects

# Bidirectional sync (requires mode: bidirectional)
mirusync sync projects
```

## SSH Setup

### Initial SSH Configuration

Ensure SSH key-based authentication is set up between your machines:

```bash
# On Laptop A, generate SSH key if you don't have one
ssh-keygen -t ed25519

# Copy public key to Laptop B
ssh-copy-id -p 22 user@laptopB-ip
```

### Test Connectivity

Test SSH connectivity manually:

```bash
ssh -p 22 user@laptopB-ip "echo ok"
```

You should see `ok` as output. If you encounter issues, see the Troubleshooting section below.

### SSH Key Permissions

Ensure your SSH keys have correct permissions:

```bash
chmod 600 ~/.ssh/id_*
chmod 644 ~/.ssh/id_*.pub
chmod 700 ~/.ssh
```

### SSH Config (Optional)

You can configure SSH in `~/.ssh/config` for easier access:

```
Host laptopB
    HostName 192.168.1.23
    User your-username
    Port 22
    IdentityFile ~/.ssh/id_ed25519
```

Then you can use the host alias in your mirusync config:

```yaml
hosts:
  laptopB:
    user: your-username
    host: laptopB  # Uses SSH config alias
    port: 22
    base_path: /Users/your-username/dev
```

## Configuration

### Configuration File Location

The configuration file is located at:
```
~/.mirusync/config.yaml
```

### Host Configuration

Each host entry defines a remote machine:

```yaml
hosts:
  hostName:
    user: username        # SSH username
    host: 192.168.1.23    # IP address or hostname
    port: 22              # SSH port (default: 22)
    base_path: /path/to   # Base path on remote machine
```

**Example:**
```yaml
hosts:
  laptopB:
    user: hanif
    host: 192.168.1.23
    port: 22
    base_path: /Users/hanif/dev
```

### Folder Configuration

Each folder entry defines what to sync:

```yaml
folders:
  folderName:
    local_path: /local/path        # Local folder path
    remote_host: hostName          # Reference to a host
    remote_subpath: subfolder      # Subpath under base_path
    mode: bidirectional            # "push", "pull", or "bidirectional"
    delete: false                  # Enable delete sync (default: false)
    checksum: false                # Use checksums (default: false)
```

**Example:**
```yaml
folders:
  projects:
    local_path: /Users/hanif/dev/projects
    remote_host: laptopB
    remote_subpath: projects
    mode: bidirectional
    delete: false
    checksum: false
```

### Mode Options

- **`push`**: One-way sync from local to remote
  - Use with `mirusync push <folder>`
  - Local changes overwrite remote
  
- **`pull`**: One-way sync from remote to local
  - Use with `mirusync pull <folder>`
  - Remote changes overwrite local
  
- **`bidirectional`**: Two-way sync
  - Use with `mirusync sync <folder>`
  - Merges changes from both sides
  - Detects and reports conflicts

### Configuration Options

#### `delete` (boolean)

When `true`, files that exist on the destination but not on the source will be deleted.

**Warning**: Use with caution. This is a destructive operation.

```yaml
folders:
  projects:
    delete: true  # Enable delete sync
```

#### `checksum` (boolean)

When `true`, uses file checksums instead of timestamps for comparison. Slower but safer.

```yaml
folders:
  projects:
    checksum: true  # Use checksums for comparison
```

**When to use checksums:**
- Critical data where timestamp accuracy is questionable
- Files that may have been modified without timestamp updates
- When you want maximum safety

**When not to use checksums:**
- Large files or many files (performance impact)
- When timestamps are reliable
- For faster syncs

## State Management

mirusync stores sync state in `~/.mirusync/state/`. Each folder has a JSON file tracking:

- Last sync timestamp
- Last sync direction (push/pull/sync)
- Last host used
- File count
- Conflicts (if any)

**State file location:**
```
~/.mirusync/state/<folder-name>.json
```

**Example state file:**
```json
{
  "last_sync": "2026-02-15T14:23:00Z",
  "last_direction": "push",
  "last_host": "laptopB",
  "file_count": 42,
  "conflicts": []
}
```

This enables:
- Conflict detection
- Status reporting
- Resume capability

## Safety Features

mirusync includes several safety guardrails:

1. **Forbidden paths**: Refuses to sync system directories like `/`, `/Users`, `/System`, `/usr`, `/bin`, etc.
2. **Dry-run by default**: The `sync` command defaults to dry-run mode
3. **Explicit delete flag**: Delete operations require explicit configuration
4. **Path validation**: Validates all paths before syncing

### Overriding Safety Checks

You can override safety checks with the `--force` flag (not recommended):

```bash
mirusync push projects --force
```

**Warning**: Only use `--force` if you understand the risks.

## Troubleshooting

### SSH Connection Issues

**Problem**: Cannot connect to remote host

**Solutions**:

1. Test SSH connectivity manually:
   ```bash
   ssh -p 22 user@host "echo ok"
   ```

2. Check SSH key authentication:
   ```bash
   ssh -v -p 22 user@host
   ```

3. Verify SSH keys are in the correct location:
   ```bash
   ls -la ~/.ssh/
   ```

4. Check SSH key permissions:
   ```bash
   chmod 600 ~/.ssh/id_*
   chmod 644 ~/.ssh/id_*.pub
   ```

5. Run mirusync doctor:
   ```bash
   mirusync doctor
   ```

### Configuration Errors

**Problem**: Configuration file errors

**Solutions**:

1. Validate configuration:
   ```bash
   mirusync doctor
   ```

2. Check config file syntax:
   ```bash
   cat ~/.mirusync/config.yaml
   ```

3. Verify YAML syntax (use a YAML validator):
   ```bash
   # Install yq if needed: brew install yq
   yq eval ~/.mirusync/config.yaml
   ```

4. Common issues:
   - Missing required fields (user, host, port, base_path)
   - Invalid YAML syntax (indentation, quotes)
   - Reference to non-existent host
   - Invalid mode value

### Permission Issues

**Problem**: Permission denied errors

**Solutions**:

1. Check local path permissions:
   ```bash
   ls -la /path/to/local/folder
   ```

2. Check remote path permissions:
   ```bash
   ssh user@host "ls -la /path/to/remote/folder"
   ```

3. Ensure SSH keys have correct permissions:
   ```bash
   chmod 600 ~/.ssh/id_*
   ```

4. Check if remote user has write access:
   ```bash
   ssh user@host "touch /path/to/remote/folder/test && rm /path/to/remote/folder/test"
   ```

### rsync Errors

**Problem**: rsync execution fails

**Solutions**:

1. Check if rsync is installed:
   ```bash
   which rsync
   rsync --version
   ```

2. Test rsync manually:
   ```bash
   rsync -avz --dry-run source/ user@host:/destination/
   ```

3. Check rsync permissions on remote:
   ```bash
   ssh user@host "which rsync"
   ```

### Network Issues

**Problem**: Connection timeouts or slow transfers

**Solutions**:

1. Check network connectivity:
   ```bash
   ping <remote-host-ip>
   ```

2. Test SSH connection speed:
   ```bash
   time ssh user@host "echo ok"
   ```

3. Consider using compression (already enabled with `-z` flag)

4. For slow networks, you may want to adjust SSH settings in `~/.ssh/config`:
   ```
   Host *
       ServerAliveInterval 60
       ServerAliveCountMax 3
   ```

### Path Not Found Errors

**Problem**: Local or remote path does not exist

**Solutions**:

1. Verify local path exists:
   ```bash
   ls -la /local/path
   ```

2. Verify remote path exists:
   ```bash
   ssh user@host "ls -la /remote/path"
   ```

3. Create missing directories:
   ```bash
   # Local
   mkdir -p /local/path
   
   # Remote
   ssh user@host "mkdir -p /remote/path"
   ```

4. Check path in configuration:
   ```bash
   cat ~/.mirusync/config.yaml
   ```

## Verification Checklist

After setup, verify everything works:

- [ ] mirusync binary is in PATH
- [ ] Configuration file created (`~/.mirusync/config.yaml`)
- [ ] SSH key authentication works
- [ ] `mirusync doctor` passes all checks
- [ ] Can run `mirusync push <folder> --dry-run`
- [ ] Can run `mirusync pull <folder> --dry-run`
- [ ] Local and remote paths exist and are accessible

## Next Steps

Once setup is complete:

1. Read the [README.md](README.md) for command reference
2. Review [ARCHITECTURE.md](ARCHITECTURE.md) for system design
3. Start syncing your folders!

## Getting Help

If you encounter issues not covered here:

1. Run `mirusync doctor` for diagnostics
2. Check the [README.md](README.md) troubleshooting section
3. Review error messages carefully
4. Open an issue on GitHub with:
   - Error message
   - Configuration (redact sensitive info)
   - Output of `mirusync doctor`

