# Release Checklist - ScribesAI

This document outlines the steps required to prepare and release a new version of ScribesAI.

## Pre-Release Preparation

### Code Quality

- [ ] All code merged to main branch
- [ ] No WIP commits or debug code
- [ ] All TODOs in code addressed or documented
- [ ] Code formatted with `go fmt`
- [ ] No linting errors: `go vet ./...`
- [ ] Dependencies up to date: `go mod tidy`

### Testing

- [ ] All E2E tests pass (see E2E_TESTING.md)
- [ ] Tested on target platforms:
  - [ ] macOS Intel (x86_64)
  - [ ] macOS Apple Silicon (ARM64)
  - [ ] Linux x86_64 (Ubuntu/Debian)
- [ ] Performance benchmarks acceptable
- [ ] No memory leaks detected
- [ ] Audio recording stable
- [ ] Transcription working correctly
- [ ] Claude AI analysis functional ⚠️ (blocked by SDK issue)
- [ ] All export formats work (Markdown, JSON, Text)

### Documentation

- [ ] README.md is up to date
- [ ] CLAUDE.md reflects current architecture
- [ ] VOSK_SETUP.md has correct installation steps
- [ ] E2E_TESTING.md is comprehensive
- [ ] CHANGELOG.md updated with new version
- [ ] Known issues documented

## Version Numbering

ScribesAI uses [Semantic Versioning](https://semver.org/): MAJOR.MINOR.PATCH

- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

**Current Version Strategy**:
- Pre-1.0: Development phase (0.x.x)
- 1.0.0: First production release
- Post-1.0: Semantic versioning

**Determine Version**:
```
Current: 0.5.0 (Phase 5 - Production Polish)
Next:    0.6.0 (if new features added)
         0.5.1 (if only bug fixes)
         1.0.0 (when production-ready)
```

## Build Process

### 1. Update Version Information

```bash
# Create VERSION file
echo "0.5.0" > VERSION

# Update version in main.go (if applicable)
# Add --version flag implementation
```

### 2. Update CHANGELOG.md

```markdown
# Changelog

## [0.5.0] - 2025-11-14

### Added
- Automatic directory creation on startup
- Functional settings view with environment configuration
- Model download UI with progress feedback
- Enhanced error messages with troubleshooting steps
- Export success feedback banners
- Comprehensive E2E testing documentation

### Fixed
- Directory creation errors on first launch
- Settings view showing hardcoded values
- Missing feedback after export operations
- Unclear error messages without recovery steps

### Changed
- Improved model download workflow with UI
- Enhanced user feedback throughout application

## [0.4.0] - 2025-11-13
...
```

### 3. Build Binaries

#### macOS Intel
```bash
GOOS=darwin GOARCH=amd64 go build -o dist/scribescli-darwin-amd64 ./cmd/scribescli
```

#### macOS Apple Silicon
```bash
GOOS=darwin GOARCH=arm64 go build -o dist/scribescli-darwin-arm64 ./cmd/scribescli
```

#### Linux x86_64
```bash
GOOS=linux GOARCH=amd64 go build -o dist/scribescli-linux-amd64 ./cmd/scribescli
```

### 4. Verify Builds

```bash
# Check each binary
file dist/scribescli-*

# Test run (if on matching platform)
./dist/scribescli-darwin-arm64 --version
```

## Release Assets

### Required Files

1. **Binaries** (from dist/)
   - `scribescli-darwin-amd64`
   - `scribescli-darwin-arm64`
   - `scribescli-linux-amd64`

2. **Documentation**
   - README.md
   - VOSK_SETUP.md
   - E2E_TESTING.md
   - PLATFORMS.md
   - LICENSE

3. **Installation Scripts**
   - scripts/install-vosk.sh

4. **Example Files**
   - .env.example (if exists)

### Create Release Archives

```bash
# macOS Intel
tar -czf scribescli-v0.5.0-darwin-amd64.tar.gz \
  dist/scribescli-darwin-amd64 \
  README.md VOSK_SETUP.md PLATFORMS.md LICENSE \
  scripts/install-vosk.sh

# macOS Apple Silicon
tar -czf scribescli-v0.5.0-darwin-arm64.tar.gz \
  dist/scribescli-darwin-arm64 \
  README.md VOSK_SETUP.md PLATFORMS.md LICENSE \
  scripts/install-vosk.sh

# Linux
tar -czf scribescli-v0.5.0-linux-amd64.tar.gz \
  dist/scribescli-linux-amd64 \
  README.md VOSK_SETUP.md PLATFORMS.md LICENSE \
  scripts/install-vosk.sh
```

### Generate Checksums

```bash
# Generate SHA256 checksums
shasum -a 256 scribescli-v0.5.0-*.tar.gz > checksums.txt

# Verify
cat checksums.txt
```

## Git Release Process

### 1. Commit Release Changes

```bash
# Stage changes
git add CHANGELOG.md VERSION README.md

# Commit
git commit -m "Release v0.5.0

- Phase 5: Production Polish & Demo Readiness
- Enhanced UX with better error messages and feedback
- Model download UI implementation
- Comprehensive testing documentation
"
```

### 2. Create Git Tag

```bash
# Create annotated tag
git tag -a v0.5.0 -m "ScribesAI v0.5.0 - Production Polish

This release focuses on production readiness and user experience improvements.

Key Features:
- Automatic setup and configuration
- Model download with progress UI
- Enhanced error messages
- Export success feedback
- Comprehensive testing guide

See CHANGELOG.md for full details.
"

# Verify tag
git tag -v v0.5.0
git show v0.5.0
```

### 3. Push to Repository

```bash
# Push main branch
git push origin main

# Push tags
git push origin v0.5.0
```

## GitHub Release

### 1. Create GitHub Release

1. Go to: https://github.com/storo/scribescli/releases/new
2. Select tag: `v0.5.0`
3. Title: `ScribesAI v0.5.0 - Production Polish`
4. Description:

```markdown
# ScribesAI v0.5.0 - Production Polish & Demo Readiness

This release focuses on making ScribesAI production-ready with enhanced user experience, better error handling, and comprehensive testing documentation.

## 🎉 New Features

- **Automatic Model Download UI**: Beautiful progress display when downloading Vosk models
- **Functional Settings View**: Real configuration from environment variables
- **Export Success Feedback**: Visual confirmation when exports complete
- **Enhanced Error Messages**: All errors now include helpful troubleshooting steps
- **Automatic Directory Setup**: No more manual directory creation

## 📚 Documentation

- **E2E_TESTING.md**: Comprehensive testing guide with 15+ test scenarios
- **Updated CLAUDE.md**: Reflects all Phase 1-5 implementation

## 🔧 Improvements

- Error messages now include recovery steps and suggestions
- Model download shows clear progress and status
- Settings view loads actual configuration from .env
- Success banners for export operations
- All critical directories created automatically

## 🐛 Bug Fixes

- Fixed directory creation errors on first launch
- Removed hardcoded values from settings view
- Improved model download error handling

## 📥 Installation

### macOS (Intel)
```bash
tar -xzf scribescli-v0.5.0-darwin-amd64.tar.gz
cd scribescli
./scripts/install-vosk.sh
./scribescli
```

### macOS (Apple Silicon)
```bash
tar -xzf scribescli-v0.5.0-darwin-arm64.tar.gz
cd scribescli
./scripts/install-vosk.sh
./scribescli
```

### Linux (x86_64)
```bash
tar -xzf scribescli-v0.5.0-linux-amd64.tar.gz
cd scribescli
./scripts/install-vosk.sh
./scribescli
```

## ⚠️ Known Issues

- **Anthropic SDK Compatibility**: Claude AI analysis currently non-functional due to SDK API changes. Recording and transcription work correctly. Fix in progress.
- **PortAudio Required**: System-level PortAudio installation required before building

See [CHANGELOG.md](https://github.com/storo/scribescli/blob/main/CHANGELOG.md) for full details.

## 📝 Checksums

See `checksums.txt` for SHA256 verification.

---

*"I'm putting myself to the fullest possible use, which is all I think that any conscious entity can ever hope to do."* - HAL 9000
```

### 2. Upload Release Assets

- [ ] scribescli-v0.5.0-darwin-amd64.tar.gz
- [ ] scribescli-v0.5.0-darwin-arm64.tar.gz
- [ ] scribescli-v0.5.0-linux-amd64.tar.gz
- [ ] checksums.txt

### 3. Mark as Pre-release

- [ ] Check "This is a pre-release" (for versions < 1.0.0)
- [ ] Publish release

## Post-Release

### Verification

- [ ] Download release assets and verify checksums
- [ ] Test installation from release tarball
- [ ] Verify all links in release notes work
- [ ] Check GitHub release appears correctly

### Communication

- [ ] Update project README with latest version
- [ ] Announce on relevant channels (if applicable)
- [ ] Update documentation site (if applicable)

### Cleanup

- [ ] Archive old release artifacts (keep last 3 versions)
- [ ] Update development version in main branch
  ```bash
  echo "0.6.0-dev" > VERSION
  git commit -am "Bump version to 0.6.0-dev"
  git push
  ```

## Rollback Procedure

If critical issues are discovered after release:

### 1. Assess Severity

- **Critical**: Data loss, security vulnerability, complete failure
  → Immediate rollback and hotfix
- **Major**: Feature broken, significant UX issue
  → Rollback and fix in next version
- **Minor**: Small bug, cosmetic issue
  → Document and fix in next version

### 2. Rollback Steps

```bash
# Delete GitHub release
# (Go to Releases → Edit → Delete)

# Delete git tag locally
git tag -d v0.5.0

# Delete remote tag
git push origin :refs/tags/v0.5.0

# Revert commits if needed
git revert <commit-hash>
```

### 3. Hotfix Process

```bash
# Create hotfix branch from tag
git checkout -b hotfix/v0.5.1 v0.5.0

# Fix issue
# ... make changes ...

# Commit fix
git commit -am "Hotfix: Fix critical issue"

# Create new tag
git tag -a v0.5.1 -m "Hotfix release"

# Push
git push origin hotfix/v0.5.1
git push origin v0.5.1

# Merge back to main
git checkout main
git merge hotfix/v0.5.1
git push
```

## Release Checklist Summary

**Pre-Release**:
- [ ] Code quality checks complete
- [ ] All tests pass on all platforms
- [ ] Documentation updated
- [ ] Version number decided
- [ ] CHANGELOG.md updated

**Build**:
- [ ] VERSION file created
- [ ] Binaries built for all platforms
- [ ] Release archives created
- [ ] Checksums generated

**Git**:
- [ ] Changes committed
- [ ] Tag created
- [ ] Pushed to repository

**GitHub**:
- [ ] Release created with description
- [ ] Assets uploaded
- [ ] Marked as pre-release (if < 1.0)
- [ ] Published

**Post-Release**:
- [ ] Downloads verified
- [ ] Installation tested
- [ ] Communication sent
- [ ] Cleanup complete

---

## Notes for v1.0.0 Release

When ready for production (v1.0.0):

1. **Resolve all blocking issues**:
   - [ ] Fix Anthropic SDK compatibility
   - [ ] Complete full E2E testing on all platforms
   - [ ] Address all known critical bugs

2. **Production criteria**:
   - [ ] All core features working
   - [ ] Performance acceptable
   - [ ] Documentation complete
   - [ ] Security reviewed
   - [ ] Installation smooth on all platforms

3. **Additional requirements**:
   - [ ] User feedback incorporated
   - [ ] Beta testing complete
   - [ ] Migration guide (if needed)
   - [ ] Support channels established

---

**Last Updated**: 2025-11-14
**Document Version**: 1.0
