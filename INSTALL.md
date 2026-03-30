# Install mirusync

This page is the easiest install path for most users.

If you are a developer and want Go/source install options, see [SETUP.md](SETUP.md).

## Option 1 (Recommended): Download from GitHub Releases

1. Open [mirusync Releases](https://github.com/hanif/mirusync/releases).
2. Download the file for your OS:
   - **Mac (Apple Silicon)**: `mirusync_<version>_darwin_arm64.tar.gz`
   - **Mac (Intel)**: `mirusync_<version>_darwin_amd64.tar.gz`
   - **Windows (64-bit)**: `mirusync_<version>_windows_amd64.zip`
3. Unzip the file.
4. Run `mirusync` from the extracted folder, or move it somewhere convenient.

## macOS quick command path

If you prefer one copy-paste command after download:

```bash
tar -xzf "mirusync_<version>_darwin_arm64.tar.gz"
chmod +x mirusync
sudo mv mirusync /usr/local/bin/
mirusync --help
```

Use the `darwin_amd64` file if your Mac is Intel.

## Windows quick command path (PowerShell)

After extracting the zip:

```powershell
.\mirusync.exe --help
```

If you want to run it from anywhere, move `mirusync.exe` into a folder that is on your `PATH`.

## Option 2 (macOS): Homebrew (optional)

Homebrew is a nice extra for users who already have brew installed.

1. Install Homebrew from [brew.sh](https://brew.sh) (if needed).
2. Then run:

```bash
brew tap hanif/mirusync
brew install mirusync
```

## After install

Run the setup wizard:

```bash
mirusync init
```

Then sync with:

```bash
mirusync push <name>
mirusync pull <name>
mirusync sync <name>
```

## Need help?

- Start with `mirusync doctor`
- See [SETUP.md](SETUP.md) for troubleshooting and advanced setup
