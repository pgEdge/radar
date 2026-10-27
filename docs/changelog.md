# Changelog

All notable changes to the pgEdge radar project will be documented in this
file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and
this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.3] - 2026-10-27

### Added

- The documentation describes installing radar from the pgEdge Enterprise
  Postgres repository and from the pre-built binaries on GitHub, with checksum
  verification.
- The documentation describes the `--exclude` and `--include` options, and how
  to find collector names with `-vv`.

### Changed

- The build instructions clone the pgEdge radar repository and require Go
  1.25.12 or later.
- The published data reference covers the control file, log directory,
  subscription statistics, freeze age, Spock, and PgBouncer collectors.
- The documentation names the minimum PostgreSQL version of each collector that
  needs one, and describes the archive layout and the summary line in detail.

## [0.6.2] - 2026-09-03

### Changed

- The `--version` and `-V` options print the radar version.
- The `--help` option writes to `stdout` rather than `stderr`.
- The `--help` and `--version` options override every other option; `--help`
  wins when both are given.

### Security

- radar creates the archive so that only its owner can read it.
- radar refuses to write to an existing output path instead of truncating it.

### Fixed

- The `radar.out` archive entry had no modification time.

## [0.6.1] - 2026-09-01

### Fixed

- This release regenerates the `--help` capture in README.md and docs/index.md
  from the real CLI.

## [0.6.0] - 2026-09-01

### Added

#### PostgreSQL

- The `control_checkpoint`, `control_init`, `control_recovery` and
  `control_system` collectors record the `pg_controldata` content through the
  `pg_control_*()` functions, so neither the data directory nor a matching
  server binary is needed.
- The `stat_subscription_stats` collector records per-subscription apply and
  sync error counts (PG15+), and per-kind conflict counters (PG18+).
- The `log_directory` collector lists the server log directory from
  `pg_ls_logdir()`, with the name, size and modification time of each file but
  never its contents. The listing is readable by `pg_monitor` and needs no
  filesystem access.

#### Per-database

- The `table_freeze_age` collector records the top 1000 relations by
  transaction-ID and multixact freeze age. The collector ranks relations by age
  rather than size, and includes the `pg_catalog` and `pg_toast` relations that
  `tables` excludes.
- The `tables` collector adds the `relfrozenxid_age` and `relminmxid_age`
  columns, which record the transaction-ID and multixact freeze age of each
  table. Both columns are empty for partitioned tables, which hold no tuples of
  their own.

#### pg_statviz

- The `blocking` collector records the blocking lock history, with blocked and
  blocker counts plus per-lock-type counts (pg_statviz 1.2+).

#### Spock

- This release adds 13 collectors under `spock/<db>/`: `channel_summary_stats`,
  `exception_log`, `lag_tracker`, `local_node`, `local_sync_status`, `node`,
  `pii`, `progress`, `replication_set`, `replication_set_table`, `resolutions`,
  `subscription` and `tables`. radar skips these collectors when Spock is not
  installed.

#### PgBouncer

- The `pgbouncer.ini` collector records the `[pgbouncer]` section without its
  comments. Every other section keeps its key names and loses its values, and
  the collector drops comments wherever they appear.
- The `pgbouncer_files` collector lists the configuration directory. The
  collector records `userlist.txt` as an entry only and never reads the file.
- The PgBouncer collectors run without a database connection, so the pooler
  configuration reaches the archive even when the instance is unreachable.
- The `--pgbouncer-conf` flag specifies a PgBouncer configuration directory
  outside `/etc/pgbouncer` and `/usr/local/etc/pgbouncer`. A nonexistent path
  given to this flag is a failure rather than a silent skip.

## [0.5.1] - 2026-08-06

### Security

- This release updates the Go toolchain to 1.25.12 (1.24 is end-of-life); the
  update fixes vulnerabilities in `crypto/tls`, `crypto/x509`, `net`, `net/url`
  and `os`.
- This release updates `github.com/jackc/pgx/v5` to v5.10.0, which fixes a SQL
  injection issue.
- This release updates `golang.org/x/text` to v0.40.0, which fixes an infinite
  loop in `norm.Iter`.

## [0.5.0] - 2026-05-08

### Added

#### PostgreSQL

- The `stat_ssl` collector records the SSL/TLS state of each backend.
- The `stat_replication_slots` collector records replication slot spill
  counters (PG14+).

#### Per-database

- The `sequences` collector records sequences with their last, minimum and
  maximum values (PG10+).
- The `bloat` collector records a heuristic table bloat estimate from
  `pg_stats`.
- The `pgstattuple` collector measures bloat through `pgstattuple_approx()`;
  users opt in with `--include pgstattuple`.

#### CLI

- The `--exclude name1,name2` flag skips the named tasks.
- The `--include name1,name2` flag enables named tasks that are disabled by
  default.

### Changed

- The `databases` collector adds the `datfrozenxid`, `datminmxid`,
  `datconnlimit`, `datistemplate` and `datallowconn` columns.
- This release rebuilds the `tables` collector on `pg_class` and
  `pg_stat_all_tables`. The collector adds dead-tuple counters, vacuum and
  analyze timestamps and ages, `reloptions`, `relpersistence`, `relpages`, and
  heap, table and toast sizes. The collector caps its output at the top 1000
  tables by size.
- This release rebuilds the `indexes` collector on `pg_index` and
  `pg_stat_all_indexes`. The collector adds semantic-key columns, `indisvalid`,
  `idx_scan`, `idx_tup_read`, `idx_tup_fetch`, `relpages` and `index_size`, and
  caps its output at the top 1000 indexes by size.
- The `running_activity_maxage` collector adds `max_lock_wait_age`.
- The `available_extensions` collector now reads from
  `pg_available_extension_versions` (one row per version with
  trust/requires/schema metadata).
- The archive root now contains `radar.out`, which records the radar binary's
  `version` (build-time stamp) and `commit`.

### Fixed

- The `read_ahead` collector ignored virtio/xen block devices and produced
  unlabeled output.
- The `hostname.out` entry reported the resolver-canonical fully qualified
  domain name (FQDN) instead of the kernel hostname. This release adds
  `hostname_fqdn.out` to record that name.
- Calling `pg_table_size` and `pg_relation_size` for every relation before
  sorting caused fstat overload on clusters with more than 1000 tables. radar
  now uses a `relpages × block_size` estimate, with an fstat fallback for rows
  where `relpages = 0`.

## [0.4.1] - 2026-04-15

### Fixed

- radar no longer reports missing PostgreSQL extensions (pg_stat_statements,
  pg_statviz) as errors when they are not installed.

## [0.4.0] - 2026-04-01

### Added

#### System (Linux)

- This release adds cgroup v2 resource limit collectors (cpu.max, memory.max,
  memory.current, pids, cpuset, io.max, etc.).
- This release adds cgroup v1 resource limit collectors (cpu.shares,
  cpu.cfs_quota, memory.limit, etc.).
- This release adds container identity collectors (cgroup membership,
  mountinfo, environment, Kubernetes namespace).
- This release adds cloud and hardware identity collectors (DMI bios_vendor,
  product_name, sys_vendor, chassis_asset_tag).
- radar automatically detects the container runtime (Docker, Kubernetes,
  containerd, LXC).

#### Authentication

- radar supports SSL/TLS with a configurable `--sslmode` (prefer, disable,
  require, verify-ca, verify-full).
- radar supports certificate authentication (`--sslcert`, `--sslkey`,
  `--sslrootcert`).
- radar supports GSSAPI/Kerberos authentication.
- radar reads the `PGSSLMODE`, `PGSSLCERT`, `PGSSLKEY` and `PGSSLROOTCERT`
  environment variables.

### Changed

- This release switches the PostgreSQL driver from lib/pq to pgx.
- The default sslmode changes from `disable` to `prefer`.
- This release updates Go from 1.23 to 1.24.13.
- Collector errors now distinguish skips (unavailable) from real failures.

### Fixed

- radar leaked a database handle when a ping failed.
- radar printed the summary before the zero-data check.
- radar applied the PGDATABASE fallback after the `--skip-system` check rather
  than before it.
- radar did not quote connection string values that contain spaces.

## [0.3.0] - 2026-02-26

### Added

#### System (Linux)

- The `clocksource` collector records the current kernel clocksource
  (tsc/hpet/kvm-clock).
- The `diskstats` collector records the raw kernel I/O counters from
  /proc/diskstats.
- The `free` collector records a memory usage summary.
- The `io-queue-depth` collector records the I/O queue depth of each block
  device.
- The `lscpu` collector records the structured CPU topology (NUMA, cores,
  threads, caches).
- The `netstat-stats` collector records protocol statistics (retransmits,
  drops).
- The `numactl` collector records the NUMA node layout and memory.
- The `numastat` collector records per-node memory allocation statistics.
- The `pg-service-status` collector records the PostgreSQL systemd service
  status.
- The `resolv.conf` collector records the DNS resolver configuration.
- The `ss-listeners` collector records listening sockets and ports.
- The `ss-summary` collector records a socket statistics summary.
- The `timedatectl` collector records the NTP sync and timezone status.

#### System (macOS)

- The `memory-pressure` collector records the system memory pressure level.
- The `netstat-stats` collector records protocol statistics.
- The `ulimit` collector records resource limits.

#### PostgreSQL

- The `connection_summary` collector records the connection distribution by
  state and wait event.
- The `database_conflicts` collector records recovery conflicts from
  pg_stat_database_conflicts.
- The `database_sizes` collector records the disk usage of each database.
- The `file_settings` collector records configuration file parse results from
  pg_file_settings.
- The `shmem_allocations` collector records the shared memory breakdown from
  pg_shmem_allocations.
- The `stat_progress_cluster` collector records CLUSTER/VACUUM FULL progress.
- The `stat_progress_create_index` collector records CREATE INDEX progress.
- The `stat_statements_calls` collector records the top 100 queries by call
  count (pg_stat_statements).
- The `stat_statements_max_time` collector records the top 100 queries by
  maximum execution time (pg_stat_statements).
- The `stat_statements_total_time` collector records the top 100 queries by
  total execution time (pg_stat_statements).
- The `tablespace_sizes` collector records tablespace disk usage.
- The `wal_position` collector records the WAL position and recovery state.
- The `wal_receiver` collector records the standby WAL receiver status.

#### Testing

- This release adds tests that detect duplicate archive paths for the system
  and PostgreSQL collectors.

## [0.2.0] - 2025-12-23

### Added

- The `pg_stat_checkpointer` collector (PostgreSQL 17+) records checkpoint
  operations.
- The `pg_stat_io` collector (PostgreSQL 16+) records I/O statistics by backend
  type.
- The `pg_stat_wal` collector (PostgreSQL 14+) records WAL generation
  statistics.
- The `pg_postmaster_start_time()` collector records the server startup
  timestamp.
- The `pg_stat_database` per-database collector records conflicts, deadlocks
  and temporary files.

## [0.1.0] - 2025-12-18

### Added

- This is the initial release.
- The release includes static binaries for Linux (amd64, arm64) and macOS
  (amd64, arm64).
- radar collects comprehensive system and PostgreSQL metrics.
