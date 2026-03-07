# mirusync Setup

Get syncing in two steps: **install** the tool, then run **`mirusync init`** and follow the prompts. No config editing required.

---

## 1. Install

### Option A: Homebrew (macOS)

If you have a Homebrew tap or the formula locally:

```bash
# From the project directory (after you have a release or built binary)
brew install --build-from-source ./Formula/mirusync.rb
```

To install from a GitHub release (once you publish releases):

```bash
# Add your tap, then:
brew tap hanif/mirusync
brew install mirusync
```

To create the tap and formula, see [Package manager (Homebrew)](#package-manager-homebrew) below.

### Option B: Download a release

1. Open the [Releases](https://github.com/hanif/mirusync/releases) page.
2. Download the archive for your OS (e.g. `mirusync_0.1.0_darwin_arm64.tar.gz` for Apple Silicon).
3. Unzip and move the `mirusync` binary into your PATH, e.g.:

   ```bash
   tar -xzf mirusync_0.1.0_darwin_arm64.tar.gz
   mv mirusync /usr/local/bin/
   ```

### Option C: Go install

```bash
go install github.com/hanif/mirusync@latest
```

Ensure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

### Option D: Build from source

```bash
git clone https://github.com/hanif/mirusync.git
cd mirusync
go build -o mirusync
sudo mv mirusync /usr/local/bin/
```

---

## 2. Run setup

```bash
mirusync init
```

The wizard will:

1. **Other machine** — Ask for the other computer’s IP or hostname and the username you use to log in.
2. **SSH key** — Show your SSH public key and optionally run `ssh-copy-id` so you can log in without a password. If you don’t have a key yet, it will tell you to run `ssh-keygen -t ed25519` and then run `mirusync init` again.
3. **Verify** — Test SSH to the other machine and only continue if that succeeds.
4. **Folder** — Ask which folder on this machine to sync and where it should live on the other machine (e.g. `~/dev/projects` ↔ `projects`).
5. **Direction** — Choose push only, pull only, or sync both ways.
6. **Save** — Write `~/.mirusync/config.yaml` and create the state directory.

After that you can run:

- `mirusync push <name>` — send folder to the other machine  
- `mirusync pull <name>` — get folder from the other machine  
- `mirusync sync <name>` — sync both ways (if you chose that)

No manual editing of config is required; you can change things later in `~/.mirusync/config.yaml` if you want.

---

## Requirements

- **This machine and the other machine**: SSH and `rsync` (macOS has these; on Linux install `openssh-client` and `rsync` if needed).
- **SSH key**: Generate with `ssh-keygen -t ed25519` if you don’t have one. The init wizard will show your public key and can run `ssh-copy-id` for you.

---

## Package manager (Homebrew)

To install via Homebrew you need a **tap** (your repo with a `Formula` directory) or a local formula file.

### 1. Add a formula file

In your repo, create a file like `Formula/mirusync.rb` (see the formula in this repo). It should point at a release tarball (e.g. GitHub release) or build from source.

### 2. Create a tap (optional)

```bash
# Create a repo named homebrew-mirusync (or homebrew-tap with multiple formulae)
# Put Formula/mirusync.rb in that repo, then:
brew tap hanif/mirusync https://github.com/hanif/homebrew-mirusync
brew install mirusync
```

### 3. Install from the project directory (no tap)

From the mirusync repo root, after building or downloading the binary:

```bash
brew install --build-from-source ./Formula/mirusync.rb
```

The formula in this repo is set up so you can run that from the cloned directory. For a stable release, point the formula’s `url` and `version` at a release tarball.

---

## Troubleshooting

| Problem | What to do |
|--------|------------|
| “No SSH public key found” | Run `ssh-keygen -t ed25519`, then run `mirusync init` again. |
| “Could not connect” during init | Make sure the other machine is on and reachable (e.g. `ping <ip>`). Ensure SSH is allowed (port 22 or your custom port). Try logging in with `ssh user@host`. |
| “Permission denied (publickey)” | Add your public key to the other machine. Let the wizard run `ssh-copy-id`, or manually append the contents of `~/.ssh/id_ed25519.pub` to `~/.ssh/authorized_keys` on the other machine. |
| Want to add another folder or host later | Edit `~/.mirusync/config.yaml` (see README for the format), or run `mirusync init` again and choose to overwrite/start fresh. |

Running `mirusync doctor` checks your config and SSH connectivity and can help narrow down issues.
