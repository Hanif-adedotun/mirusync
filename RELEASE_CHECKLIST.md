# Release and Publish Checklist

Use this checklist to publish `mirusync` for non-technical users first, then Homebrew as an optional extra.

## 1) Prepare release artifacts

- Build archives for:
  - macOS Apple Silicon (`darwin_arm64`)
  - macOS Intel (`darwin_amd64`)
  - Windows 64-bit (`windows_amd64`)
- Use clear filenames like:
  - `mirusync_<version>_darwin_arm64.tar.gz`
  - `mirusync_<version>_darwin_amd64.tar.gz`
  - `mirusync_<version>_windows_amd64.zip`

## 2) Tag and publish GitHub Release

```bash
git tag v<version>
git push origin v<version>
```

Then create a GitHub Release for `v<version>` and upload the artifacts.

In the release description, include one short install paragraph:
- Download the file for your OS
- Unzip
- Run `mirusync init`

## 3) Keep docs install-first

- `README.md` points users to `INSTALL.md`
- `INSTALL.md` is simple and user-focused
- `SETUP.md` remains advanced/developer-oriented

## 4) Update Homebrew tap (optional but recommended)

In your tap repo (`homebrew-mirusync`), update `Formula/mirusync.rb`:
- `url` to `https://github.com/hanif/mirusync/archive/refs/tags/v<version>.tar.gz`
- `sha256` to the tarball checksum
- `version` to `<version>`

Then publish the formula update so users can run:

```bash
brew update
brew upgrade mirusync
```

## 5) Verify install paths

- Verify release install from `INSTALL.md`
- Verify Homebrew install:

```bash
brew tap hanif/mirusync
brew install mirusync
mirusync --help
```
