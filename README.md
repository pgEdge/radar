# radar

> Agentless, zero-dependency diagnostic data collection tool for PostgreSQL and system metrics.

## What it Does

Collects a comprehensive snapshot of system and PostgreSQL metrics for troubleshooting and analysis:

- Runs system commands and PostgreSQL queries to collect diagnostic information
- Archives all collected data into a timestamped ZIP file
- Handles unavailable collectors gracefully (no errors if commands/files are missing)

## How it Works

1. Connects to PostgreSQL (optional - can collect system data only)
2. Executes collection tasks:
   - System tasks: shell commands, file reads
   - PostgreSQL tasks: SQL queries, config file reads
3. Streams output directly to ZIP archive
4. Skips unavailable collectors without failing

## Installation

radar supports Linux and macOS.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/vyruss/radar.git
cd radar

# Build for your platform
CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .

# Optionally move to system path
sudo mv radar /usr/local/bin/
```

### Platform-Specific Notes

**macOS**: Install PostgreSQL via Homebrew if needed:
```bash
brew install postgresql@18
```

**Collectors**: Some system collectors are platform-specific. Linux-only collectors (systemd, SELinux, /proc, /sys filesystems) will be skipped on macOS. macOS uses platform-specific collectors (sysctl, system_profiler, diskutil, launchctl, etc.) instead.

For detailed build instructions and cross-compilation, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Quick Start

```bash
# Collect both system and PostgreSQL data
./radar -d mydatabase

# System data only
./radar --skip-postgres

# PostgreSQL data only
./radar -d mydatabase --skip-system
```

## Permissions & Security

### Recommended: Root + PostgreSQL Superuser

For complete diagnostic collection, run radar as **root** with a PostgreSQL **superuser** account:

```bash
# As root, run collection with postgres superuser
PGPASSWORD='postgres_password' ./radar -d mydatabase -U postgres
```

This combination provides:
- Full system-level access to collect OS metrics, logs, and configuration files
- Complete PostgreSQL catalog access (all views, tables, and system information)
- Maximum diagnostic data for troubleshooting

### Alternatives

**Root + pg_monitor Role**: For production environments where superuser access is restricted, use the `pg_monitor` role instead. This provides most diagnostic data while using a non-superuser account:

```bash
# As postgres user, create a monitoring user with pg_monitor role
psql -c "CREATE USER radaruser WITH PASSWORD 'secure_password';"
psql -c "GRANT pg_monitor TO radaruser;"

# As root, run collection
PGPASSWORD='secure_password' ./radar -d mydatabase -U radaruser
```

Note: The pg_monitor role collects 68 of 71 collectors. Missing data includes some replication catalog views and `pg_hba_file_rules` (though the actual `pg_hba.conf` file is still collected via filesystem access).

**Limited Permissions**: radar can run as a non-root user with limited PostgreSQL permissions. Some system collectors will be skipped, and some PostgreSQL catalog queries may fail gracefully.

### Data Privacy

radar collects **metadata only** - no user data, query results, or table contents. However, collected archives may contain potentially sensitive configuration information:

- Database and table names, schemas, and object definitions
- PostgreSQL configuration settings (but not passwords)
- System configuration and resource utilization metrics
- Active connection counts and database statistics

The tool does **not** collect: passwords, query result data, table contents, or user-generated data. Review archive contents before sharing externally.

## Usage

```
Usage: radar [options]

Options:
  -U string
    	database user (default postgres)
  -d string
    	database name (default "postgres")
  -data-dir string
    	PostgreSQL data directory
  -h string
    	database host (default "localhost")
  -p int
    	database port (default 5432)
  -skip-postgres
    	skip PostgreSQL data collection
  -skip-system
    	skip system data collection
  -v	verbose output (summary)
  -vv
    	very verbose output (detailed)
```

### Environment Variables

- `PGHOST` - PostgreSQL host
- `PGPORT` - PostgreSQL port
- `PGUSER` - PostgreSQL username
- `PGPASSWORD` - PostgreSQL password

### Sample Output

```
$ ./radar -d mydatabase
✓ Archive created: radar-hostname-20260115-133700.zip (1.2 MB)
```

## Data Collected

For a complete reference of all collected data, see [docs/data.md](docs/data.md).

**System Information**

- **Block devices & storage**: `blockdev`, `df`, `fstab`, `io-schedulers`, `iostat`, `lsblk`, `ls`, `mount`, `mounts`, `nfsiostat`, `swaps`
- **CPU & performance**: `cpuinfo`, `energy_perf_bias`, `intel_pstate`, `loadavg`, `mpstat`, `sar`, `scaling_available_governors`, `scaling_driver`, `scaling_governor`, `top`, `uptime`, `vmstat`
- **Memory & kernel**: `dmesg`, `ipcs`, `meminfo`, `pressure/cpu`, `pressure/io`, `pressure/memory`, `sysctl`, `transparent_hugepage`
- **Operating system**: `apt`, `dnf`, `dpkg`, `hostname`, `hosts`, `limits.conf`, `locale`, `locale.conf`, `localectl`, `lsmod`, `lspci`, `machine-id`, `os-release`, `ps`, `rpm`, `system-release`, `systemctl`, `systemd-detect-virt`, `tuned-adm`, `uname`, `yum`
- **Network**: `ifconfig`, `ip`
- **Security**: `fips-mode-setup`, `openssl`, `sestatus`, `update-crypto-policies`

**PostgreSQL Instance**

- **Configuration & files**: `pg_db_role_setting`, `pg_hba.conf`, `pg_hba_file_rules`, `pg_ident.conf`, `pg_settings`, `pg_tablespace`, `postgresql.auto.conf`, `postgresql.conf`, `recovery.conf`, `recovery.done`
- **Activity & monitoring**: `pg_locks`, `pg_postmaster_start_time()`, `pg_prepared_xacts`, `pg_stat_activity`
- **Statistics views**: `pg_stat_archiver`, `pg_stat_bgwriter`, `pg_stat_checkpointer` (PG17+), `pg_stat_io` (PG16+), `pg_stat_slru`, `pg_stat_wal` (PG14+)
- **Replication**: `pg_replication_origin_status`, `pg_replication_slots`, `pg_stat_replication`, `pg_subscription`
- **Progress tracking**: `pg_stat_progress_analyze`, `pg_stat_progress_basebackup`, `pg_stat_progress_copy`, `pg_stat_progress_vacuum`
- **Catalog**: `pg_available_extensions`, `pg_database`, `pg_roles`, `version()`

**Per-Database**

- **Schema objects**: `pg_indexes`, `pg_namespace`, `pg_operator`, `pg_tables`, `pg_type`
- **Functions & procedures**: `pg_proc`
- **Triggers & partitioning**: `pg_inherits`, `pg_partitioned_table`, `pg_trigger`
- **Logical replication**: `pg_publication`, `pg_publication_tables`, `pg_subscription_rel`
- **Extensions**: `pg_extension`, `pg_language`, `pg_statistic_ext`
- **Statistics**: `pg_stat_database` (conflicts, deadlocks, temp files, stats reset)

**[pg_statviz](https://github.com/vyruss/pg_statviz) Extension** (if present)

- **Time-series statistics**: `pgstatviz.buf`, `pgstatviz.conf`, `pgstatviz.conn`, `pgstatviz.db`, `pgstatviz.io`, `pgstatviz.lock`, `pgstatviz.snapshots`, `pgstatviz.wait`, `pgstatviz.wal`

## Output Structure

All data is collected in a single ZIP file named `radar-{hostname}-{timestamp}.zip`:

```
radar-hostname-20260115-133700.zip
├── system/              (Linux system data)
│   ├── lsblk.out
│   ├── mount.out
│   ├── df.out
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
└── databases/           (Per-database data)
    ├── postgres/
    │   ├── extensions.tsv
    │   ├── tables.tsv
    │   └── ...
    └── mydb/
        ├── extensions.tsv
        └── ...
```

**File Formats**:
- **TSV files (.tsv)**: Tab-separated values with headers (PostgreSQL query results)
- **Command output (.out)**: Raw command output (stdout + stderr combined)
- **Config files (.conf)**: PostgreSQL configuration files (raw contents)

## Requirements

- **System**: Linux with standard utilities (lsblk, mount, df, ps, etc.)
- **PostgreSQL**: Version 12+ (some features require 13+, 14+, or 16+)
- **Go**: 1.23+ (for building from source - see [CONTRIBUTING.md](CONTRIBUTING.md))

## Performance

- Data streams directly to ZIP file without buffering in memory
- Sequential execution with minimal memory footprint
- Complete collection typically takes seconds

## Author

Created by Jimmy Angelakos.

## Support

- **Issues**: Report bugs and feature requests on [GitHub Issues](https://github.com/vyruss/radar/issues)
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines
- **Code of Conduct**: See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## License

See [LICENCE.md](LICENCE.md) for details.
