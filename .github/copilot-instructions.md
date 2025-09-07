# obt - GitHub Release Binary Downloader

`obt` is a Go CLI tool that automatically downloads the latest release binaries from GitHub repositories based on your OS and architecture. Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build
- Install dependencies: `go mod download` - completes instantly
- Build the binary: `go build -o obt .` - takes ~5 seconds. NEVER CANCEL.
- Run tests: `go test -v ./...` - takes ~30 seconds. NEVER CANCEL. Note: Some network-dependent tests may fail in restricted environments, this is expected and not a problem.
- Install linter: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` - takes ~5 minutes first time. NEVER CANCEL. Set timeout to 10+ minutes.
- Run linter: `golangci-lint run -D errcheck` - takes ~30 seconds. NEVER CANCEL. Set timeout to 2+ minutes.

### Development Workflow  
- Always run `go mod tidy` before building
- Build command: `go build` (creates `obt` binary)
- Test command: `go test -v ./...` 
- Lint command: `golangci-lint run -D errcheck`
- Clean build cache: `go clean -testcache`

## Validation

### Manual Testing Scenarios
After making any changes to the code, ALWAYS run through these validation scenarios:

1. **Basic CLI functionality:**
   ```bash
   ./obt --help      # Should show usage information
   ./obt -v          # Should show version (typically "obt devel")
   ```

2. **Configuration commands:**
   ```bash
   ./obt -s /tmp/test-install    # Set install path
   ./obt -installed              # Show installed binaries table
   ```

3. **Build and test cycle:**
   ```bash
   go build -o obt .             # Build successfully 
   go test -v ./...              # Run tests (network tests may fail)
   golangci-lint run -D errcheck  # Lint passes
   ```

### Expected Test Behavior
- Unit tests for downloaders (TestDownloader_*) should PASS
- Integration tests (TestDownload*) may FAIL due to network restrictions - this is expected
- History and validation tests should PASS
- Total test time: ~30 seconds

### CI Requirements
- Always run `golangci-lint run -D errcheck` before committing - CI will fail without it
- Tests run in GitHub Actions with network access, so integration tests pass in CI
- Build uses Go 1.23+ (check .github/workflows/ci.yml)

## Key Commands and Timing

| Command | Time | Timeout | Notes |
|---------|------|---------|-------|
| `go mod download` | instant | 2 min | Downloads dependencies |
| `go build` | ~5s | 2 min | Creates obt binary |
| `go test -v ./...` | ~30s | 5 min | Some network tests fail locally |
| `golangci-lint run -D errcheck` | ~30s | 5 min | Requires prior installation |
| `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` | ~5 min | 15 min | First-time install only |

**CRITICAL**: NEVER CANCEL any of these commands. Set appropriate timeouts as shown above.

## Repository Structure

### Key Files
- `main.go` - CLI entry point, flag parsing, main application logic
- `downloader.go` - GitHub release download functionality  
- `history.go`, `history_file.go` - Installation history tracking
- `updater.go` - Binary update functionality
- `*_test.go` - Comprehensive test suite
- `go.mod` - Go module dependencies
- `.goreleaser.yml` - Release automation config
- `.github/workflows/ci.yml` - CI pipeline

### Important Directories
- `.github/workflows/` - GitHub Actions CI/CD
- `testdata/` - Test fixtures and sample files

## Common Tasks

### Building and Testing
```bash
# Full build and test cycle
go mod tidy
go build -o obt .
go test -v ./...
golangci-lint run -D errcheck
```

### Using the Tool
```bash
# Show help
./obt --help

# Install a binary from GitHub releases  
./obt https://github.com/user/repo

# Install to specific path
./obt -p /usr/local/bin https://github.com/user/repo

# Install specific binary name
./obt -b binary-name https://github.com/user/repo

# Set default install path
./obt -s /path/to/install/dir

# Show installed binaries
./obt -installed

# Update all installed binaries
./obt -U
```

### Development Notes
- The tool supports multiple archive formats: tar.gz, zip, gzip, tar.xz
- Uses GitHub API to find appropriate releases for current OS/architecture
- Maintains installation history for updates and tracking
- Network-dependent functionality requires internet access
- Default install paths: Linux/macOS `/usr/local/bin/`, Windows `.`

## Troubleshooting
- If network tests fail locally: Expected behavior, they pass in CI with network access
- If `golangci-lint` not found: Run the install command shown above
- If build fails: Run `go mod tidy` and retry
- If binary doesn't work: Verify it was built with `go build -o obt .`

Always validate that the binary works with `./obt --help` and `./obt -v` after any changes.