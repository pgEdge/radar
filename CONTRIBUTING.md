# Contributing to radar

Thank you for your interest in contributing to radar! This document provides guidelines and information for developers and contributors.

## Design Philosophy

radar is built on these core principles:

- **Zero Dependencies**: The final `radar` binary should be self-contained and not require any runtime dependencies (like Python, Bash scripts, etc.) other than the tools it uses for collection (like `iostat`, `psql`, etc.). This makes it highly portable.
- **Agentless**: It runs on demand and doesn't require a background agent.
- **Minimal Impact**: The tool should be as lightweight as possible and avoid any changes to the system it's running on. Data collection is read-only.
- **Graceful Failure**: If a collector fails (e.g., a command isn't installed or a file isn't readable), the tool silently skips it and moves on. This ensures that it collects as much data as possible without getting stopped by minor issues.
- **Comprehensive by Default**: It aims to collect a wide array of data that is commonly needed for diagnosing system and database performance issues.
- **Focused Collection**: Simple, targeted data gathering—raw metrics and configuration only.
- **Streaming Architecture**: Minimal memory usage even with large datasets through direct streaming to ZIP.

## Exit Codes

radar uses the following exit codes:

- `0` - Success
- `1` - Usage error (missing required flags)
- `3` - Collection/archive error
- `4` - No data collected

**Note**: PostgreSQL connection failures during normal collection will not cause the tool to exit with an error code. Instead, the tool will continue with system-only collection and report the connection issue to the user.

## Development Setup

### Prerequisites

- **Go**: 1.23 or higher
- **Git**: For version control
- **golangci-lint**: For linting (install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- **Docker**: For integration testing (optional but recommended)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/vyruss/radar.git
cd radar

# Build the binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .

# Run the binary
./radar --help
```

### Cross-Platform Development

radar uses Go build tags for platform-specific code:

- `system_tasks_linux.go` - Linux-only system collectors (`//go:build linux`)
- `system_tasks_darwin.go` - macOS-only system collectors (`//go:build darwin`)
- `system_tasks_shared.go` - Cross-platform system collectors (`//go:build linux || darwin`)
- `postgres_tasks.go` - PostgreSQL collectors (platform-independent)

#### Building for Specific Platforms

```bash
# Build for current platform
go build -o radar .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o radar-linux .

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -o radar-darwin .

# Build for both platforms
GOOS=linux GOARCH=amd64 go build -o radar-linux . && \
GOOS=darwin GOARCH=amd64 go build -o radar-darwin .
```

#### Adding Platform-Specific Collectors

When adding a new system collector:

1. **Determine platform compatibility**:
   - Works on all Unix-like systems → add to `system_tasks_shared.go`
   - Linux-specific (uses /proc, /sys, systemd, etc.) → add to `system_tasks_linux.go`
   - macOS-specific (uses sysctl, system_profiler, etc.) → add to `system_tasks_darwin.go`

2. **Add the collector definition** to the appropriate file

3. **Test on target platform(s)** - ensure the command/file exists and produces expected output

4. **Update tests** if the collector should be verified in CI

Example - adding a cross-platform collector to `system_tasks_shared.go`:

```go
{
    Name:        "my-collector",
    ArchivePath: "system/my_collector.out",
    Command:     "some-command",
    Args:        []string{"--flag"},
},
```

#### Local Testing on macOS

On macOS, you can test locally without Docker:

```bash
# Install PostgreSQL via Homebrew
brew install postgresql@18
brew services start postgresql@18

# Create test database
createdb testdb
psql -d testdb -c "CREATE EXTENSION pg_statviz;"
psql -d testdb -c "SELECT pgstatviz.snapshot();"

# Build and run
go build -o radar .
./radar -d testdb -v

# Verify output
unzip -l radar-*.zip | head -50
```

### Development Dependencies

For a complete development environment:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Ensure it's in your PATH
export PATH="$HOME/go/bin:$PATH"
```

## Adding Collectors

The tool is designed to be easily extensible with new collectors.

### System Collectors

**Command-based collector:**

```go
// In system_tasks.go
systemCommandTasks = append(systemCommandTasks, CollectionTask{
    Name: "my-command",
    Path: "system/my-command.out",
    Collector: func(db *sql.DB, cfg *Config, zw *zip.Writer) error {
        return execCommandCollector(zw, "system/my-command.out", "mycommand", "--flags")
    },
})
```

**File-based collector:**

```go
// In system_tasks.go
systemFileTasks = append(systemFileTasks, CollectionTask{
    Name: "/path/to/file",
    Path: "system/myfile.out",
    Collector: func(db *sql.DB, cfg *Config, zw *zip.Writer) error {
        return readFileCollector(zw, "system/myfile.out", "/path/to/file")
    },
})
```

### PostgreSQL Collectors

**Instance-level query collector:**

```go
// In postgres_tasks.go (postgresQueryTasks)
postgresQueryTasks = append(postgresQueryTasks, CollectionTask{
    Name: "my_view",
    Path: "postgresql/my_view.tsv",
    Collector: func(db *sql.DB, cfg *Config, zw *zip.Writer) error {
        query := `SELECT col1, col2 FROM my_view ORDER BY col1`
        return execQueryCollector(db, zw, "postgresql/my_view.tsv", query)
    },
})
```

**Per-database query collector:**

```go
// In postgres_tasks.go (perDatabaseQueryTasks)
// The framework automatically connects to each database to run per-database queries
perDatabaseQueryTasks = append(perDatabaseQueryTasks, CollectionTask{
    Name: "my_database_view",
    Path: "my_database_view.tsv",
    Collector: func(db *sql.DB, cfg *Config, zw *zip.Writer) error {
        query := `SELECT * FROM my_database_view`
        return execQueryCollector(db, zw, "my_database_view.tsv", query)
    },
})
```

**pg_statviz collector:**

```go
// In postgres_tasks.go (pgStatvizQueryTasks)
pgStatvizQueryTasks = append(pgStatvizQueryTasks, CollectionTask{
    Name: "pg_statviz_table",
    Path: "pg_statviz_table.tsv",
    Collector: func(db *sql.DB, cfg *Config, zw *zip.Writer) error {
        query := `SELECT * FROM pgstatviz.table_name ORDER BY snap_ts`
        return execQueryCollector(db, zw, "pg_statviz_table.tsv", query)
    },
})
```

### Testing Requirements

When adding collectors:

1. **Update unit tests** in `radar_test.go`
2. **Update integration tests** if permission-dependent
3. **Test all 4 permission scenarios** if applicable
4. **Ensure silent handling** of missing tools/permissions

## Architecture

### Code Organization

```
radar/
├── radar.go              # Main entry point, CLI parsing, orchestration
├── radar_test.go         # Unit tests
├── postgres.go           # PostgreSQL connection and collection
├── postgres_tasks.go     # PostgreSQL collector definitions
├── system_tasks.go       # System collector definitions
├── test-radar.sh         # Integration test script
├── run-ci-local.sh       # Local CI/CD script
├── Dockerfile            # Docker container for testing
└── README.md             # User documentation
```

### Task-Driven System

radar uses a table-driven architecture with `CollectionTask` structs:

```go
type CollectionTask struct {
    Name        string              # Collector name
    Path        string              # Output file path in ZIP
    Collector   func(...) error     # Collection function
}
```

Three types of collectors:

1. **Query Collectors** - Execute PostgreSQL queries
   - Use `execQueryCollector` helper
   - Store results as TSV files

2. **Command Collectors** - Execute system commands
   - Use `execCommandCollector` helper
   - Store stdout/stderr as .out files

3. **File Collectors** - Read files from filesystem
   - Use `readFileCollector` helper
   - Store raw file contents

All tasks are registered in slices:
- `postgresQueryTasks` - PostgreSQL query collectors
- `systemCommandTasks` - System command collectors
- `systemFileTasks` - System file collectors
- `postgresConfigFileTasks` - PostgreSQL config file collectors
- `pgStatvizQueryTasks` - pg_statviz collectors

### Error Handling

Three-state tracking for collectors:

1. **Collected** (✓) - Successfully collected data
2. **Skipped** (⊘) - Tool/file unavailable or permissions insufficient
3. **Failed** - Genuine error (reported to user)

Custom error type `SkipError` signals intentional skips:

```go
type SkipError struct {
    Reason string
}
```

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test -v ./...

# Run specific test
go test -v -run TestSkipFlagValidation
```

### Local CI Script

Run the complete CI/CD test suite locally:

```bash
./run-ci-local.sh
```

This script performs:
1. Code formatting check (gofmt)
2. Linting (golangci-lint)
3. Unit tests (go test)
4. Binary build
5. Integration tests in Docker with PostgreSQL 18 and pg_statviz
6. Creates timestamped log file (ci-YYYYMMDD-HHMMSS.log)

### Docker-Based Integration Testing

```bash
# Build and test in Debian container
docker build -t radar-test .
docker run --rm radar-test

# Manual testing after build
./radar --skip-postgres -v

# Verify ZIP contents
unzip -l radar-*.zip
```

## Continuous Integration

The project uses GitHub Actions for CI/CD. See [.github/workflows/ci.yml](.github/workflows/ci.yml) for details.

### Integration Test Scenarios

The integration test validates radar works correctly in all permission combinations:

**Scenario 1: Root + PostgreSQL superuser**
- Full system access (all commands available to root)
- Full PostgreSQL access (all views, config files)
- Expected: ~66 system collectors, ~32 PostgreSQL collectors, pg_statviz data

**Scenario 2: Root + PostgreSQL pg_monitor role**
- Full system access (root privileges)
- Monitoring-level PostgreSQL access (most views, limited config)
- Expected: ~66 system collectors, ~29 PostgreSQL collectors, pg_statviz data
- Note: Some collectors unavailable (e.g., pg_hba_file_rules, subscriptions)

**Scenario 3: Non-root + PostgreSQL superuser**
- Limited system access (some commands fail without root)
- Full PostgreSQL access (all views, config files)
- Expected: ~63 system collectors, ~32 PostgreSQL collectors, pg_statviz data
- Note: Some system collectors unavailable (e.g., ifconfig, sysctl)

**Scenario 4: Non-root + PostgreSQL pg_monitor role**
- Limited system access (non-root user)
- Monitoring-level PostgreSQL access (pg_monitor role)
- Expected: ~63 system collectors, ~29 PostgreSQL collectors, pg_statviz data
- Note: Combines limitations of both non-root and pg_monitor

All scenarios verify that radar handles permission limitations and collects maximum available data.

### Collector Availability Differences by Permission Scenario

This table shows **only collectors with meaningful differences** between permission scenarios. Collectors that work identically across all scenarios are omitted for clarity.

**Permission-Dependent Collectors:**

| Collector | S1 | S2 | S3 | S4 | Requirement |
|-----------|----|----|----|----|-------------|
| **dmesg** | ✓ | ✓ | ⊘ | ⊘ | Root system access |
| **dmesg-t** | ✓ | ✓ | ⊘ | ⊘ | Root system access |
| **pg_hba.conf** | ✓ | ✓ | ⊘ | ⊘ | Filesystem access to data directory |
| **pg_ident.conf** | ✓ | ✓ | ⊘ | ⊘ | Filesystem access to data directory |
| **postgresql.conf** | ✓ | ✓ | ⊘ | ⊘ | Filesystem access to data directory |
| **postgresql.auto.conf** | ✓ | ✓ | ⊘ | ⊘ | Filesystem access to data directory |
| **pg_hba_file_rules** | ✓ | ⊘ | ✓ | ⊘ | PostgreSQL superuser |
| **replication_origin** | ✓ | ⊘ | ✓ | ⊘ | PostgreSQL superuser |
| **subscriptions** | ✓ | ⊘ | ✓ | ⊘ | PostgreSQL superuser |

**Environment-Dependent Collectors (N/A in some environments):**

| Collector | S1 | S2 | S3 | S4 | Reason |
|-----------|----|----|----|----|--------|
| hypervisor | N/A | N/A | N/A | N/A | Not running in VM |
| intel_pstate | N/A | N/A | N/A | N/A | CPU governor not available |
| localectl | N/A | N/A | N/A | N/A | systemd not running |
| lsdevmapper | N/A | N/A | N/A | N/A | No device mapper devices |
| machine-id | N/A | N/A | N/A | N/A | File not present |
| nfsiostat | N/A | N/A | N/A | N/A | No NFS mounts |

**Summary:**
- **Root access impact**: Enables kernel ring buffer access (`dmesg`) and PostgreSQL data directory file access
- **Superuser vs pg_monitor**: Superuser grants access to 3 additional catalog views
- **Note**: On real servers, most system collectors work without root. Only `dmesg`/`dmesg-t` and data directory files require elevated permissions.

**Scenarios:**
- **S1**: Root + PostgreSQL superuser (recommended - complete collection)
- **S2**: Root + PostgreSQL pg_monitor (production-safe alternative)
- **S3**: Non-root + PostgreSQL superuser
- **S4**: Non-root + PostgreSQL pg_monitor (most restricted)

## Code Style & Standards

### Formatting

```bash
# Check formatting
gofmt -l .

# Apply formatting
gofmt -w .
```

### Linting

```bash
# Run linter
golangci-lint run --timeout=5m
```

All code must pass:
- `gofmt` with no changes
- `golangci-lint` with no errors

### Testing

- All new functions must have unit tests
- Use table-driven tests where appropriate
- Mock external dependencies when needed
- Target 80%+ code coverage for new code

## Pull Request Process

1. **Fork** the repository
2. **Create a branch** for your feature (`git checkout -b feature/my-feature`)
3. **Make changes** following code style guidelines
4. **Add tests** for new functionality
5. **Run CI locally**: `./run-ci-local.sh`
6. **Commit changes** with clear messages
7. **Push to fork** and create pull request
8. **Address review feedback** promptly

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add iostat collector for disk I/O statistics
fix: handle missing pg_hba.conf silently
refactor: consolidate system tasks into task_definitions.go
test: add unit tests for skip flag validation
docs: update README with new collectors
```

### PR Requirements

- All CI checks must pass
- Code must be formatted (gofmt)
- Linting must pass (golangci-lint)
- Unit tests must pass
- Integration tests must pass (all 4 scenarios)
- Documentation updated if applicable

## Getting Help

- **Questions**: Open a GitHub issue with the "question" label
- **Bug Reports**: Open a GitHub issue with detailed reproduction steps
- **Feature Requests**: Open a GitHub issue describing the use case

## Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community guidelines.

## License

By contributing to radar, you agree that your contributions will be licensed under the same license as the project. See [LICENSE](LICENSE) for details.
