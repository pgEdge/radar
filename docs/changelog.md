# Changelog

All notable changes to the pgEdge radar project will be
documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.2] - 2026-09-03

### Changed
- `--version` and `-V` options - print the radar version
- `--help` writes to stdout rather than stderr
- `--help` and `--version` override every other option; `--help` wins when both are given

### Security
- Archive created readable only by its owner
- radar refuses to write to an existing output path instead of truncating it

### Fixed
- `radar.out` carried no modification time

## [0.6.1] - 2026-09-01

### Fixed
- `--help` capture in README.md and docs/index.md regenerated from the real CLI

## [0.6.0] - 2026-09-01

### Added

#### PostgreSQL
- `control_checkpoint`, `control_init`, `control_recovery`, `control_system` collectors - `pg_controldata` content through the `pg_control_*()` functions, so neither the data directory nor a matching server binary is needed
- `stat_subscription_stats` collector - per-subscription apply and sync error counts, and per-kind conflict counters (PG15+)
- `log_directory` collector - server log directory listing from `pg_ls_logdir()`: name, size and modification time per file, never contents. Readable by `pg_monitor` and needs no filesystem access

#### Per-database
- `table_freeze_age` collector - top 1000 relations by transaction-ID and multixact freeze age, ranked by age rather than size, and reaching the `pg_catalog` and `pg_toast` relations that `tables` excludes
- `relfrozenxid_age` and `relminmxid_age` columns in the `tables` collector - per-table transaction-ID and multixact freeze age, empty for partitioned tables, which hold no tuples of their own

#### pg_statviz
- `blocking` collector - blocking lock history: blocked and blocker counts plus per-lock-type counts (pg_statviz 1.2+)

#### Spock
- 13 collectors under `spock/<db>/`, skipped when Spock is not installed: `channel_summary_stats`, `exception_log`, `lag_tracker`, `local_node`, `local_sync_status`, `node`, `pii`, `progress`, `replication_set`, `replication_set_table`, `resolutions`, `subscription`, `tables`

#### PgBouncer
- `pgbouncer.ini` collector - the `[pgbouncer]` section verbatim; every other section keeps its key names and loses its values, and comments are dropped wherever they appear
- `pgbouncer_files` collector - listing of the configuration directory. `userlist.txt` is recorded as an entry only, never read
- Collected without a database connection, so pooler configuration reaches the archive even when the instance is unreachable
- `-pgbouncer-conf` flag - PgBouncer configuration directory outside `/etc/pgbouncer` and `/usr/local/etc/pgbouncer`. A path given here that does not exist is a failure rather than a silent skip

## [0.5.1] - 2026-08-06

### Security
- Go toolchain updated to 1.25.12 (1.24 is end-of-life) - fixes `crypto/tls`, `crypto/x509`, `net`, `net/url` and `os` vulnerabilities
- `github.com/jackc/pgx/v5` updated to v5.10.0 - fixes a SQL injection issue
- `golang.org/x/text` updated to v0.40.0 - fixes an infinite loop in `norm.Iter`

## [0.5.0] - 2026-05-08

### Added

#### PostgreSQL
- `stat_ssl` collector - per-backend SSL/TLS state
- `stat_replication_slots` collector - replication slot spill counters (PG14+)

#### Per-database
- `sequences` collector - sequences with last/min/max values (PG10+)
- `bloat` collector - heuristic table bloat estimate from `pg_stats`
- `pgstattuple` collector - authoritative bloat via `pgstattuple_approx()`; opt-in with `-include pgstattuple`

#### CLI
- `-exclude name1,name2` flag - skip named tasks
- `-include name1,name2` flag - enable default-disabled tasks

### Changed
- `databases` collector adds `datfrozenxid`, `datminmxid`, `datconnlimit`, `datistemplate`, `datallowconn`
- `tables` collector rebuilt on `pg_class` + `pg_stat_all_tables`; adds dead-tup counters, vacuum/analyze timestamps + ages, `reloptions`, `relpersistence`, `relpages`, heap/table/toast sizes; capped at top 1000 by size
- `indexes` collector rebuilt on `pg_index` + `pg_stat_all_indexes`; adds semantic-key columns, `indisvalid`, `idx_scan`/`idx_tup_read`/`idx_tup_fetch`, `relpages`, `index_size`; capped at top 1000 by size
- `running_activity_maxage` collector adds `max_lock_wait_age`
- `available_extensions` collector now sourced from `pg_available_extension_versions` (one row per version with trust/requires/schema metadata)
- Archive root now contains `radar.out` with the radar binary's `version` (build-time stamp) and `commit`

### Fixed
- `read_ahead` collector ignored virtio/xen block devices and produced unlabeled output
- `hostname.out` reported resolver-canonical FQDN instead of kernel hostname; added `hostname_fqdn.out` for the FQDN
- `pg_table_size`/`pg_relation_size` called for every relation pre-sort caused fstat overload on >1000-table clusters; now uses `relpages × block_size` estimate with fstat fallback for `relpages = 0` rows

## [0.4.1] - 2026-04-15

### Fixed
- Missing PostgreSQL extensions (pg_stat_statements, pg_statviz) no longer reported as errors when not installed

## [0.4.0] - 2026-04-01

### Added

#### System (Linux)
- Cgroup v2 resource limit collectors (cpu.max, memory.max, memory.current, pids, cpuset, io.max, etc.)
- Cgroup v1 resource limit collectors (cpu.shares, cpu.cfs_quota, memory.limit, etc.)
- Container identity collectors (cgroup membership, mountinfo, environment, K8s namespace)
- Cloud/hardware identity collectors (DMI bios_vendor, product_name, sys_vendor, chassis_asset_tag)
- Auto-detection of container runtime (Docker, Kubernetes, containerd, LXC)

#### Authentication
- SSL/TLS with configurable `-sslmode` (prefer, disable, require, verify-ca, verify-full)
- Certificate authentication (`-sslcert`, `-sslkey`, `-sslrootcert`)
- GSSAPI/Kerberos authentication
- `PGSSLMODE`, `PGSSLCERT`, `PGSSLKEY`, `PGSSLROOTCERT` environment variables

### Changed
- PostgreSQL driver switched from lib/pq to pgx
- Default sslmode changed from `disable` to `prefer`
- Go bumped from 1.23 to 1.24.13
- Collector errors now distinguish skip (unavailable) from real failures

### Fixed
- db handle leak on ping failure
- Summary printed before zero-data check
- PGDATABASE fallback ordering with `--skip-system`
- Connection string quoting for values with spaces

## [0.3.0] - 2026-02-26

### Added

#### System (Linux)
- `clocksource` collector - current kernel clocksource (tsc/hpet/kvm-clock)
- `diskstats` collector - raw kernel I/O counters from /proc/diskstats
- `free` collector - memory usage summary
- `io-queue-depth` collector - I/O queue depth per block device
- `lscpu` collector - structured CPU topology (NUMA, cores, threads, caches)
- `netstat-stats` collector - protocol statistics (retransmits, drops)
- `numactl` collector - NUMA node layout and memory
- `numastat` collector - per-node memory allocation statistics
- `pg-service-status` collector - PostgreSQL systemd service status
- `resolv.conf` collector - DNS resolver configuration
- `ss-listeners` collector - listening sockets and ports
- `ss-summary` collector - socket statistics summary
- `timedatectl` collector - NTP sync and timezone status

#### System (macOS)
- `memory-pressure` collector - system memory pressure level
- `netstat-stats` collector - protocol statistics
- `ulimit` collector - resource limits

#### PostgreSQL
- `connection_summary` collector - connection distribution by state and wait event
- `database_conflicts` collector - recovery conflicts from pg_stat_database_conflicts
- `database_sizes` collector - disk usage per database
- `file_settings` collector - config file parse results from pg_file_settings
- `shmem_allocations` collector - shared memory breakdown from pg_shmem_allocations
- `stat_progress_cluster` collector - CLUSTER/VACUUM FULL progress
- `stat_progress_create_index` collector - CREATE INDEX progress
- `stat_statements_calls` collector - top 100 queries by call count (pg_stat_statements)
- `stat_statements_max_time` collector - top 100 queries by max execution time (pg_stat_statements)
- `stat_statements_total_time` collector - top 100 queries by total execution time (pg_stat_statements)
- `tablespace_sizes` collector - tablespace disk usage
- `wal_position` collector - WAL position and recovery state
- `wal_receiver` collector - standby WAL receiver status

#### Testing
- Duplicate archive path detection tests for system and PostgreSQL collectors

## [0.2.0] - 2025-12-23

### Added
- `pg_stat_checkpointer` collector (PostgreSQL 17+) - checkpoint operations
- `pg_stat_io` collector (PostgreSQL 16+) - I/O statistics by backend type
- `pg_stat_wal` collector (PostgreSQL 14+) - WAL generation statistics
- `pg_postmaster_start_time()` collector - server startup timestamp
- `pg_stat_database` per-database collector - conflicts, deadlocks, temp files

## [0.1.0] - 2025-12-18

### Added
- Initial release
- Static binaries for Linux (amd64, arm64) and macOS (amd64, arm64)
- Comprehensive system and PostgreSQL metrics collection
