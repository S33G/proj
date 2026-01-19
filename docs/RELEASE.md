# Release Guide

This guide explains how to release a new version of proj.

## Prerequisites

- Git repository access
- Go 1.24+ installed locally
- A GitHub account with push access to the repository

## Release Process

### 1. Prepare Your Changes

Ensure all changes are:
- Committed to the `main` branch
- Tested locally with `make test`
- Documented in `CHANGELOG.md`

Example:
```bash
git log --oneline origin/main..main  # Review commits to be released
make test                             # Run tests
```

### 2. Update Version Files

Update the version in `CHANGELOG.md` at the top with:
- New version number (following [Semantic Versioning](https://semver.org/))
- Release date
- Summary of changes

Example CHANGELOG.md entry:
```markdown
## [v0.9.0] - 2026-01-19

### Added
- New cmd/proj entry point with full CLI support
- Release automation with GitHub Actions
- Binary distribution for Linux and macOS

### Fixed
- Installation script now validates downloaded binaries
- Fixed .gitignore to allow cmd/proj source files

### Changed
- Release workflow now uploads raw binaries instead of tarballs
```

### 3. Create a Git Tag

Use semantic versioning (e.g., `v1.0.0`, `v0.9.0`, `v0.8.1-beta`).

```bash
# Create an annotated tag with release notes
git tag -a v0.9.0 -m "Release v0.9.0

- Added CLI entry point
- Automated release pipeline
- See CHANGELOG.md for details"

# Push the tag to GitHub
git push origin v0.9.0
```

**Important:** The tag must start with `v` and be a valid semantic version.

### 4. GitHub Actions Workflow Runs Automatically

Once you push the tag, the [release workflow](.github/workflows/release.yml) automatically:

1. **Detects the new release tag**
2. **Builds binaries** for all supported platforms:
   - `proj-linux-amd64`
   - `proj-linux-arm64`
   - `proj-darwin-amd64` (macOS Intel)
   - `proj-darwin-arm64` (macOS Apple Silicon)
3. **Creates SHA256 checksums** for each binary
4. **Uploads assets** to the GitHub release page

**Monitor the workflow:**
- Go to: https://github.com/s33g/proj/actions
- Look for the release workflow run for your tag
- Check that all build jobs completed successfully

### 5. Verify the Release

Once the workflow completes:

```bash
# View the release on GitHub
open https://github.com/s33g/proj/releases/tag/v0.9.0

# Download and test a binary
curl -L https://github.com/s33g/proj/releases/download/v0.9.0/proj-linux-amd64 \
  -o /tmp/proj-test && chmod +x /tmp/proj-test && /tmp/proj-test --version
```

### 6. Test the Installation Script

Verify that the new installation script works with the released binary:

```bash
# Run install script (it will download the latest release)
bash <(curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh)

# Test the installed binary
proj --version

# Verify it can detect projects
proj --init
proj --set-path ~/code
proj --list
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0): Breaking changes to API or CLI
- **MINOR** (v0.9.0): New features, backwards compatible
- **PATCH** (v0.8.1): Bug fixes only

Examples:
- `v1.0.0` - First stable release
- `v0.9.0` - New features during development
- `v0.8.1` - Bug fix for v0.8.0
- `v0.9.0-beta` - Pre-release version (optional)

## Changelog Format

Keep `CHANGELOG.md` organized with:
- Version and date in heading
- Subsections: Added, Changed, Fixed, Removed, Deprecated
- Bullet points with brief descriptions
- Reference issue/PR numbers if applicable

Example:
```markdown
## [v0.9.0] - 2026-01-19

### Added
- CLI entry point with --version, --help, --init commands (#42)
- Automated release pipeline with multi-platform builds

### Fixed
- Installation script now validates downloaded binaries (#41)
- Fixed binary upload format in release workflow

### Changed
- Release assets now include raw binaries (was tarballs)
```

## Troubleshooting

### Release workflow fails to build

Check the GitHub Actions logs:
1. Visit https://github.com/s33g/proj/actions
2. Click the failed workflow run
3. Review build logs for errors (usually Go compilation or missing dependencies)

**Common causes:**
- Go version mismatch (requires 1.24+)
- Missing `cmd/proj/main.go` entry point
- Syntax errors in code

### Binaries not uploaded to release

Verify:
1. Workflow completed successfully (check Actions tab)
2. Tag name starts with `v` (e.g., `v0.9.0`)
3. `softprops/action-gh-release@v2` has correct `files:` configuration

### Installation script can't find binary

Check that:
1. Binary was uploaded to the release (visit GitHub release page)
2. Binary name matches what script expects: `proj-linux-amd64`, `proj-darwin-arm64`, etc.
3. HTTP status code check in install.sh is working (should see "HTTP 200")

## Manual Release (if needed)

If GitHub Actions is unavailable:

```bash
# Build all platforms locally
make build-all VERSION=v0.9.0

# Create a release draft on GitHub
open https://github.com/s33g/proj/releases/new

# Upload binaries and checksums manually
# Then publish the release
```

## Post-Release

After a successful release:

1. **Update main branch** with any version bumps or release notes
2. **Announce** on relevant channels (social media, forums, etc.)
3. **Monitor issues** for bug reports from the new release
4. **Plan next iteration** based on feedback

## CI/CD Pipeline

The release pipeline is defined in [`.github/workflows/release.yml`](../.github/workflows/release.yml).

**Triggered by:** Creating a GitHub Release with a `v*` tag  
**Steps:**
1. Checkout code
2. Setup Go 1.24
3. Build binary for each platform
4. Generate SHA256 checksums
5. Upload assets to GitHub release

**Matrix builds for:**
- linux-amd64, linux-arm64
- darwin-amd64, darwin-arm64

Build time: ~2-3 minutes total

## Security

- Tags are signed commits recommended (not required)
- SHA256 checksums provided for verification
- GitHub Actions are pinned to specific versions

To verify a downloaded binary:
```bash
# Download checksum
curl -L https://github.com/s33g/proj/releases/download/v0.9.0/proj-linux-amd64.sha256 \
  -o proj-linux-amd64.sha256

# Verify download
sha256sum -c proj-linux-amd64.sha256

# Expected output: proj-linux-amd64: OK
```

## Questions?

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
