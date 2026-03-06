# mirusync Architecture

## Overview

mirusync is a production-grade folder synchronization tool designed for developers working across multiple machines. It orchestrates rsync over SSH to provide safe, efficient file synchronization with strong correctness guarantees.

## Design Principles

1. **Minimal Surface Area**: Uses system SSH and rsync, no embedded implementations
2. **Strong Correctness**: Multiple validation layers and safety guardrails
3. **Extensibility**: Clean separation of concerns, modular architecture
4. **Production-Ready**: Comprehensive error handling and state management

## System Model

```
+----------------+             SSH              +----------------+
|   Laptop A     |  <------------------------>  |   Laptop B     |
|                |                               |                |
|  mirusync CLI  |                               |  mirusync CLI  |
|  config.yaml   |                               |  config.yaml   |
|  rsync engine  |                               |  rsync engine  |
+----------------+                               +----------------+
```

**Key Insight**: mirusync orchestrates. rsync transfers. SSH transports.

## Architecture Layers

### 1. CLI Layer (`cmd/`)

**Responsibility**: User interface and command orchestration

**Technology**: Cobra CLI framework

**Commands**:
- `init`: Initialize configuration
- `push`: One-way sync (local → remote)
- `pull`: One-way sync (remote → local)
- `sync`: Bidirectional sync
- `status`: Show sync status
- `doctor`: Diagnose configuration

**Key Files**:
- `cmd/root.go`: Root command and initialization
- `cmd/push.go`: Push command implementation
- `cmd/pull.go`: Pull command implementation
- `cmd/sync.go`: Sync command implementation
- `cmd/status.go`: Status command implementation
- `cmd/doctor.go`: Diagnostic command implementation

### 2. Configuration Layer (`internal/config/`)

**Responsibility**: Configuration management and abstraction

**Technology**: Viper + YAML

**Abstractions**:
- **Host**: Named remote machine configuration
- **Folder**: Named folder sync configuration

**Configuration Structure**:
```yaml
hosts:
  hostName:
    user: string
    host: string
    port: int
    base_path: string

folders:
  folderName:
    local_path: string
    remote_host: string
    remote_subpath: string
    mode: "push" | "pull" | "bidirectional"
    delete: bool
    checksum: bool
```

**Key Files**:
- `internal/config/config.go`: Configuration loading and access

### 3. Validation Layer (`internal/validator/`)

**Responsibility**: Safety checks and path validation

**Validations**:
- Configuration completeness
- Path existence
- Forbidden path detection (`/`, `/Users`, `/System`, etc.)
- Mode validation
- Host validation

**Safety Guardrails**:
1. Refuses to sync system directories
2. Requires explicit `--force` for dangerous operations
3. Validates all paths before execution
4. Checks SSH connectivity before sync

**Key Files**:
- `internal/validator/validator.go`: Validation logic

### 4. SSH Abstraction Layer (`internal/ssh/`)

**Responsibility**: SSH connectivity and command building

**Features**:
- Connectivity testing with timeout
- SSH command construction
- Remote path building

**Design Decision**: Uses system SSH, no embedded implementation

**Key Files**:
- `internal/ssh/ssh.go`: SSH utilities

### 5. State Management (`internal/state/`)

**Responsibility**: Sync history and conflict tracking

**State Structure**:
```json
{
  "last_sync": "2026-02-15T14:23:00Z",
  "last_direction": "push",
  "last_host": "laptopB",
  "file_count": 42,
  "conflicts": ["file1.conflict.local", "file2.conflict.remote"]
}
```

**Storage**: `~/.mirusync/state/<folder>.json`

**Key Files**:
- `internal/state/state.go`: State persistence

### 6. Sync Engine (`internal/engine/`)

**Responsibility**: Orchestrating sync operations

**Operations**:
1. **Push**: Local → Remote
2. **Pull**: Remote → Local
3. **Sync**: Bidirectional with conflict detection

**Flow**:
```
1. Validate configuration
2. Check SSH connectivity
3. Dry-run (always)
4. Show preview
5. Execute (if not dry-run)
6. Update state
```

**Key Files**:
- `internal/engine/engine.go`: Engine implementation

### 7. rsync Wrapper (`pkg/rsync/`)

**Responsibility**: rsync execution and output parsing

**Features**:
- Dry-run execution
- Output parsing for preview
- Flag management
- Error handling

**rsync Flags Used**:
- `-avz`: Archive, verbose, compress
- `-n`: Dry-run
- `--delete`: Delete extraneous files
- `--checksum`: Use checksums instead of timestamps
- `-e`: SSH command specification

**Key Files**:
- `pkg/rsync/rsync.go`: rsync wrapper

## Data Flow

### Push Operation

```
User Command
    ↓
CLI Layer (cmd/push.go)
    ↓
Validation Layer (validator)
    ├─ Validate folder config
    ├─ Check forbidden paths
    └─ Validate host
    ↓
SSH Layer (ssh)
    └─ Test connectivity
    ↓
Engine (engine)
    ├─ Build paths
    ├─ Build SSH command
    └─ Execute rsync
        ↓
rsync Wrapper (pkg/rsync)
    ├─ Dry-run first
    ├─ Parse output
    └─ Execute if not dry-run
    ↓
State Management (state)
    └─ Save sync state
    ↓
Success
```

### Sync Operation (Bidirectional)

```
User Command
    ↓
CLI Layer (cmd/sync.go)
    ↓
Validation Layer
    └─ Ensure mode is "bidirectional"
    ↓
Engine (engine)
    ├─ Step 1: Pull remote changes
    │   └─ (same flow as pull)
    ├─ Step 2: Push local changes
    │   └─ (same flow as push)
    └─ Step 3: Conflict detection
        └─ Compare state
    ↓
State Management
    └─ Save sync state with conflicts
    ↓
Success
```

## Error Handling Strategy

### Error Categories

**Category A - Configuration Error**
- Missing folder/host
- Invalid configuration
- **Response**: Clear error message, exit non-zero

**Category B - Network Error**
- SSH unreachable
- Connection timeout
- **Response**: Diagnostic message, suggest `mirusync doctor`

**Category C - Execution Error**
- rsync exit code ≠ 0
- Permission denied
- **Response**: Show rsync output, exit non-zero

### Error Propagation

All errors bubble up through layers with context:
```
rsync error → engine error → CLI error → user
```

## Safety Mechanisms

### 1. Forbidden Paths

Prevents syncing:
- `/`
- `/Users`
- `/System`
- `/usr`, `/bin`, `/sbin`
- `/etc`, `/var`, `/tmp`
- `/opt`, `/private`

**Override**: `--force` flag (not recommended)

### 2. Dry-Run Default

- `sync` command defaults to `--dry-run`
- All operations show preview before execution
- User must explicitly opt-in to changes

### 3. Explicit Delete

- `delete: false` by default
- Requires explicit configuration
- Never auto-deletes without user intent

### 4. Path Validation

- Validates all paths exist
- Resolves to absolute paths
- Checks permissions

## Conflict Detection

### Current Implementation

Simplified conflict detection based on:
- Last sync timestamp
- File modification times
- State comparison

### Future Enhancement

Full conflict detection would:
1. Compute file hashes (MD5/SHA256)
2. Compare local vs remote hashes
3. Compare modification times
4. Detect concurrent modifications
5. Create `.conflict.local` and `.conflict.remote` files

## File Comparison Strategy

### Default Mode: Timestamp + Size

- Fast comparison
- Uses rsync's default algorithm
- Good for most use cases

### Checksum Mode

- Enabled with `checksum: true`
- Slower but safer
- Compares file contents
- Use for critical data

## State Persistence

### Storage Location

`~/.mirusync/state/<folder>.json`

### State Schema

```go
type SyncState struct {
    LastSync     time.Time
    LastDirection string  // "push", "pull", "sync"
    LastHost     string
    FileCount    int
    Conflicts    []string
}
```

### Use Cases

1. **Conflict Detection**: Compare last sync time with file mtimes
2. **Status Reporting**: Show sync history
3. **Resume Capability**: Track what was synced
4. **Conflict Tracking**: List unresolved conflicts

## Extension Points

### 1. Watch Mode

Add `fsnotify` integration:
```go
// internal/watch/watcher.go
type Watcher struct {
    folder string
    events chan fsnotify.Event
}
```

### 2. Parallel Transfers

Add concurrent rsync execution:
```go
// internal/engine/parallel.go
func (e *Engine) SyncParallel(folders []string) error
```

### 3. Encryption

Add encryption layer:
```go
// internal/encrypt/encrypt.go
func EncryptPath(path string) (string, error)
```

### 4. Cloud Relay

Add optional relay server:
```go
// internal/relay/relay.go
type RelayClient struct {
    endpoint string
}
```

## Testing Strategy

### Unit Tests

- Configuration loading
- Validation logic
- State management
- Path utilities

### Integration Tests

- SSH connectivity
- rsync execution
- End-to-end sync operations

### Manual Testing

- Two-machine setup
- Network failure scenarios
- Conflict scenarios
- Permission edge cases

## Performance Considerations

### rsync Efficiency

- Delta transfer (only changed parts)
- Compression (`-z` flag)
- Incremental sync

### Network Optimization

- SSH connection reuse (future)
- Compression tuning
- Bandwidth limiting (future)

### Local Optimization

- Parallel file operations (future)
- Caching (future)

## Security Considerations

### SSH Security

- Uses system SSH configuration
- Respects `~/.ssh/config`
- No credential storage
- Key-based authentication recommended

### Path Security

- Validates all paths
- Prevents directory traversal
- Forbidden path checks

### State Security

- State files contain no secrets
- Only sync metadata
- User-readable JSON

## Deployment

### Binary Distribution

Single static binary:
```bash
go build -o mirusync
```

### Installation

1. Copy binary to PATH
2. Run `mirusync init`
3. Configure `~/.mirusync/config.yaml`
4. Run `mirusync doctor`

## Future Architecture Enhancements

1. **Plugin System**: Extensible sync strategies
2. **Metrics**: Sync performance tracking
3. **Notifications**: Desktop notifications for sync completion
4. **Web UI**: Optional web dashboard
5. **Multi-host**: Sync to multiple hosts simultaneously

## Summary

mirusync follows a layered architecture with clear separation of concerns:

| Layer         | Responsibility          | Technology    |
| ------------- | ----------------------- | ------------- |
| CLI           | User interface          | Cobra         |
| Config        | Configuration           | Viper + YAML  |
| Validator     | Safety checks           | Go            |
| Engine        | Sync orchestration      | Go            |
| rsync wrapper | File transfer           | rsync         |
| State         | History tracking        | JSON          |
| SSH           | Connectivity            | system SSH    |

This architecture provides:
- ✅ Clear boundaries
- ✅ Testability
- ✅ Extensibility
- ✅ Maintainability
- ✅ Production readiness


