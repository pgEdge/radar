# radar

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/2645bf9703e7474ba14d618707231276)](https://app.codacy.com/gh/pgEdge/radar?utm_source=github.com&utm_medium=referral&utm_content=pgEdge/radar&utm_campaign=Badge_Grade)

## Table of Contents

The radar documentation consists of the following pages:

- [Introduction](docs/index.md) - Introduces radar and the data it collects.
- [Installing radar](docs/installing/index.md) - Describes the ways to install
  radar.
    - [Reviewing the Requirements](docs/installing/requirements.md) - Lists the
      platform and PostgreSQL requirements.
    - [Installing the pgEdge Package](docs/installing/pgedge_package.md) -
      Installs radar from the pgEdge Enterprise Postgres repository.
    - [Downloading a Pre-built Binary](docs/installing/prebuilt_binary.md) -
      Installs a release binary and verifies its checksum.
    - [Understanding Platform Differences](docs/installing/platform_differences.md) -
      Describes the collectors that run on Linux and on macOS.
- [Configuring radar](docs/configuring/index.md) - Describes radar's
  command-line options and environment variables.
    - [Setting the Connection Defaults](docs/configuring/connection_defaults.md) -
      Documents how radar resolves the host, user, and database.
    - [Setting Environment Variables](docs/configuring/environment_variables.md) -
      Lists the PostgreSQL environment variables that radar reads.
    - [Choosing an Authentication Method](docs/configuring/authentication.md) -
      Documents password, LDAP, certificate, and GSSAPI authentication.
    - [Choosing Permissions](docs/configuring/permissions.md) - Compares
      superuser, `pg_monitor`, and limited-permission collections.
- [Using radar](docs/using/index.md) - Describes running a collection and
  reading the results.
    - [Running a Collection](docs/using/running_a_collection.md) - Documents
      collecting system data, PostgreSQL data, or both.
    - [Reviewing the Command-Line Options](docs/using/command_line_options.md) -
      Lists every radar option.
    - [Excluding Collectors](docs/using/excluding_collectors.md) - Documents
      skipping and enabling individual collectors.
    - [Reading the Archive](docs/using/reading_the_archive.md) - Describes the
      archive layout and file formats.
    - [Understanding Performance](docs/using/performance.md) - Describes
      radar's memory and connection use.
- [Collector Reference](docs/collector_reference/index.md) - Describes the
  collected data and links to a reference page for each collector.
    - [Collecting Cross-Platform System Data](docs/collector_reference/cross_platform_system.md) -
      Documents the collectors that run on both Linux and macOS.
    - [Collecting Linux System Data](docs/collector_reference/linux_system.md) -
      Documents the Linux collectors, including cgroup and container detection.
    - [Collecting macOS System Data](docs/collector_reference/macos_system.md) -
      Documents the macOS collectors.
    - [Collecting PostgreSQL Instance Data](docs/collector_reference/postgresql_instance.md) -
      Documents the instance-level PostgreSQL collectors.
    - [Collecting Per-Database Data](docs/collector_reference/per_database.md) -
      Documents the schema and object data that radar collects per database.
    - [Collecting pg_statviz Data](docs/collector_reference/pg_statviz.md) -
      Documents the optional pg_statviz time-series collectors.
    - [Collecting Spock Data](docs/collector_reference/spock.md) - Documents
      the optional Spock replication collectors.
    - [Collecting PgBouncer Configuration](docs/collector_reference/pgbouncer.md) -
      Documents the optional PgBouncer configuration collectors.
- [Support & Resources](docs/support.md) - Lists radar's documentation, issue
  tracker, and security policy.
- [Changelog](docs/changelog.md) - Lists the changes in each radar release.
- [License](docs/LICENCE.md) - Provides the PostgreSQL License, under which
  pgEdge distributes radar.

radar is an agentless diagnostic data collection tool for PostgreSQL and for
the system that hosts it.

radar collects a snapshot of system and PostgreSQL diagnostic data for
troubleshooting and analysis. radar runs system commands and PostgreSQL
queries, and streams the output into a timestamped ZIP archive. A PostgreSQL
connection is optional; without one, radar collects system data only. When a
command, file, or view is unavailable, radar skips that collector and continues
the run.

pgEdge publishes radar as RPM and DEB packages and as pre-built binaries for
Linux and macOS, so you do not need to build radar yourself. The
[Installing radar](#installing-radar) section describes each method, starting
with the simplest.

## Installing radar

radar runs on Linux and macOS as a single static binary. You can install the
pgEdge package, download a pre-built binary, or build radar from source.

### Reviewing the Requirements

The requirements for radar include:

- a Linux or macOS host; radar uses standard utilities such as `df` and `ps`
  when they are present.
- PostgreSQL 12 or later; some collectors require a later version, as the
  [data reference](docs/collector_reference/index.md) notes.
- Git and Go 1.25.12 or later for a build from source; the `go` directive in
  [go.mod](go.mod) sets this minimum.

### Installing the pgEdge Package

pgEdge publishes radar as the `pgedge-radar` package in the pgEdge Enterprise
Postgres repository for Enterprise Linux, Debian, and Ubuntu. The repository
configuration enables signature checks with the pgEdge signing key. The package
installs the `radar` binary in `/usr/bin`.

Use the following steps to install the package:

1. Configure the pgEdge Enterprise Postgres repository by following the
   instructions for
   [Enterprise Linux](https://docs.pgedge.com/enterprise/el/configure-repo/) or
   for
   [Debian and Ubuntu](https://docs.pgedge.com/enterprise/debian/configure-repo/).

2. Install the package; on Enterprise Linux, run the following command:

    ```bash
    sudo dnf install -y pgedge-radar
    ```

    On Debian and Ubuntu, run the following commands:

    ```bash
    sudo apt-get update
    sudo apt-get install -y pgedge-radar
    ```

3. Confirm the installation by printing the radar version:

    ```bash
    radar --version
    ```

### Downloading a Pre-built Binary

Each [GitHub release](https://github.com/pgEdge/radar/releases) includes static
radar binaries for Linux and macOS on amd64 and arm64. The release also
includes a `checksums.txt` file with the SHA-256 checksum of each binary.

Use the following steps to install a pre-built binary on Linux:

1. Download the binary for your architecture and the checksum file from the
   latest release; the following commands download the amd64 binary:

    ```bash
    curl -fLO https://github.com/pgEdge/radar/releases/latest/download/radar-linux-amd64
    curl -fLO https://github.com/pgEdge/radar/releases/latest/download/checksums.txt
    ```

2. Verify the binary against the published checksum; the command prints `OK`
   after the name of a matching binary:

    ```bash
    sha256sum --check --ignore-missing checksums.txt
    ```

3. Make the binary executable and move the binary into a directory on your
   `PATH`:

    ```bash
    chmod +x radar-linux-amd64
    sudo mv radar-linux-amd64 /usr/local/bin/radar
    ```

On macOS, download `radar-darwin-arm64` or `radar-darwin-amd64` instead. Run
`shasum -a 256` on the binary, and compare the output with the matching line in
`checksums.txt`.

### Building from Source

Building radar from source requires Git and Go 1.25.12 or later. The
[Go installation guide](https://go.dev/doc/install) describes how to install a
current Go release on any platform. On Ubuntu 22.04, 24.04, and 26.04, the
`golang-1.25` package provides a suitable release; the following commands
install the package and add its `go` command to your `PATH`:

```bash
sudo apt-get install -y golang-1.25
export PATH=/usr/lib/go-1.25/bin:$PATH
```

Use the following steps to build radar from source:

1. Confirm that `go version` reports Go 1.25.12 or later:

    ```bash
    go version
    ```

2. Clone the pgEdge radar repository and change into its directory:

    ```bash
    git clone https://github.com/pgEdge/radar.git
    cd radar
    ```

3. Build a static binary for your platform:

    ```bash
    CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .
    ```

4. Optionally, move the binary into a directory on your `PATH`:

    ```bash
    sudo mv radar /usr/local/bin/
    ```

The [CONTRIBUTING.md](CONTRIBUTING.md) file describes cross-compilation and the
development setup.

### Understanding Platform Differences

Some system collectors are specific to one platform. On macOS, radar skips the
Linux-only collectors for systemd, SELinux, and the `/proc` and `/sys`
filesystems. radar uses macOS collectors such as `sysctl`, `system_profiler`,
`diskutil`, and `launchctl` instead.

On macOS, Homebrew can install PostgreSQL if the host needs it:

```bash
brew install postgresql@18
```

## Documentation

The [radar documentation](https://docs.pgedge.com/radar/) on the pgEdge
documentation site covers installation, usage, and the data reference; the
`docs` directory holds the source of that documentation. To preview the
documentation locally, install MkDocs Material and start the MkDocs development
server from the repository root:

```bash
pip install mkdocs-material
mkdocs serve
```

## Support & Resources

pgEdge provides documentation, an issue tracker, and a security policy for
radar. For more information, visit [docs.pgedge.com](https://docs.pgedge.com).

To report an issue with the software, visit
[GitHub Issues](https://github.com/pgEdge/radar/issues). To report a security
vulnerability, follow the
[security policy](https://github.com/pgEdge/radar/security/policy).

## For Developers

We welcome your project contributions; for more information, see
[CONTRIBUTING.md](CONTRIBUTING.md). All contributors follow the
[Code of Conduct](CODE_OF_CONDUCT.md).

### Author

Jimmy Angelakos created radar.

## License

This project is licensed under the [PostgreSQL License](LICENCE.md).
