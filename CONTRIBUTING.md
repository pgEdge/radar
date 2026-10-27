# Contributing to radar

This document provides guidelines and information for developers and
contributors to radar. The document covers the design principles, the
development environment, the collector architecture, the test suite, and the
pull request process.

## Understanding the Design Philosophy

radar is built on the following core principles:

- The final `radar` binary should have zero dependencies; the binary should be
  self-contained and should not require any runtime dependencies (such as
  Python or Bash scripts) other than the tools it uses for collection (such as
  `iostat` and `psql`). This design makes the binary highly portable.
- radar is agentless; the tool runs on demand and does not require a background
  agent.
- radar aims for minimal impact, so the tool should be as lightweight as
  possible and should avoid any changes to the system it runs on. Data
  collection is read-only.
- radar fails gracefully; if a collector fails (for example, because a command
  is not installed or a file is not readable), the tool silently skips that
  collector and moves on. This approach ensures that radar collects as much
  data as possible without stopping on minor issues.
- radar is comprehensive by default and aims to collect a wide array of data
  that is commonly needed to diagnose system and database performance issues.
- radar keeps collection focused on simple, targeted data gathering of raw
  metrics and configuration only.
- radar uses a streaming architecture that writes data directly to the ZIP
  archive, which keeps memory usage minimal even with large datasets.

## Interpreting Exit Codes

radar returns an exit code that indicates the result of each run. The following
table describes the exit codes:

| Exit Code | Meaning |
|-----------|---------|
| `0` | The run succeeded. |
| `1` | A usage error occurred, such as missing required flags. |
| `3` | A collection or archive error occurred. |
| `4` | radar collected no data. |

A PostgreSQL connection failure during normal collection does not cause radar
to exit with an error code. Instead, radar continues with system-only
collection and reports the connection issue to the user.

## Setting Up the Development Environment

This section describes the tools that radar development uses, how to build
radar from source, and how to work with platform-specific code.

### Installing the Prerequisites

The development environment uses the following tools:

- Go 1.25.12 or later is required; the `go` directive in `go.mod` sets this
  minimum.
- Git is required for version control.
- golangci-lint v2.12.2 is used for linting; the `run-ci-local.sh` script
  installs this pinned version itself, as
  [Installing golangci-lint](#installing-golangci-lint) describes.
- Docker is optional but recommended for integration testing.

The `go version` command prints the installed Go version. The
[Go installation guide](https://go.dev/doc/install) describes how to install a
current release.

### Building from Source

The radar source code is hosted in the
[radar repository](https://github.com/pgEdge/radar) on GitHub. The following
commands clone the repository, build the binary, and run the binary to display
the help:

```bash
# Clone the repository
git clone https://github.com/pgEdge/radar.git
cd radar

# Build the binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .

# Run the binary
./radar --help
```

### Developing Across Platforms

radar uses Go build tags for platform-specific code. The collectors are defined
in the following files:

- `system_tasks_linux.go` contains the Linux-only system collectors
  (`//go:build linux`).
- `system_tasks_darwin.go` contains the macOS-only system collectors
  (`//go:build darwin`).
- `system_tasks_shared.go` contains the cross-platform system collectors
  (`//go:build linux || darwin`).
- `postgres_tasks.go` contains the PostgreSQL collectors, which are
  platform-independent.

#### Building for Specific Platforms

The following commands build radar for the current platform, cross-compile
radar for Linux or macOS, and build for both platforms in one command:

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

Follow these steps to add a new system collector:

1. Determine the platform compatibility of the collector, and choose the file
   as follows:

   - If the collector works on all Unix-like systems, add the collector to
     `system_tasks_shared.go`.
   - If the collector is Linux-specific (it uses `/proc`, `/sys`, systemd, or
     similar), add the collector to `system_tasks_linux.go`.
   - If the collector is macOS-specific (it uses `sysctl`, `system_profiler`,
     or similar), add the collector to `system_tasks_darwin.go`.

2. Add a `SimpleCommandTask` or `SimpleFileTask` entry to the matching slice in
   that file: `sharedCommandTasks` or `sharedFileTasks` in
   `system_tasks_shared.go`, or `systemCommandTasks` or `systemFileTasks` in a
   platform-specific file.
3. Test the collector on each target platform, and ensure that the command or
   file exists and produces the expected output.
4. Update the tests if continuous integration (CI) should verify the collector.

The [Adding System Collectors](#adding-system-collectors) section shows a real
entry of each type.

#### Testing Locally on macOS

On macOS, you can test radar locally without Docker. The following commands
install PostgreSQL with Homebrew, install the pg_statviz extension from source
as the CI `macos-test` job does, create a test database, build and run radar,
and list the archive contents. The paths assume the Homebrew prefix
`/opt/homebrew`, which the CI job also uses:

```bash
# Install PostgreSQL via Homebrew
brew install postgresql@18
export PATH="/opt/homebrew/opt/postgresql@18/bin:$PATH"
brew services start postgresql@18

# Install the pg_statviz extension
git clone https://github.com/vyruss/pg_statviz.git /tmp/pg_statviz
(cd /tmp/pg_statviz && make install PG_CONFIG=/opt/homebrew/opt/postgresql@18/bin/pg_config)

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

### Installing golangci-lint

The `run-ci-local.sh` script installs its own pinned golangci-lint, version
v2.12.2, which is the version that CI installs. The script downloads the linter
into the `.ci-tools/` directory, which Git ignores, whenever that copy is
missing or reports a different version. The script never uses a golangci-lint
from your `PATH`, because a different version does not read the v2-schema
`.golangci.yml` file of the repository. Running the CI script therefore needs
no separate installation of golangci-lint.

To run the linter outside the CI script, the following commands install the
same pinned version into `.ci-tools/` with the installer that `run-ci-local.sh`
uses:

```bash
# Install the pinned golangci-lint into .ci-tools/
mkdir -p .ci-tools
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v2.12.2/install.sh \
    | sh -s -- -b .ci-tools v2.12.2
```

## Adding Collectors

radar is designed to be easily extensible with new collectors. Most collectors
are entries in a task slice, and a builder function converts each entry into a
`CollectionTask`. This section shows the entry types, a real entry of each
kind, and the testing requirements for a new collector.

### Adding System Collectors

A system collector is either a `SimpleCommandTask`, which runs a command, or a
`SimpleFileTask`, which reads a file. The following definitions from
`system_tasks_shared.go` show both types:

```go
// SimpleCommandTask defines a shell command-based collection
type SimpleCommandTask struct {
    Name        string
    ArchivePath string
    Command     string
    Args        []string
}

// SimpleFileTask defines a file read-based collection
type SimpleFileTask struct {
    Name        string
    ArchivePath string
    Path        string
}
```

The `buildCommandTasks` function wraps each command entry with
`execCommandCollector`, and the `buildFileTasks` function wraps each file entry
with `readFileCollector`. The following entry from `sharedCommandTasks` defines
a command-based collector that runs `df -h`:

```go
// In system_tasks_shared.go (sharedCommandTasks)
{
    Name:        "df",
    ArchivePath: "system/diskspace.out",
    Command:     "df",
    Args:        []string{"-h"},
},
```

The following entry from `sharedFileTasks` defines a file-based collector that
reads `/etc/hosts`:

```go
// In system_tasks_shared.go (sharedFileTasks)
{
    Name:        "hosts",
    ArchivePath: "system/hosts.out",
    Path:        "/etc/hosts",
},
```

Command collectors that apply only inside a container belong in
`containerCommandTasks` in `system_tasks_linux.go`; radar includes those
collectors only when `isContainer()` returns true. A collector that needs its
own logic, such as the PgBouncer collectors in `pgbouncer.go`, defines a
`CollectionTask` directly with its own collector function.

### Adding PostgreSQL Collectors

A PostgreSQL query collector is a `SimpleQueryTask`, and a PostgreSQL
configuration file collector is a `SimpleConfigFileTask`. The following
definitions from `postgres_tasks.go` show both types:

```go
// SimpleQueryTask defines a PostgreSQL query-based collection
type SimpleQueryTask struct {
    Name        string
    ArchivePath string
    Query       string
}

// SimpleConfigFileTask defines a PostgreSQL config file collection
type SimpleConfigFileTask struct {
    Name        string
    ArchivePath string
    Filename    string
}
```

The `buildQueryTasks` function wraps each instance-level query entry with
`pgQueryCollector`. The following entry from `postgresQueryTasks` defines an
instance-level query collector:

```go
// In postgres_tasks.go (postgresQueryTasks)
{
    Name:        "archiver",
    ArchivePath: "postgresql/archiver.tsv",
    Query:       "SELECT * FROM pg_stat_archiver",
},
```

radar runs each per-database query in every database that allows connections,
except `template0` and `template1`, and connects to each database in turn. The
`ArchivePath` of a per-database entry contains a `%s` placeholder, which radar
replaces with the database name. The following entry from
`perDatabaseQueryTasks` defines a per-database query collector:

```go
// In postgres_tasks.go (perDatabaseQueryTasks)
{
    Name:        "extensions",
    ArchivePath: "databases/%s/extensions.tsv",
    Query:       "SELECT * FROM pg_extension ORDER BY extname",
},
```

The pg_statviz collectors in `pgStatvizQueryTasks` and the Spock collectors in
`spockQueryTasks` are also per-database query collectors. The following entry
from `pgStatvizQueryTasks` defines a pg_statviz collector:

```go
// In postgres_tasks.go (pgStatvizQueryTasks)
{
    Name:        "pg_statviz_buf",
    ArchivePath: "pg_statviz/%s/buf.tsv",
    Query:       "SELECT * FROM pgstatviz.buf ORDER BY snapshot_tstamp",
},
```

The `buildConfigFileTasks` function wraps each configuration file entry with
`collectPGConfigFile`, which reads the named file from the PostgreSQL data
directory. The following entry from `postgresConfigFileTasks` defines a
configuration file collector:

```go
// In postgres.go (postgresConfigFileTasks)
{
    Name:        "pg_hba.conf",
    ArchivePath: "postgresql/pg_hba.conf",
    Filename:    "pg_hba.conf",
},
```

### Meeting the Testing Requirements

When you add a collector, meet the following requirements:

- Update the unit tests in `radar_test.go`.
- Update the integration tests if the collector depends on permissions.
- Test all six integration scenarios, if applicable.
- Ensure that radar silently handles missing tools and permissions.

## Architecture

This section describes how the source code is organized, how collection tasks
are defined, and how radar handles collector errors.

### Locating the Source Files

The following tree shows the Go source files and the main scripts and documents
in the repository:

```text
radar/
├── radar.go                  # Main entry point, CLI parsing, orchestration, collector helpers
├── radar_test.go             # Unit tests
├── system.go                 # System task assembly
├── system_tasks_shared.go    # Cross-platform system collector definitions and task builders
├── system_tasks_linux.go     # Linux-only system and container collector definitions
├── system_tasks_darwin.go    # macOS-only system collector definitions
├── container_linux_test.go   # Container detection unit tests
├── postgres.go               # PostgreSQL task assembly, config file and per-database collection
├── postgres_tasks.go         # PostgreSQL query collector definitions and task builders
├── postgres_test.go          # PostgreSQL unit tests
├── pgbouncer.go              # PgBouncer collectors
├── pgbouncer_test.go         # PgBouncer unit tests
├── test-radar.sh             # Integration test script
├── run-ci-local.sh           # Local CI/CD script
├── Dockerfile                # Docker container for testing
├── docs/data.md              # Reference for the data collected
└── README.md                 # User documentation
```

### Defining Collection Tasks

radar uses a table-driven architecture in which every collector runs as a
`CollectionTask`. The following definition from `radar.go` shows the fields of
a task:

```go
// CollectionTask defines a single data collection task
type CollectionTask struct {
    Category    string // "system", "postgresql", "database"
    Name        string // Descriptive name for logging
    ArchivePath string // Path within ZIP archive
    Collector   func(*Config, io.Writer) error
}
```

The collector function writes its output to the `io.Writer`, and radar creates
the archive entry at `ArchivePath` on the first write. A collector that writes
nothing leaves no entry in the archive.

The following table describes the helper functions in `radar.go` that return
collector functions:

| Helper | Description |
|--------|-------------|
| `execCommandCollector(name string, args ...string) func(*Config, io.Writer) error` | The returned collector runs a system command and writes the combined `stdout` and `stderr` output. |
| `readFileCollector(path string) func(*Config, io.Writer) error` | The returned collector reads a file and writes the raw file contents. |
| `pgQueryCollector(db *sql.DB, query string) func(*Config, io.Writer) error` | The returned collector executes a PostgreSQL query and streams the results as tab-separated values (TSV). |

The following table describes every task slice, the file that defines the
slice, and the type of its entries:

| Slice | File | Entry Type | Contents |
|-------|------|------------|----------|
| `systemCommandTasks` | `system_tasks_linux.go`, `system_tasks_darwin.go` | `SimpleCommandTask` | Platform-specific system command collectors. |
| `systemFileTasks` | `system_tasks_linux.go`, `system_tasks_darwin.go` | `SimpleFileTask` | Platform-specific system file collectors. |
| `sharedCommandTasks` | `system_tasks_shared.go` | `SimpleCommandTask` | Cross-platform system command collectors. |
| `sharedFileTasks` | `system_tasks_shared.go` | `SimpleFileTask` | Cross-platform system file collectors. |
| `containerCommandTasks` | `system_tasks_linux.go` | `SimpleCommandTask` | Command collectors that run only inside a container. |
| `postgresQueryTasks` | `postgres_tasks.go` | `SimpleQueryTask` | PostgreSQL instance-level query collectors. |
| `postgresConfigFileTasks` | `postgres.go` | `SimpleConfigFileTask` | PostgreSQL configuration file collectors. |
| `perDatabaseQueryTasks` | `postgres_tasks.go` | `SimpleQueryTask` | Per-database query collectors. |
| `pgStatvizQueryTasks` | `postgres_tasks.go` | `SimpleQueryTask` | Per-database pg_statviz collectors. |
| `spockQueryTasks` | `postgres_tasks.go` | `SimpleQueryTask` | Per-database Spock collectors. |

The `getSystemTasks` function in `system.go` assembles the system tasks,
including the PgBouncer tasks that `getPgBouncerTasks` in `pgbouncer.go`
returns. The `getPostgreSQLTasks` and `generateDatabaseTasks` functions in
`postgres.go` assemble the instance-level and per-database PostgreSQL tasks.

### Handling Errors

radar tracks three states for each collector:

- Collected (`✓`) means that the collector successfully collected data.
- Skipped (`⊘`) means that the tool or file is unavailable, or that the
  permissions are insufficient.
- Failed means that a real error occurred; radar reports the error to the user.

The custom error type `SkipError` signals an intentional skip:

```go
type SkipError struct {
    Reason string
}
```

## Running Tests

This section describes how to run the unit tests, the local CI script, and the
Docker-based integration tests.

### Running Unit Tests

The following commands run all unit tests or a specific test:

```bash
# Run all unit tests
go test -v ./...

# Run specific test
go test -v -run TestSkipFlagValidation
```

### Running the Local CI Script

The following command runs the complete CI/CD test suite locally:

```bash
./run-ci-local.sh
```

The script performs the following steps in order:

1. The script pins the Go toolchain by setting `GOTOOLCHAIN=go1.25.12`, so that
   the compiler, the standard library, and `gofmt` match CI regardless of the
   installed Go version.
2. The script checks the code formatting with the `gofmt` of the pinned
   toolchain.
3. The script lints the code with its own pinned golangci-lint in `.ci-tools/`,
   as [Installing golangci-lint](#installing-golangci-lint) describes.
4. The script runs the unit tests with `go test`.
5. The script builds the binary.
6. The script runs the integration tests in Docker with PostgreSQL 18 and
   pg_statviz.

The script also creates a timestamped log file named `ci-YYYYMMDD-HHMMSS.log`.

### Testing with Docker

The following commands build and run the integration tests in a Debian
container, run radar manually without PostgreSQL collection, and list the
contents of the archive:

```bash
# Build and test in Debian container
docker build -t radar-test .
docker run --rm radar-test

# Manual testing after build
./radar --skip-postgres -v

# Verify ZIP contents
unzip -l radar-*.zip
```

## Understanding Continuous Integration

The project uses GitHub Actions for CI/CD. For details, see
[.github/workflows/ci.yml](.github/workflows/ci.yml).

### Reviewing the Integration Test Scenarios

The integration test runs six scenarios: the four permission combinations,
followed by certificate authentication and Generic Security Service Application
Program Interface (GSSAPI) authentication. The following table describes each
scenario and the collection that the scenario expects:

| Scenario | Name | Description | Expected Collection |
|----------|------|-------------|---------------------|
| S1 | Root and PostgreSQL superuser | radar has full system access, with all commands available to root, and full PostgreSQL access to all views and configuration files. This scenario is recommended because the collection is complete. | About 71 system collectors, about 55 PostgreSQL collectors, and pg_statviz data. |
| S2 | Root and PostgreSQL pg_monitor role | radar has full system access with root privileges, and monitoring-level PostgreSQL access to most views with limited configuration access. Some collectors are unavailable (for example, `pg_hba_file_rules` and `subscriptions`). This scenario is the production-safe alternative. | About 71 system collectors, about 51 PostgreSQL collectors, and pg_statviz data. |
| S3 | Non-root and PostgreSQL superuser | radar has limited system access, because some commands fail without root, and full PostgreSQL access to all views and configuration files. Some system collectors are unavailable (for example, `ifconfig` and `sysctl`). | About 68 system collectors, about 51 PostgreSQL collectors, and pg_statviz data. |
| S4 | Non-root and PostgreSQL pg_monitor role | radar runs as a non-root user with monitoring-level PostgreSQL access through the pg_monitor role. This scenario combines the limitations of both non-root and pg_monitor, and is the most restricted. | About 68 system collectors, about 47 PostgreSQL collectors, and pg_statviz data. |
| S5 | Certificate authentication | radar runs as root and connects as the pg_monitor role with a client certificate over TLS. This scenario exercises `-sslmode verify-full` with `-sslcert`, `-sslkey` and `-sslrootcert`. | About 71 system collectors, about 51 PostgreSQL collectors, and pg_statviz data. |
| S6 | GSSAPI/Kerberos authentication | radar runs as root and connects as the pg_monitor role with a Kerberos ticket from `kinit`. Because `pg_hba.conf` requires `hostgssenc`, the connection must be GSSAPI-encrypted. | About 71 system collectors, about 51 PostgreSQL collectors, and pg_statviz data. |

All scenarios verify that radar handles permission limitations and collects the
maximum available data.

### Comparing Collector Availability by Scenario

The tables in this section list only the collectors with meaningful differences
between permission scenarios; collectors that work identically across all
scenarios are omitted for clarity. The S1 to S4 columns refer to the scenarios
in the previous section.

The following table describes the permission-dependent collectors; Collected
means that radar collected the data, and Skipped means that radar skipped the
collector:

| Collector | S1 | S2 | S3 | S4 | Requirement |
|-----------|----|----|----|----|-------------|
| `dmesg` | Collected | Collected | Skipped | Skipped | Root system access |
| `dmesg-t` | Collected | Collected | Skipped | Skipped | Root system access |
| `pg_hba.conf` | Collected | Collected | Skipped | Skipped | Filesystem access to data directory |
| `pg_ident.conf` | Collected | Collected | Skipped | Skipped | Filesystem access to data directory |
| `postgresql.conf` | Collected | Collected | Skipped | Skipped | Filesystem access to data directory |
| `postgresql.auto.conf` | Collected | Collected | Skipped | Skipped | Filesystem access to data directory |
| `pg_hba_file_rules` | Collected | Skipped | Collected | Skipped | PostgreSQL superuser |
| `replication_origin` | Collected | Skipped | Collected | Skipped | PostgreSQL superuser |
| `subscriptions` | Collected | Skipped | Collected | Skipped | PostgreSQL superuser |

The following table describes the environment-dependent collectors, which are
not applicable (N/A) in some environments:

| Collector | S1 | S2 | S3 | S4 | Reason |
|-----------|----|----|----|----|--------|
| `hypervisor` | N/A | N/A | N/A | N/A | Not running in VM |
| `intel_pstate` | N/A | N/A | N/A | N/A | CPU governor not available |
| `localectl` | N/A | N/A | N/A | N/A | systemd not running |
| `lsdevmapper` | N/A | N/A | N/A | N/A | No device mapper devices |
| `machine-id` | N/A | N/A | N/A | N/A | File not present |
| `nfsiostat` | N/A | N/A | N/A | N/A | No NFS mounts |

The availability tables show the following effects of each permission:

- Root access enables kernel ring buffer access (`dmesg`) and access to the
  files in the PostgreSQL data directory.
- Connecting as a superuser instead of as pg_monitor grants access to three
  additional catalog views.
- On real servers, most system collectors work without root; only `dmesg`,
  `dmesg-t`, and the data directory files require elevated permissions.

## Following the Code Style and Standards

All code must meet the formatting, linting, and testing standards that this
section describes.

### Formatting the Code

The following commands check the code formatting and apply the formatting in
place:

```bash
# Check formatting
gofmt -l .

# Apply formatting
gofmt -w .
```

### Linting the Code

The following command runs the pinned linter from `.ci-tools/` with the same
timeout that `run-ci-local.sh` uses:

```bash
# Run linter
.ci-tools/golangci-lint run --timeout=5m
```

All code must pass:

- `gofmt` with no changes.
- `golangci-lint` with no errors.

### Writing Tests

New code follows these testing standards:

- All new functions must have unit tests.
- Use table-driven tests where appropriate.
- Mock external dependencies when needed.
- Target 80% or higher code coverage for new code.

## Submitting a Pull Request

Follow these steps to submit a change to radar:

1. Fork the [radar repository](https://github.com/pgEdge/radar).
2. Create a branch for your feature (`git checkout -b feature/my-feature`).
3. Make your changes, following the code style guidelines.
4. Add tests for new functionality.
5. Run the CI locally with `./run-ci-local.sh`.
6. Commit the changes with clear messages.
7. Push the branch to your fork, and create a pull request.
8. Address review feedback promptly.

### Writing Commit Messages

Use clear, descriptive commit messages, as in the following examples:

```text
feat: add iostat collector for disk I/O statistics
fix: handle missing pg_hba.conf silently
refactor: consolidate system tasks into task_definitions.go
test: add unit tests for skip flag validation
docs: update README with new collectors
```

### Meeting the Pull Request Requirements

A pull request must meet the following requirements:

- All CI checks must pass.
- The code must be formatted with `gofmt`.
- Linting with `golangci-lint` must pass.
- The unit tests must pass.
- The integration tests must pass in all six scenarios.
- The documentation must be updated, if applicable.

## Getting Help

Use the [radar issue tracker](https://github.com/pgEdge/radar/issues) on GitHub
for questions, bug reports, and feature requests, as follows:

- To ask a question, open a GitHub issue with the `question` label.
- To report a bug, open a GitHub issue with detailed reproduction steps.
- To request a feature, open a GitHub issue that describes the use case.

## Following the Code of Conduct

The [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) file describes the community
guidelines for the project.

## License

By contributing to radar, you agree that your contributions will be licensed
under the same license as the project. For details, see
[LICENCE.md](LICENCE.md).
