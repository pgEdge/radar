# Collecting PostgreSQL Instance Data

The PostgreSQL instance collectors gather instance-level data and store their
files in the `postgresql/` directory. radar copies the configuration files as
they are, so they can contain passwords, such as a password in
`primary_conninfo` on a standby. The following table describes the files that
these collectors produce:

| File | Source | Description |
|------|--------|-------------|
| `postgresql/archiver.tsv` | `pg_stat_archiver` | Reports WAL archiver statistics. |
| `postgresql/available_extensions.tsv` | `pg_available_extension_versions` | Lists the available extension versions with the installed flag and the trust and requires metadata (PostgreSQL 13 and later). |
| `postgresql/bgwriter.tsv` | `pg_stat_bgwriter` | Reports background writer statistics. |
| `postgresql/blocking_locks.tsv` | Complex query | Lists blocking and blocked lock pairs. |
| `postgresql/checkpointer.tsv` | `pg_stat_checkpointer` | Reports checkpointer statistics (PostgreSQL 17 and later). |
| `postgresql/configuration.tsv` | `pg_settings` | Lists the configuration parameters. |
| `postgresql/connection_summary.tsv` | `pg_stat_activity` | Counts the backends by `state` and `wait_event_type`. |
| `postgresql/control_checkpoint.tsv` | `pg_control_checkpoint()` | Reports the latest checkpoint, with its LSNs, timeline, and next and oldest XIDs and multixacts. |
| `postgresql/control_init.tsv` | `pg_control_init()` | Reports the cluster initialisation constants, including the block and WAL segment size and whether data checksums are enabled. |
| `postgresql/control_recovery.tsv` | `pg_control_recovery()` | Reports the recovery state, with the minimum recovery end point and the backup start and end LSNs. |
| `postgresql/control_system.tsv` | `pg_control_system()` | Reports the system identifier and the control and catalog version numbers. |
| `postgresql/database_conflicts.tsv` | `pg_stat_database_conflicts` | Reports recovery conflict statistics. |
| `postgresql/database_sizes.tsv` | `pg_database_size()` | Reports database disk usage. |
| `postgresql/databases.tsv` | `pg_database` | Lists the databases. |
| `postgresql/databases_blk.tsv` | `pg_stat_database` | Reports block read and write statistics. |
| `postgresql/databases_checksums.tsv` | `pg_stat_database` | Reports checksum failure counts. |
| `postgresql/databases_tup.tsv` | `pg_stat_database` | Reports tuple operation statistics. |
| `postgresql/databases_xact.tsv` | `pg_stat_database` | Reports transaction commit and rollback counts. |
| `postgresql/db_role_setting.tsv` | `pg_db_role_setting` | Lists the per-database and per-role settings. |
| `postgresql/file_settings.tsv` | `pg_file_settings` | Reports the configuration file parse results and errors. |
| `postgresql/log_directory.tsv` | `pg_ls_logdir()` | Lists the server log directory with the name, size and modification time of each file. The listing includes file names only and never log contents. The listing is readable by `pg_monitor` and needs no filesystem access, so the collector works when radar runs as an OS user other than `postgres`. The file contains only a header, with no rows, when `logging_collector` is off. |
| `postgresql/pg_hba.conf` | Data directory | Contains the host-based authentication configuration. |
| `postgresql/pg_hba_file_rules.tsv` | `pg_hba_file_rules` | Lists the parsed `pg_hba.conf` rules (PostgreSQL 10 and later). |
| `postgresql/pg_ident.conf` | Data directory | Contains the user name mapping configuration. |
| `postgresql/postgresql.auto.conf` | Data directory | Contains the auto-generated configuration. |
| `postgresql/postgresql.conf` | Data directory | Contains the main configuration file. |
| `postgresql/postmaster_start_time.tsv` | `pg_postmaster_start_time()` | Reports the time at which the server started. |
| `postgresql/prepared_xacts.tsv` | `pg_prepared_xacts` | Lists the prepared transactions. |
| `postgresql/recovery.conf` | Data directory | Contains the recovery configuration (PostgreSQL 11 and earlier). |
| `postgresql/recovery.done` | Data directory | Contains the recovery completion marker. |
| `postgresql/replication.tsv` | `pg_stat_replication` | Reports the replication status. |
| `postgresql/replication_origin.tsv` | `pg_replication_origin_status` | Reports the replication origin status. |
| `postgresql/replication_slots.tsv` | `pg_replication_slots` | Lists the replication slots. |
| `postgresql/roles.tsv` | `pg_roles` | Lists the database roles. |
| `postgresql/running_activity.tsv` | `pg_stat_activity` | Lists every backend in `pg_stat_activity` without a filter, including idle and background backends. |
| `postgresql/running_activity_maxage.tsv` | Complex query | Reports one row with the maximum query, transaction, backend and lock wait ages (`max_query_age`, `max_xact_age`, `max_backend_age` and `max_lock_wait_age`) across backends whose `state` is not `idle`. The file contains no query text. |
| `postgresql/running_locks.tsv` | `pg_locks WHERE granted` | Lists the held locks. |
| `postgresql/shmem_allocations.tsv` | `pg_shmem_allocations` | Reports the shared memory breakdown (PostgreSQL 13 and later). |
| `postgresql/stat_io.tsv` | `pg_stat_io` | Reports I/O statistics (PostgreSQL 16 and later). |
| `postgresql/stat_progress_analyze.tsv` | `pg_stat_progress_analyze` | Reports ANALYZE progress (PostgreSQL 13 and later). |
| `postgresql/stat_progress_basebackup.tsv` | `pg_stat_progress_basebackup` | Reports base backup progress (PostgreSQL 13 and later). |
| `postgresql/stat_progress_cluster.tsv` | `pg_stat_progress_cluster` | Reports CLUSTER and VACUUM FULL progress (PostgreSQL 12 and later). |
| `postgresql/stat_progress_copy.tsv` | `pg_stat_progress_copy` | Reports COPY progress (PostgreSQL 14 and later). |
| `postgresql/stat_progress_create_index.tsv` | `pg_stat_progress_create_index` | Reports CREATE INDEX progress (PostgreSQL 12 and later). |
| `postgresql/stat_progress_vacuum.tsv` | `pg_stat_progress_vacuum` | Reports VACUUM progress (PostgreSQL 9.6 and later). |
| `postgresql/stat_replication_slots.tsv` | `pg_stat_replication_slots` | Reports replication slot statistics, including spill counters (PostgreSQL 14 and later). |
| `postgresql/stat_slru.tsv` | `pg_stat_slru` | Reports simple least-recently-used (SLRU) cache statistics (PostgreSQL 13 and later). |
| `postgresql/stat_ssl.tsv` | `pg_stat_ssl` | Reports the SSL/TLS state of each backend. |
| `postgresql/stat_statements_calls.tsv` | `pg_stat_statements` | Lists the top 100 queries by call count (PostgreSQL 13 and later). |
| `postgresql/stat_statements_max_time.tsv` | `pg_stat_statements` | Lists the top 100 queries by maximum execution time (PostgreSQL 13 and later). |
| `postgresql/stat_statements_total_time.tsv` | `pg_stat_statements` | Lists the top 100 queries by total execution time (PostgreSQL 13 and later). |
| `postgresql/stat_subscription_stats.tsv` | `pg_stat_subscription_stats` | Reports the apply and sync error counts of each subscription (PostgreSQL 15 and later), and the per-kind conflict counters (PostgreSQL 18 and later). |
| `postgresql/stat_wal.tsv` | `pg_stat_wal` | Reports WAL statistics (PostgreSQL 14 and later). |
| `postgresql/subscriptions.tsv` | `pg_subscription` | Lists the logical replication subscriptions. |
| `postgresql/tablespace_sizes.tsv` | `pg_tablespace_size()` | Reports tablespace disk usage. |
| `postgresql/tablespaces.tsv` | `pg_tablespace` | Lists the tablespace definitions. |
| `postgresql/version.tsv` | `version()` | Reports the PostgreSQL version. |
| `postgresql/waits_sample.tsv` | `pg_stat_activity` | Lists the active wait events. |
| `postgresql/wal_position.tsv` | `pg_current_wal_lsn()` | Reports the current WAL write, insert and flush LSNs, and whether the server is in recovery. The last received and replayed LSNs and the last replay timestamp apply only during recovery, and are empty on a primary. |
| `postgresql/wal_receiver.tsv` | `pg_stat_wal_receiver` | Reports the WAL receiver status on a standby. |
