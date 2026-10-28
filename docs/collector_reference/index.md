# radar Data Collection Reference

This reference describes every file that radar collects. radar writes all
collected data to a single ZIP file named `radar-{hostname}-{timestamp}.zip`.
The archive root contains a `radar.out` entry that identifies the radar binary
that produced the archive. The `version` value in that entry comes from the
build-time `-X main.version` stamp. The `commit` value comes from the version
control information that Go embeds in the binary.

radar collects metadata and statistics from the operating system, the
PostgreSQL instance, each database, and optional extensions and tools.

## Collecting System Information

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

## Collecting PostgreSQL Instance Data

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

## Collecting Per-Database Data

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

## Collecting pg_statviz Data

When a database has the [pg_statviz](https://github.com/vyruss/pg_statviz)
extension installed, radar collects its time-series statistics from
`pgstatviz.blocking`, `pgstatviz.buf`, `pgstatviz.conf`, `pgstatviz.conn`,
`pgstatviz.db`, `pgstatviz.io`, `pgstatviz.lock`, `pgstatviz.repl`,
`pgstatviz.slru`, `pgstatviz.snapshots`, `pgstatviz.wait`, and `pgstatviz.wal`.

## Collecting PgBouncer Configuration

When [PgBouncer](https://www.pgbouncer.org/) is installed, radar collects the
`[pgbouncer]` section of `pgbouncer.ini` and a listing of the configuration
directory. radar replaces the values outside the `[pgbouncer]` section with
`<redacted by radar>`, because `[databases]` connection strings routinely
contain `password=`, and drops all comments. radar records `userlist.txt` as a
directory entry only; the file holds credentials, and radar never collects its
contents. radar reads the first of `/etc/pgbouncer` and
`/usr/local/etc/pgbouncer` that exists; use `--pgbouncer-conf` to name another
configuration directory.

## Collecting Spock Replication State

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

## Protecting Sensitive Data

radar collects metadata and statistics, and does not collect table contents or
query result sets. A collected archive includes:

- database and table names, schemas, and object definitions.
- PostgreSQL configuration files and settings.
- system configuration and resource utilization metrics.
- active connection counts, database statistics, and query text.

!!! warning

    A collected archive can contain passwords and other secrets. PostgreSQL
    configuration files and settings can hold connection passwords, such as a
    password in `primary_conninfo` on a standby, and query text and error
    messages can include values from user queries. Review the contents of an
    archive before you share it outside your organization.

## Browsing the Collector Reference

The following table links to the reference page for each collector:

| Page | Description |
|------|-------------|
| [Collecting Cross-Platform System Data](cross_platform_system.md) | Cross-platform system collectors work on both Linux and macOS. |
| [Collecting Linux System Data](linux_system.md) | Linux-specific system collectors run only on Linux and use sources such as systemd, `/proc` and `/sys`. |
| [Collecting macOS System Data](macos_system.md) | macOS-specific system collectors run only on macOS and use tools such as `sysctl`, `system_profiler` and `diskutil`. |
| [Collecting PostgreSQL Instance Data](postgresql_instance.md) | PostgreSQL instance collectors gather instance-level PostgreSQL data. |
| [Collecting Per-Database Data](per_database.md) | Per-database collectors gather schema and object data for each database. |
| [Collecting pg_statviz Data](pg_statviz.md) | The optional pg_statviz collectors gather time-series statistics for each database. |
| [Collecting Spock Data](spock.md) | The optional Spock collectors gather multi-master replication state for each database. |
| [Collecting PgBouncer Configuration](pgbouncer.md) | The optional PgBouncer collectors gather the pooler configuration when PgBouncer is installed. |
