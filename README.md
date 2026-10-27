# radar

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/2645bf9703e7474ba14d618707231276)](https://app.codacy.com/gh/pgEdge/radar?utm_source=github.com&utm_medium=referral&utm_content=pgEdge/radar&utm_campaign=Badge_Grade)

radar is an agentless diagnostic data collection tool for PostgreSQL and for
the system that hosts it.

## Table of Contents

The radar documentation consists of the following pages:

- [Introduction](docs/index.md)
    - [Installing radar](docs/index.md#installing-radar)
    - [Configuring radar](docs/index.md#configuring-radar)
    - [Using radar](docs/index.md#using-radar)
    - [Reviewing the Collected Data](docs/index.md#reviewing-the-collected-data)
- [Data Reference](docs/data.md)
- [Changelog](docs/changelog.md)
- [License](docs/LICENCE.md)

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
  [data reference](docs/data.md) notes.
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

## Configuring radar

radar takes its settings from command-line options and from the standard
PostgreSQL environment variables; radar has no configuration file.

### Setting the Connection Defaults

The `-h`, `-U`, and `-d` options are optional, and radar resolves each value as
follows:

- `-h` falls back to `PGHOST`, and then to `localhost` over TCP; radar does not
  default to a Unix-domain socket.
- `-U` falls back to `PGUSER`, then to the name of the current operating system
  user, and then to `postgres` when radar cannot determine that user.
- `-d` falls back to `PGDATABASE`, and then to `postgres`; with
  `--skip-system`, radar requires a database name from `-d` or `PGDATABASE`.

### Setting Environment Variables

The following table describes the environment variables that radar reads:

| Variable | Description |
|----------|-------------|
| `PGHOST` | Specifies the PostgreSQL host. |
| `PGPORT` | Specifies the PostgreSQL port. |
| `PGUSER` | Specifies the PostgreSQL user name. |
| `PGPASSWORD` | Specifies the PostgreSQL password. |
| `PGPASSFILE` | Names the password file that the driver reads when no password is supplied; the default is `~/.pgpass`. |
| `PGDATABASE` | Specifies the PostgreSQL database name. |
| `PGSSLMODE` | Specifies the SSL mode: `prefer`, `disable`, `require`, `verify-ca`, or `verify-full`. radar rejects any other value. |
| `PGSSLCERT` | Specifies the client SSL certificate file. |
| `PGSSLKEY` | Specifies the client SSL private key file. |
| `PGSSLROOTCERT` | Specifies the SSL root (CA) certificate file. |
| `KRB5CCNAME` | Names the Kerberos credential cache for GSSAPI authentication; radar reads only a `FILE:` cache, and the default is `/tmp/krb5cc_<uid>`. |
| `KRB5_CONFIG` | Names the Kerberos configuration file; the default is `/etc/krb5.conf`. |

The pgx driver also applies the other libpq environment variables that radar
does not set itself, such as `PGAPPNAME` and `PGCONNECT_TIMEOUT`.

### Choosing an Authentication Method

The following table describes how to configure radar for each authentication
method:

| Method | Configuration |
|--------|---------------|
| Password | Supply the user with `-U` and the password with `PGPASSWORD`, or rely on the operating system user defaults. |
| LDAP | The server performs LDAP authentication, so the client needs no changes; supply the user name and password as for password authentication. |
| Certificate | Run radar with `--sslmode verify-full --sslcert client.crt --sslkey client.key --sslrootcert ca.crt`. |
| GSSAPI (Kerberos) | Obtain a Kerberos ticket with `kinit` as the user that runs radar; radar reads the ticket from a file credential cache. With a KCM or KEYRING cache, run `kinit -c FILE:/tmp/krb5cc_$(id -u)` instead. radar supports GSSAPI authentication but not GSSAPI encryption, so a `hostgssenc` line in `pg_hba.conf` does not match radar's connections. |

### Choosing Permissions

The privileges of the operating system user and of the PostgreSQL role
determine how much data radar can collect.

#### Running as Root with a Superuser

For a complete collection, run radar as root with a PostgreSQL superuser
account; this is the recommended combination. The following command runs a
collection as root with the `postgres` superuser:

```bash
PGPASSWORD='postgres_password' radar -d mydatabase -U postgres
```

This combination gives radar:

- full system-level access to operating system metrics, logs, and configuration
  files.
- complete access to the PostgreSQL catalog, including all views, tables, and
  system information.
- the most diagnostic data for troubleshooting.

#### Running as Root with the pg_monitor Role

Where your policy restricts superuser access in production, grant the
`pg_monitor` role to a dedicated user instead. The `pg_monitor` role provides
most of the diagnostic data without superuser privileges.

The following commands create a monitoring user as the `postgres` operating
system user, and then run a collection as root:

```bash
# As the postgres user, create a monitoring user with the pg_monitor role
psql -c "CREATE USER radaruser WITH PASSWORD 'secure_password';"
psql -c "GRANT pg_monitor TO radaruser;"

# As root, run a collection
PGPASSWORD='secure_password' radar -d mydatabase -U radaruser
```

The `pg_monitor` role can read nearly every source that radar queries. A few
replication catalog views and `pg_hba_file_rules` remain unreadable, and radar
reports a permission denial for each affected collector. The run still
completes and writes an archive, and radar still reads `pg_hba.conf` from the
data directory.

#### Running with Limited Permissions

radar can run as a non-root user with limited PostgreSQL permissions. radar
then reports an error for each system file that it cannot read and each command
that fails, and some PostgreSQL catalog queries fail; the run continues and
writes an archive.

## Using radar

Each radar invocation runs one collection, writes a single ZIP archive, and
exits.

### Running a Collection

The following commands collect both kinds of data, system data only, and
PostgreSQL data only:

```bash
# Collect system and PostgreSQL data (the default)
radar -d mydatabase -U databaseuser

# Collect system data only
radar --skip-postgres

# Collect PostgreSQL data only
radar -d mydatabase -U databaseuser --skip-system
```

The `--skip-system` option also skips the PgBouncer collectors, which radar
treats as system collectors.

radar prints the archive name and its size in kilobytes when the collection
finishes:

```text
$ radar -d mydatabase
Collecting diagnostic data...
✓ Archive created: radar-hostname-20260115-133700.zip (1234 KB)
```

### Reviewing the Command-Line Options

The `radar --help` command prints the following usage summary:

```text
Usage: radar [options]

GLOBAL OPTIONS:
   --help         show help
   --version, -V  print the version

Options:
  -U string
    	database user
  -d string
    	database name
  -data-dir string
    	PostgreSQL data directory
  -exclude string
    	comma-separated task names to skip (e.g. -exclude bloat,stat_ssl)
  -h string
    	database host
  -include string
    	comma-separated default-disabled task names to enable (e.g. pgstattuple, disabled by default)
  -p int
    	database port (default 5432)
  -pgbouncer-conf string
    	PgBouncer configuration directory (default: /etc/pgbouncer, /usr/local/etc/pgbouncer)
  -skip-postgres
    	skip PostgreSQL data collection
  -skip-system
    	skip system data collection
  -sslcert string
    	client SSL certificate file
  -sslkey string
    	client SSL key file
  -sslmode string
    	SSL mode (prefer, disable, require, verify-ca, verify-full) (default "prefer")
  -sslrootcert string
    	SSL root certificate file
  -v	verbose output (summary)
  -vv
    	very verbose output (detailed)
```

### Excluding Collectors

The `--exclude` option skips the collectors that you name, and the `--include`
option enables the collectors that are disabled by default; `--include` takes
precedence over `--exclude`.

Run radar with `-vv` to list each collector by name; for a per-database
collector, the name after the database prefix is the name that `--exclude`
accepts:

```text
✓ stat_ssl
✓ postgres/tables
```

The following command skips the `bloat` and `stat_ssl` collectors:

```bash
radar -d mydatabase -U databaseuser --exclude bloat,stat_ssl
```

### Reading the Archive

radar writes all data to a single ZIP file named
`radar-{hostname}-{timestamp}.zip` in the current directory. Only the owner can
read the archive, and radar refuses to overwrite an existing file. The
following tree shows the layout of the archive:

```text
radar-hostname-20260115-133700.zip
├── radar.out            (radar version and commit)
├── system/              (Linux system data)
│   ├── lsblk.out
│   ├── mount.out
│   ├── diskspace.out
│   ├── proc/
│   │   ├── meminfo.out
│   │   └── cpuinfo.out
│   └── ...
├── postgresql/          (PostgreSQL instance data)
│   ├── version.tsv
│   ├── databases.tsv
│   ├── configuration.tsv
│   ├── postgresql.conf
│   ├── pg_hba.conf
│   └── ...
├── databases/           (Per-database data)
│   ├── postgres/
│   │   ├── extensions.tsv
│   │   ├── tables.tsv
│   │   └── ...
│   └── mydb/
│       ├── extensions.tsv
│       └── ...
├── pg_statviz/          (Per-database pg_statviz time series, if installed)
│   └── mydb/
│       ├── snapshots.tsv
│       └── ...
├── spock/               (Per-database Spock replication state, if installed)
│   └── mydb/
│       ├── lag_tracker.tsv
│       └── ...
└── pgbouncer/           (PgBouncer configuration, if installed)
    ├── pgbouncer.ini
    └── files.tsv
```

The following table describes the file formats in the archive:

| Extension | Format |
|-----------|--------|
| `.tsv` | Tab-separated values with a header row; query results and directory listings use this format. |
| `.out` | Command output, with `stdout` and `stderr` combined, or a copy of a system file; `radar.out` records the radar version and commit. |
| `.conf` | The raw contents of a configuration file, such as `postgresql.conf`. |
| `.ini` | The `pgbouncer.ini` file, with the values outside the `[pgbouncer]` section redacted. |
| `.plist` | A macOS property list, such as the system version. |

### Understanding Performance

radar streams data directly into the ZIP file without buffering the data in
memory. radar runs the collectors sequentially with a small memory footprint,
and opens one connection per database for all of the queries in that database.
A complete collection typically takes seconds.

## Reviewing the Collected Data

radar collects metadata and statistics from the operating system, the
PostgreSQL instance, each database, and optional extensions and tools. The
[data reference](docs/data.md) describes every file in the archive.

### Collecting System Information

The following table describes the system data that radar collects:

| Area | Sources |
|------|---------|
| Block devices and storage | `blockdev`, `df`, `diskstats`, `fstab`, `io-queue-depth`, `io-schedulers`, `iostat`, `lsblk`, `ls`, `mount`, `mounts`, `nfsiostat`, `swaps` |
| CPU and performance | `cpuinfo`, `clocksource`, `energy_perf_bias`, `free`, `intel_pstate`, `loadavg`, `lscpu`, `mpstat`, `numactl`, `numastat`, `sar`, `scaling_available_governors`, `scaling_driver`, `scaling_governor`, `top`, `uptime`, `vmstat` |
| Memory and kernel | `dmesg`, `ipcs`, `meminfo`, `pressure/cpu`, `pressure/io`, `pressure/memory`, `sysctl`, `transparent_hugepage` |
| Operating system | `apt`, `dnf`, `dpkg`, `hostname`, `hosts`, `limits.conf`, `locale`, `locale.conf`, `localectl`, `lsmod`, `lspci`, `machine-id`, `os-release`, `ps`, `rpm`, `system-release`, `systemctl`, `systemd-detect-virt`, `timedatectl`, `tuned-adm`, `uname`, `yum` |
| Network | `ifconfig`, `ip`, `netstat`, `resolv.conf`, `ss` |
| Security | `fips-mode-setup`, `openssl`, `sestatus`, `update-crypto-policies` |
| Cgroup v2 | `cpu.max`, `cpu.weight`, `cpuset.cpus.effective`, `io.max`, `memory.current`, `memory.max`, `memory.stat`, `memory.swap.max`, `pids.current`, `pids.max` |
| Cgroup v1 | `cpu.cfs_period_us`, `cpu.cfs_quota_us`, `cpu.shares`, `cpuset.cpus`, `memory.limit_in_bytes`, `memory.stat`, `memory.usage_in_bytes` |
| Cloud and hardware identity | The DMI fields `bios_vendor`, `chassis_asset_tag`, `product_name`, and `sys_vendor`. |
| Container detection | radar detects a container when `/.dockerenv` exists, when `/proc/1/cgroup` names `docker`, `kubepods`, `containerd`, or `lxc`, or when `KUBERNETES_SERVICE_HOST` is set; inside a container, radar also collects the container environment variables and the Kubernetes namespace. |

### Collecting PostgreSQL Instance Data

The following table describes the instance-wide PostgreSQL data that radar
collects:

| Area | Sources |
|------|---------|
| Configuration and files | `pg_db_role_setting`, `pg_file_settings`, `pg_hba.conf`, `pg_hba_file_rules`, `pg_ident.conf`, `pg_settings`, `pg_tablespace`, `postgresql.auto.conf`, `postgresql.conf`, `recovery.conf`, `recovery.done` |
| Activity and monitoring | `pg_locks`, `pg_postmaster_start_time()`, `pg_prepared_xacts`, `pg_shmem_allocations`, `pg_stat_activity` |
| Statistics views | `pg_stat_archiver`, `pg_stat_bgwriter`, `pg_stat_checkpointer` (PG17+), `pg_stat_database_conflicts`, `pg_stat_io` (PG16+), `pg_stat_slru`, `pg_stat_ssl`, `pg_stat_statements` (if installed), `pg_stat_wal` (PG14+), `pg_stat_wal_receiver` |
| Replication and WAL | `pg_current_wal_lsn()`, `pg_replication_origin_status`, `pg_replication_slots`, `pg_stat_replication`, `pg_stat_replication_slots` (PG14+, spill counters), `pg_stat_subscription_stats` (PG15+ apply and sync error counters, PG18+ conflict counters), `pg_subscription` |
| Progress tracking | `pg_stat_progress_analyze`, `pg_stat_progress_basebackup`, `pg_stat_progress_cluster`, `pg_stat_progress_copy`, `pg_stat_progress_create_index`, `pg_stat_progress_vacuum` |
| Catalog | `pg_available_extension_versions`, `pg_database` (including the `datfrozenxid` and `datminmxid` wraparound headroom and `datconnlimit`), `pg_database_size()`, `pg_roles`, `pg_tablespace_size()`, `version()` |
| Control file | `pg_control_checkpoint()`, `pg_control_init()`, `pg_control_recovery()`, and `pg_control_system()` return the `pg_controldata` content as SQL. |
| Log directory | `pg_ls_logdir()` lists the name, size, and modification time of each file; radar never reads log contents. |

### Collecting Per-Database Data

The following table describes the data that radar collects from each database:

| Area | Sources |
|------|---------|
| Tables | `pg_class` and `pg_stat_all_tables`, including dead-tuple counters, vacuum and analyze timestamps, per-table `relfrozenxid` and `relminmxid` freeze ages, `reloptions`, `reltoastrelid`, and `relpersistence`. A separate freeze-age listing ranks tables by age rather than size, so the listing includes a small table near the wraparound threshold. |
| Indexes | `pg_index` and `pg_stat_all_indexes`, including validity, scan counters, and size. |
| Other schema objects | `pg_namespace`, `pg_operator`, `pg_sequences`, `pg_type` |
| Functions and procedures | `pg_proc` |
| Triggers and partitioning | `pg_inherits`, `pg_partitioned_table`, `pg_trigger` |
| Logical replication | `pg_publication`, `pg_publication_tables`, `pg_subscription_rel` |
| Extensions | `pg_extension`, `pg_language`, `pg_statistic_ext`, and `pgstattuple`; when `pgstattuple` is installed and `--include pgstattuple` enables the collector, `pgstattuple_approx()` measures bloat, with exact dead-tuple figures. |
| Statistics | `pg_stat_database`, including conflicts, deadlocks, temporary files, and the statistics reset time. |
| Bloat estimate | radar always collects a heuristic bloat estimate from `pg_stats`. |

### Collecting pg_statviz Data

When a database has the [pg_statviz](https://github.com/vyruss/pg_statviz)
extension installed, radar collects its time-series statistics from
`pgstatviz.blocking`, `pgstatviz.buf`, `pgstatviz.conf`, `pgstatviz.conn`,
`pgstatviz.db`, `pgstatviz.io`, `pgstatviz.lock`, `pgstatviz.repl`,
`pgstatviz.slru`, `pgstatviz.snapshots`, `pgstatviz.wait`, and `pgstatviz.wal`.

### Collecting PgBouncer Configuration

When [PgBouncer](https://www.pgbouncer.org/) is installed, radar collects the
`[pgbouncer]` section of `pgbouncer.ini` and a listing of the configuration
directory. radar replaces the values outside the `[pgbouncer]` section with
`<redacted by radar>`, because `[databases]` connection strings routinely
contain `password=`, and drops all comments. radar records `userlist.txt` as a
directory entry only; the file holds credentials, and radar never collects its
contents. radar reads the first of `/etc/pgbouncer` and
`/usr/local/etc/pgbouncer` that exists; use `--pgbouncer-conf` to name another
configuration directory.

### Collecting Spock Replication State

When a database has the [Spock](https://github.com/pgEdge/spock) extension
installed, radar collects the replication state from
`spock.channel_summary_stats`, `spock.lag_tracker`, `spock.local_node`,
`spock.local_sync_status`, `spock.node`, `spock.pii`, `spock.progress`,
`spock.replication_set`, `spock.replication_set_table`, `spock.subscription`,
and `spock.tables`.

radar also collects the counters, timestamps, conflict types, and apply error
messages from `spock.exception_log` and `spock.resolutions`. radar excludes the
jsonb and text images of the conflicting rows. radar does not collect
`spock.node_interface` at all, because its `if_dsn` column holds the node
password. radar keeps `error_message`; Spock stores no `DETAIL`, so the
offending key of a constraint violation is absent. An error whose primary
message embeds a value still includes that value.

### Protecting Sensitive Data

radar collects metadata and statistics, and does not collect table contents or
query result sets. A collected archive includes:

- database and table names, schemas, and object definitions.
- PostgreSQL configuration files and settings.
- system configuration and resource utilization metrics.
- active connection counts, database statistics, and query text.

> [!WARNING]
> A collected archive can contain passwords and other secrets. PostgreSQL
> configuration files and settings can hold connection passwords, such as a
> password in `primary_conninfo` on a standby, and query text and error
> messages can include values from user queries. Review the contents of an
> archive before you share it outside your organization.

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

## Contributing

We welcome your project contributions; for more information, see
[CONTRIBUTING.md](CONTRIBUTING.md). All contributors follow the
[Code of Conduct](CODE_OF_CONDUCT.md).

## Author

Jimmy Angelakos created radar.

## License

This project is licensed under the [PostgreSQL License](LICENCE.md).
