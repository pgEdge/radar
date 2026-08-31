/*-------------------------------------------------------------------------
 *
 * radar
 *
 * Portions copyright (c) 2026, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package main

import (
	"database/sql"
	"io"
)

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

// PostgreSQL instance-level query tasks (sorted alphabetically by name)
var postgresQueryTasks = []SimpleQueryTask{
	{
		Name:        "activity",
		ArchivePath: "postgresql/running_activity.tsv",
		Query:       "SELECT * FROM pg_stat_activity ORDER BY pid",
	},
	{
		Name:        "archiver",
		ArchivePath: "postgresql/archiver.tsv",
		Query:       "SELECT * FROM pg_stat_archiver",
	},
	{
		Name:        "available_extensions",
		ArchivePath: "postgresql/available_extensions.tsv",
		Query: `SELECT name, version, installed, superuser, trusted,
       relocatable, schema, requires, comment
FROM pg_available_extension_versions
ORDER BY name, version`,
	},
	{
		Name:        "bgwriter",
		ArchivePath: "postgresql/bgwriter.tsv",
		Query:       "SELECT * FROM pg_stat_bgwriter",
	},
	{
		Name:        "blocking_locks",
		ArchivePath: "postgresql/blocking_locks.tsv",
		Query: `SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted`,
	},
	{
		Name:        "checkpointer",
		ArchivePath: "postgresql/checkpointer.tsv",
		Query:       "SELECT * FROM pg_stat_checkpointer",
	},
	{
		Name:        "configuration",
		ArchivePath: "postgresql/configuration.tsv",
		Query:       "SELECT name, setting, unit, category, short_desc FROM pg_settings ORDER BY category, name",
	},
	{
		Name:        "connection_summary",
		ArchivePath: "postgresql/connection_summary.tsv",
		Query:       "SELECT state, wait_event_type, count(*) FROM pg_stat_activity GROUP BY state, wait_event_type ORDER BY count(*) DESC",
	},
	{
		// pg_controldata's content, read through the catalogue functions rather
		// than by shelling out to the binary, which needs the data directory
		// and a matching binary version on PATH. Readable by pg_monitor.
		Name:        "control_checkpoint",
		ArchivePath: "postgresql/control_checkpoint.tsv",
		Query:       "SELECT * FROM pg_control_checkpoint()",
	},
	{
		// Carries data_page_checksum_version, which is the only place in the
		// archive saying whether checksums are enabled, because
		// databases_checksums.tsv reports failure counts rather than the setting.
		Name:        "control_init",
		ArchivePath: "postgresql/control_init.tsv",
		Query:       "SELECT * FROM pg_control_init()",
	},
	{
		Name:        "control_recovery",
		ArchivePath: "postgresql/control_recovery.tsv",
		Query:       "SELECT * FROM pg_control_recovery()",
	},
	{
		Name:        "control_system",
		ArchivePath: "postgresql/control_system.tsv",
		Query:       "SELECT * FROM pg_control_system()",
	},
	{
		Name:        "database_conflicts",
		ArchivePath: "postgresql/database_conflicts.tsv",
		Query:       "SELECT * FROM pg_stat_database_conflicts ORDER BY datname",
	},
	{
		Name:        "database_sizes",
		ArchivePath: "postgresql/database_sizes.tsv",
		Query:       "SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size FROM pg_database WHERE datallowconn ORDER BY pg_database_size(datname) DESC",
	},
	{
		Name:        "databases",
		ArchivePath: "postgresql/databases.tsv",
		Query: `SELECT oid, datname, datdba, encoding, datcollate, datctype,
       datistemplate, datallowconn, datconnlimit,
       datfrozenxid, age(datfrozenxid) AS frozenxid_age,
       datminmxid, mxid_age(datminmxid) AS minmxid_age
FROM pg_database
ORDER BY datname`,
	},
	{
		Name:        "databases_blk",
		ArchivePath: "postgresql/databases_blk.tsv",
		Query:       "SELECT datname, blks_read, blks_hit, blk_read_time, blk_write_time FROM pg_stat_database WHERE datname IS NOT NULL ORDER BY datname",
	},
	{
		Name:        "databases_checksums",
		ArchivePath: "postgresql/databases_checksums.tsv",
		Query:       "SELECT datname, checksum_failures, checksum_last_failure FROM pg_stat_database WHERE datname IS NOT NULL ORDER BY datname",
	},
	{
		Name:        "databases_tup",
		ArchivePath: "postgresql/databases_tup.tsv",
		Query:       "SELECT datname, tup_returned, tup_fetched, tup_inserted, tup_updated, tup_deleted FROM pg_stat_database WHERE datname IS NOT NULL ORDER BY datname",
	},
	{
		Name:        "databases_xact",
		ArchivePath: "postgresql/databases_xact.tsv",
		Query:       "SELECT datname, xact_commit, xact_rollback FROM pg_stat_database WHERE datname IS NOT NULL ORDER BY datname",
	},
	{
		Name:        "db_role_setting",
		ArchivePath: "postgresql/db_role_setting.tsv",
		Query:       "SELECT setdatabase, setrole, setconfig FROM pg_db_role_setting",
	},
	{
		Name:        "file_settings",
		ArchivePath: "postgresql/file_settings.tsv",
		Query:       "SELECT * FROM pg_file_settings ORDER BY sourcefile, seqno",
	},
	{
		// pg_ls_logdir() rather than walking the directory: it is executable by
		// pg_monitor and needs no filesystem access, so the listing works when
		// radar runs as an OS user with no rights on the data directory. Names,
		// sizes and timestamps only, never log contents.
		//
		// pg_ls_logdir() raises 58P01 when the log directory is absent, which is
		// the case whenever the logging collector is off. current_setting is
		// STABLE so the planner considers it a pseudo-constant. It then works as
		// a gate on a Result above the Function Scan, so pg_ls_logdir() is never
		// executed.
		Name:        "log_directory",
		ArchivePath: "postgresql/log_directory.tsv",
		Query: `SELECT l.name, l.size, l.modification
FROM (SELECT 1 WHERE current_setting('logging_collector')::bool) g,
     LATERAL pg_ls_logdir() l
ORDER BY l.modification DESC, l.name`,
	},
	{
		Name:        "pg_hba_file_rules",
		ArchivePath: "postgresql/pg_hba_file_rules.tsv",
		Query:       "SELECT * FROM pg_hba_file_rules ORDER BY line_number",
	},
	{
		Name:        "postmaster_start_time",
		ArchivePath: "postgresql/postmaster_start_time.tsv",
		Query:       "SELECT pg_postmaster_start_time() AS start_time",
	},
	{
		Name:        "prepared_xacts",
		ArchivePath: "postgresql/prepared_xacts.tsv",
		Query:       "SELECT * FROM pg_prepared_xacts ORDER BY prepared",
	},
	{
		Name:        "replication",
		ArchivePath: "postgresql/replication.tsv",
		Query:       "SELECT * FROM pg_stat_replication",
	},
	{
		Name:        "replication_origin",
		ArchivePath: "postgresql/replication_origin.tsv",
		Query:       "SELECT * FROM pg_replication_origin_status",
	},
	{
		Name:        "replication_slots",
		ArchivePath: "postgresql/replication_slots.tsv",
		Query:       "SELECT * FROM pg_replication_slots ORDER BY slot_name",
	},
	{
		Name:        "roles",
		ArchivePath: "postgresql/roles.tsv",
		Query:       "SELECT * FROM pg_roles ORDER BY rolname",
	},
	{
		Name:        "running_activity_maxage",
		ArchivePath: "postgresql/running_activity_maxage.tsv",
		Query: `SELECT
    max(clock_timestamp() - query_start) AS max_query_age,
    max(clock_timestamp() - xact_start) AS max_xact_age,
    max(clock_timestamp() - backend_start) AS max_backend_age,
    max(clock_timestamp() - state_change)
        FILTER (WHERE wait_event_type = 'Lock')
        AS max_lock_wait_age
FROM pg_stat_activity
WHERE state != 'idle'`,
	},
	{
		Name:        "running_locks",
		ArchivePath: "postgresql/running_locks.tsv",
		Query:       "SELECT * FROM pg_locks WHERE granted ORDER BY pid, locktype",
	},
	{
		Name:        "shmem_allocations",
		ArchivePath: "postgresql/shmem_allocations.tsv",
		Query:       "SELECT * FROM pg_shmem_allocations ORDER BY size DESC",
	},
	{
		Name:        "stat_io",
		ArchivePath: "postgresql/stat_io.tsv",
		Query:       "SELECT * FROM pg_stat_io ORDER BY backend_type, context, object",
	},
	{
		Name:        "stat_progress_analyze",
		ArchivePath: "postgresql/stat_progress_analyze.tsv",
		Query:       "SELECT * FROM pg_stat_progress_analyze",
	},
	{
		Name:        "stat_progress_basebackup",
		ArchivePath: "postgresql/stat_progress_basebackup.tsv",
		Query:       "SELECT * FROM pg_stat_progress_basebackup",
	},
	{
		Name:        "stat_progress_cluster",
		ArchivePath: "postgresql/stat_progress_cluster.tsv",
		Query:       "SELECT * FROM pg_stat_progress_cluster",
	},
	{
		Name:        "stat_progress_copy",
		ArchivePath: "postgresql/stat_progress_copy.tsv",
		Query:       "SELECT * FROM pg_stat_progress_copy",
	},
	{
		Name:        "stat_progress_create_index",
		ArchivePath: "postgresql/stat_progress_create_index.tsv",
		Query:       "SELECT * FROM pg_stat_progress_create_index",
	},
	{
		Name:        "stat_progress_vacuum",
		ArchivePath: "postgresql/stat_progress_vacuum.tsv",
		Query:       "SELECT * FROM pg_stat_progress_vacuum",
	},
	{
		Name:        "stat_replication_slots",
		ArchivePath: "postgresql/stat_replication_slots.tsv",
		Query:       "SELECT * FROM pg_stat_replication_slots ORDER BY slot_name",
	},
	{
		Name:        "stat_slru",
		ArchivePath: "postgresql/stat_slru.tsv",
		Query:       "SELECT * FROM pg_stat_slru ORDER BY name",
	},
	{
		Name:        "stat_ssl",
		ArchivePath: "postgresql/stat_ssl.tsv",
		Query: `SELECT s.pid, s.ssl, s.version, s.cipher, s.bits,
       s.client_dn, s.client_serial, s.issuer_dn,
       a.usename, a.application_name, a.client_addr
FROM pg_stat_ssl s
LEFT JOIN pg_stat_activity a ON a.pid = s.pid
ORDER BY s.pid`,
	},
	{
		Name:        "stat_statements_calls",
		ArchivePath: "postgresql/stat_statements_calls.tsv",
		Query:       "SELECT userid, dbid, query, calls, total_exec_time, mean_exec_time, max_exec_time, rows FROM pg_stat_statements ORDER BY calls DESC LIMIT 100",
	},
	{
		Name:        "stat_statements_max_time",
		ArchivePath: "postgresql/stat_statements_max_time.tsv",
		Query:       "SELECT userid, dbid, query, calls, total_exec_time, mean_exec_time, max_exec_time, rows FROM pg_stat_statements ORDER BY max_exec_time DESC LIMIT 100",
	},
	{
		Name:        "stat_statements_total_time",
		ArchivePath: "postgresql/stat_statements_total_time.tsv",
		Query:       "SELECT userid, dbid, query, calls, total_exec_time, mean_exec_time, max_exec_time, rows FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 100",
	},
	{
		// SELECT * rather than a column list, because the confl_* counters
		// differ between major versions and a fixed list would break on any
		// version lacking one of them.
		Name:        "stat_subscription_stats",
		ArchivePath: "postgresql/stat_subscription_stats.tsv",
		Query:       "SELECT * FROM pg_stat_subscription_stats ORDER BY subname",
	},
	{
		Name:        "stat_wal",
		ArchivePath: "postgresql/stat_wal.tsv",
		Query:       "SELECT * FROM pg_stat_wal",
	},
	{
		Name:        "subscriptions",
		ArchivePath: "postgresql/subscriptions.tsv",
		Query:       "SELECT * FROM pg_subscription ORDER BY subname",
	},
	{
		Name:        "tablespace_sizes",
		ArchivePath: "postgresql/tablespace_sizes.tsv",
		Query:       "SELECT spcname, pg_size_pretty(pg_tablespace_size(oid)) AS size FROM pg_tablespace ORDER BY pg_tablespace_size(oid) DESC",
	},
	{
		Name:        "tablespaces",
		ArchivePath: "postgresql/tablespaces.tsv",
		Query:       "SELECT oid, spcname, spcowner, spcacl, spcoptions, pg_tablespace_location(oid) as spclocation FROM pg_tablespace ORDER BY spcname",
	},
	{
		Name:        "version",
		ArchivePath: "postgresql/version.tsv",
		Query:       "SELECT version()",
	},
	{
		Name:        "waits_sample",
		ArchivePath: "postgresql/waits_sample.tsv",
		Query:       "SELECT pid, wait_event_type, wait_event, state, query FROM pg_stat_activity WHERE wait_event IS NOT NULL ORDER BY pid",
	},
	{
		Name:        "wal_position",
		ArchivePath: "postgresql/wal_position.tsv",
		Query: `SELECT pg_current_wal_lsn() AS current_wal_lsn,
       pg_current_wal_insert_lsn() AS current_wal_insert_lsn,
       pg_current_wal_flush_lsn() AS current_wal_flush_lsn,
       pg_is_in_recovery() AS is_in_recovery,
       CASE WHEN pg_is_in_recovery() THEN pg_last_wal_receive_lsn() END AS last_wal_receive_lsn,
       CASE WHEN pg_is_in_recovery() THEN pg_last_wal_replay_lsn() END AS last_wal_replay_lsn,
       CASE WHEN pg_is_in_recovery() THEN pg_last_xact_replay_timestamp() END AS last_xact_replay_timestamp`,
	},
	{
		Name:        "wal_receiver",
		ArchivePath: "postgresql/wal_receiver.tsv",
		Query:       "SELECT * FROM pg_stat_wal_receiver",
	},
}

// Per-database query tasks (sorted alphabetically by name)
// These are per-database tasks - ArchivePath will be formatted with dbname
var perDatabaseQueryTasks = []SimpleQueryTask{
	{
		Name:        "bloat",
		ArchivePath: "databases/%s/bloat.tsv",
		Query: `
SELECT current_database() AS current_database,
       schemaname,
       tablename,
       ROUND((CASE WHEN otta = 0 THEN 0.0
                   ELSE sml.relpages::FLOAT / otta END)::NUMERIC, 1)
         AS table_bloat_ratio,
       CASE WHEN relpages < otta THEN 0
            ELSE bs * (sml.relpages - otta)::BIGINT END AS wastedbytes,
       iname,
       ituples,
       ipages,
       iotta
FROM (
  SELECT schemaname, tablename, cc.reltuples, cc.relpages, bs,
         CEIL((cc.reltuples *
               ((datahdr + ma -
                 (CASE WHEN datahdr % ma = 0 THEN ma
                       ELSE datahdr % ma END)) + nullhdr2 + 4))
              / (bs - 20::FLOAT)) AS otta,
         COALESCE(c2.relname, '?') AS iname,
         COALESCE(c2.reltuples, 0) AS ituples,
         COALESCE(c2.relpages, 0) AS ipages,
         COALESCE(CEIL((c2.reltuples * (datahdr - 12))
                       / (bs - 20::FLOAT)), 0) AS iotta
  FROM (
    SELECT ma, bs, schemaname, tablename,
           (datawidth +
            (hdr + ma -
             (CASE WHEN hdr % ma = 0 THEN ma
                   ELSE hdr % ma END)))::NUMERIC AS datahdr,
           (maxfracsum *
            (nullhdr + ma -
             (CASE WHEN nullhdr % ma = 0 THEN ma
                   ELSE nullhdr % ma END))) AS nullhdr2
    FROM (
      SELECT schemaname, tablename, hdr, ma, bs,
             SUM((1 - null_frac) * avg_width) AS datawidth,
             MAX(null_frac) AS maxfracsum,
             hdr + (SELECT 1 + COUNT(*) / 8
                    FROM pg_stats s2
                    WHERE null_frac <> 0
                      AND s2.schemaname = s.schemaname
                      AND s2.tablename = s.tablename) AS nullhdr
      FROM pg_stats s,
           (SELECT (SELECT current_setting('block_size')::NUMERIC)
                       AS bs,
                   CASE WHEN SUBSTRING(v, 12, 3)
                              IN ('8.0', '8.1', '8.2') THEN 27
                        ELSE 23 END AS hdr,
                   CASE WHEN v ~ 'mingw32' THEN 8
                        ELSE 4 END AS ma
            FROM (SELECT version() AS v) AS foo) AS constants
      GROUP BY 1, 2, 3, 4, 5) AS foo) AS rs
  JOIN pg_class cc ON cc.relname = rs.tablename
  JOIN pg_namespace nn ON cc.relnamespace = nn.oid
                      AND nn.nspname = rs.schemaname
  LEFT JOIN pg_index i ON indrelid = cc.oid
  LEFT JOIN pg_class c2 ON c2.oid = i.indexrelid) AS sml
ORDER BY wastedbytes DESC, schemaname, tablename
		`,
	},
	{
		Name:        "extensions",
		ArchivePath: "databases/%s/extensions.tsv",
		Query:       "SELECT * FROM pg_extension ORDER BY extname",
	},
	{
		Name:        "funcs",
		ArchivePath: "databases/%s/funcs.tsv",
		Query:       "SELECT oid, proname, pronamespace, proowner, prolang, prokind FROM pg_proc WHERE prokind = 'f' ORDER BY proname",
	},
	{
		Name:        "indexes",
		ArchivePath: "databases/%s/indexes.tsv",
		Query: `
			WITH params AS MATERIALIZED (
			    SELECT (count(*) > 1000) AS many_indexes
			    FROM pg_index i
			    JOIN pg_class c ON c.oid = i.indexrelid
			    JOIN pg_namespace n ON n.oid = c.relnamespace
			    WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
			      AND n.nspname NOT LIKE 'pg_toast%'
			)
			SELECT n.nspname AS schemaname,
			       t.relname AS tablename,
			       c.relname AS indexname,
			       pg_get_indexdef(i.indexrelid) AS indexdef,
			       i.indrelid::regclass AS indrelid,
			       i.indexrelid::regclass AS indexrelid,
			       i.indisunique,
			       i.indisprimary,
			       i.indisvalid,
			       i.indclass::text AS indclass,
			       i.indkey::text AS indkey,
			       pg_get_expr(i.indexprs, i.indrelid) AS indexprs,
			       pg_get_expr(i.indpred, i.indrelid) AS indpred,
			       c.relpages,
			       CASE WHEN p.many_indexes AND c.relpages > 0
			            THEN c.relpages::bigint * current_setting('block_size')::bigint
			            ELSE pg_relation_size(i.indexrelid)
			       END AS index_size,
			       s.idx_scan,
			       s.idx_tup_read,
			       s.idx_tup_fetch
			FROM pg_index i
			CROSS JOIN params p
			JOIN pg_class c ON c.oid = i.indexrelid
			JOIN pg_class t ON t.oid = i.indrelid
			JOIN pg_namespace n ON n.oid = c.relnamespace
			LEFT JOIN pg_stat_all_indexes s ON s.indexrelid = i.indexrelid
			WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
			  AND n.nspname NOT LIKE 'pg_toast%'
			ORDER BY CASE WHEN p.many_indexes AND c.relpages > 0
			              THEN c.relpages::bigint * current_setting('block_size')::bigint
			              ELSE pg_relation_size(i.indexrelid)
			         END DESC NULLS LAST,
			         schemaname, tablename, indexname
			LIMIT 1000
		`,
	},
	{
		Name:        "languages",
		ArchivePath: "databases/%s/languages.tsv",
		Query:       "SELECT * FROM pg_language ORDER BY lanname",
	},
	{
		Name:        "operators",
		ArchivePath: "databases/%s/operators.tsv",
		Query:       "SELECT oid, oprname, oprkind, oprcanmerge, oprcanhash FROM pg_operator ORDER BY oprname",
	},
	{
		Name:        "partitioned_tables",
		ArchivePath: "databases/%s/partitioned_tables.tsv",
		Query:       "SELECT * FROM pg_partitioned_table ORDER BY partrelid",
	},
	{
		Name:        "partitions",
		ArchivePath: "databases/%s/partitions.tsv",
		Query: `
			SELECT inhrelid::regclass AS partition,
			       inhparent::regclass AS parent,
			       inhseqno
			FROM pg_inherits
			ORDER BY inhparent, inhseqno
		`,
	},
	{
		Name:        "pgstattuple",
		ArchivePath: "databases/%s/pgstattuple.tsv",
		Query: `SELECT n.nspname AS schemaname,
       c.relname AS tablename,
       p.table_len,
       p.approx_tuple_count AS tuple_count,
       p.approx_tuple_len AS tuple_len,
       p.approx_tuple_percent AS tuple_percent,
       p.dead_tuple_count,
       p.dead_tuple_len,
       p.dead_tuple_percent,
       p.approx_free_space AS free_space,
       p.approx_free_percent AS free_percent
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
CROSS JOIN LATERAL pgstattuple_approx(c.oid) p
WHERE c.relkind IN ('r', 'm')
  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
  AND n.nspname NOT LIKE 'pg_toast%'
ORDER BY p.dead_tuple_percent DESC NULLS LAST`,
	},
	{
		Name:        "procs",
		ArchivePath: "databases/%s/procs.tsv",
		Query:       "SELECT oid, proname, pronamespace, proowner, prolang, prokind FROM pg_proc WHERE prokind = 'p' ORDER BY proname",
	},
	{
		Name:        "publication_tables",
		ArchivePath: "databases/%s/publication_tables.tsv",
		Query:       "SELECT * FROM pg_publication_tables ORDER BY pubname, schemaname, tablename",
	},
	{
		Name:        "publications",
		ArchivePath: "databases/%s/publications.tsv",
		Query:       "SELECT * FROM pg_publication ORDER BY pubname",
	},
	{
		Name:        "schemas",
		ArchivePath: "databases/%s/schemas.tsv",
		Query:       "SELECT * FROM pg_namespace ORDER BY nspname",
	},
	{
		Name:        "sequences",
		ArchivePath: "databases/%s/sequences.tsv",
		Query: `SELECT schemaname, sequencename,
       data_type::regtype::text AS data_type,
       last_value, max_value, min_value, increment_by,
       cycle, cache_size
FROM pg_sequences
ORDER BY schemaname, sequencename`,
	},
	{
		Name:        "stat_database",
		ArchivePath: "databases/%s/stat_database.tsv",
		Query: `SELECT datname,
       conflicts,
       deadlocks,
       temp_files,
       temp_bytes,
       stats_reset
FROM pg_stat_database
WHERE datname = current_database()`,
	},
	{
		Name:        "statistics",
		ArchivePath: "databases/%s/statistics.tsv",
		Query:       "SELECT * FROM pg_statistic_ext ORDER BY stxname",
	},
	{
		Name:        "subscription_tables",
		ArchivePath: "databases/%s/subscription_tables.tsv",
		Query:       "SELECT * FROM pg_subscription_rel ORDER BY srsubid, srrelid",
	},
	{
		// Ranked by freeze age, not size, because tables.tsv is capped at the
		// largest 1000 and a small table can hold the wraparound horizon back.
		// Includes pg_catalog and pg_toast relations, which tables.tsv excludes
		// and which are common culprits. Narrow columns only: no size
		// functions, so this stays cheap on an instance with many tables.
		// relfrozenxid = 0 relations are left out, having no tuples of their own.
		Name:        "table_freeze_age",
		ArchivePath: "databases/%s/table_freeze_age.tsv",
		Query: `SELECT n.nspname AS schemaname,
       c.relname AS tablename,
       c.relkind,
       c.relpages,
       age(c.relfrozenxid) AS relfrozenxid_age,
       mxid_age(c.relminmxid) AS relminmxid_age
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r', 'm', 't')
  AND c.relfrozenxid <> '0'::xid
ORDER BY GREATEST(age(c.relfrozenxid), mxid_age(c.relminmxid)) DESC
LIMIT 1000`,
	},
	{
		Name:        "tables",
		ArchivePath: "databases/%s/tables.tsv",
		Query: `
			WITH params AS MATERIALIZED (
			    SELECT (count(*) > 1000) AS many_tables
			    FROM pg_class c
			    JOIN pg_namespace n ON n.oid = c.relnamespace
			    WHERE c.relkind IN ('r', 'p')
			      AND n.nspname NOT IN ('pg_catalog', 'information_schema')
			      AND n.nspname NOT LIKE 'pg_toast%'
			)
			SELECT n.nspname AS schemaname,
			       c.relname AS tablename,
			       pg_get_userbyid(c.relowner) AS tableowner,
			       t.spcname AS tablespace,
			       c.relhasindex AS hasindexes,
			       c.relhasrules AS hasrules,
			       c.relhastriggers AS hastriggers,
			       c.relpersistence,
			       c.reltuples,
			       c.relpages,
			       c.reloptions,
			       CASE WHEN p.many_tables AND c.relpages > 0
			            THEN c.relpages::bigint * current_setting('block_size')::bigint
			            ELSE pg_relation_size(c.oid)
			       END AS heap_size,
			       CASE WHEN p.many_tables AND c.relpages > 0
			            THEN (c.relpages + COALESCE(toast.relpages, 0))::bigint
			                 * current_setting('block_size')::bigint
			            ELSE pg_table_size(c.oid)
			       END AS table_size,
			       c.reltoastrelid::regclass AS toast_table,
			       CASE WHEN c.reltoastrelid = 0 THEN NULL
			            WHEN p.many_tables AND toast.relpages > 0
			            THEN toast.relpages::bigint * current_setting('block_size')::bigint
			            ELSE pg_relation_size(c.reltoastrelid)
			       END AS toast_size,
			       -- A partitioned table carries relfrozenxid = 0, and
			       -- age('0'::xid) returns a huge number that reads as an
			       -- imminent wraparound, so the CASE guards leave those rows
			       -- empty. Transaction-ID age uses age() and multixact age
			       -- uses mxid_age().
			       CASE WHEN c.relfrozenxid <> '0'::xid THEN age(c.relfrozenxid) END
			           AS relfrozenxid_age,
			       CASE WHEN c.relminmxid <> '0'::xid THEN mxid_age(c.relminmxid) END
			           AS relminmxid_age,
			       s.n_live_tup,
			       s.n_dead_tup,
			       s.n_mod_since_analyze,
			       s.n_ins_since_vacuum,
			       s.last_vacuum,
			       s.last_autovacuum,
			       s.last_analyze,
			       s.last_autoanalyze,
			       EXTRACT(EPOCH FROM (clock_timestamp() - s.last_vacuum))::bigint
			           AS last_vacuum_age_seconds,
			       EXTRACT(EPOCH FROM (clock_timestamp() - s.last_autovacuum))::bigint
			           AS last_autovacuum_age_seconds,
			       EXTRACT(EPOCH FROM (clock_timestamp() - s.last_analyze))::bigint
			           AS last_analyze_age_seconds,
			       EXTRACT(EPOCH FROM (clock_timestamp() - s.last_autoanalyze))::bigint
			           AS last_autoanalyze_age_seconds
			FROM pg_class c
			CROSS JOIN params p
			JOIN pg_namespace n ON n.oid = c.relnamespace
			LEFT JOIN pg_class toast ON toast.oid = c.reltoastrelid
			LEFT JOIN pg_tablespace t ON t.oid = c.reltablespace
			LEFT JOIN pg_stat_all_tables s ON s.relid = c.oid
			WHERE c.relkind IN ('r', 'p')
			  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
			  AND n.nspname NOT LIKE 'pg_toast%'
			ORDER BY CASE WHEN p.many_tables AND c.relpages > 0
			              THEN (c.relpages + COALESCE(toast.relpages, 0))::bigint
			                   * current_setting('block_size')::bigint
			              ELSE pg_table_size(c.oid)
			         END DESC NULLS LAST,
			         schemaname, tablename
			LIMIT 1000
		`,
	},
	{
		Name:        "triggers",
		ArchivePath: "databases/%s/triggers.tsv",
		Query:       "SELECT * FROM pg_trigger ORDER BY tgname",
	},
	{
		Name:        "types",
		ArchivePath: "databases/%s/types.tsv",
		Query:       "SELECT oid, typname, typnamespace, typtype, typcategory FROM pg_type ORDER BY typname",
	},
}

// pg_statviz extension query tasks (sorted alphabetically by name)
// These are per-database tasks - ArchivePath will be formatted with dbname
var pgStatvizQueryTasks = []SimpleQueryTask{
	{
		Name:        "pg_statviz_blocking",
		ArchivePath: "pg_statviz/%s/blocking.tsv",
		Query:       "SELECT * FROM pgstatviz.blocking ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_buf",
		ArchivePath: "pg_statviz/%s/buf.tsv",
		Query:       "SELECT * FROM pgstatviz.buf ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_conf",
		ArchivePath: "pg_statviz/%s/conf.tsv",
		Query:       "SELECT * FROM pgstatviz.conf ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_conn",
		ArchivePath: "pg_statviz/%s/conn.tsv",
		Query:       "SELECT * FROM pgstatviz.conn ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_db",
		ArchivePath: "pg_statviz/%s/db.tsv",
		Query:       "SELECT * FROM pgstatviz.db ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_io",
		ArchivePath: "pg_statviz/%s/io.tsv",
		Query:       "SELECT * FROM pgstatviz.io ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_lock",
		ArchivePath: "pg_statviz/%s/lock.tsv",
		Query:       "SELECT * FROM pgstatviz.lock ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_repl",
		ArchivePath: "pg_statviz/%s/repl.tsv",
		Query:       "SELECT * FROM pgstatviz.repl ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_slru",
		ArchivePath: "pg_statviz/%s/slru.tsv",
		Query:       "SELECT * FROM pgstatviz.slru ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_snapshots",
		ArchivePath: "pg_statviz/%s/snapshots.tsv",
		Query:       "SELECT * FROM pgstatviz.snapshots ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_wait",
		ArchivePath: "pg_statviz/%s/wait.tsv",
		Query:       "SELECT * FROM pgstatviz.wait ORDER BY snapshot_tstamp",
	},
	{
		Name:        "pg_statviz_wal",
		ArchivePath: "pg_statviz/%s/wal.tsv",
		Query:       "SELECT * FROM pgstatviz.wal ORDER BY snapshot_tstamp",
	},
}

// Spock replication query tasks, one file per catalogue relation under
// spock/{dbname}/ (sorted alphabetically by name). Spock's catalogue is
// per-database, so these are generated per database like the pg_statviz ones.
//
// Three Spock relations carry data that must never enter an archive, because an
// archive gets attached to a support ticket: exception_log holds jsonb images of
// the conflicting rows, resolutions holds the same as text, and
// node_interface.if_dsn holds the node connection string with its password. The
// two tasks that read those tables name their columns explicitly instead of
// using SELECT *, and node_interface is not collected at all.
var spockQueryTasks = []SimpleQueryTask{
	{
		Name:        "spock_channel_summary_stats",
		ArchivePath: "spock/%s/channel_summary_stats.tsv",
		Query:       "SELECT * FROM spock.channel_summary_stats ORDER BY sub_name",
	},
	{
		// local_tup, remote_old_tup and remote_new_tup are excluded
		// deliberately: they are jsonb images of the conflicting rows.
		// error_message is kept: Spock stores the primary message with its
		// SQLSTATE and no DETAIL, so a constraint violation's offending key is
		// not included, and without the message the log says only that
		// something failed.
		Name:        "spock_exception_log",
		ArchivePath: "spock/%s/exception_log.tsv",
		Query: `SELECT remote_origin, remote_commit_ts, command_counter,
       retry_errored_at, remote_xid, local_origin, local_commit_ts,
       table_schema, table_name, operation, ddl_statement, ddl_user,
       error_message
FROM spock.exception_log
ORDER BY retry_errored_at DESC NULLS LAST
LIMIT 1000`,
	},
	{
		Name:        "spock_lag_tracker",
		ArchivePath: "spock/%s/lag_tracker.tsv",
		Query:       "SELECT * FROM spock.lag_tracker ORDER BY origin_name, receiver_name",
	},
	{
		Name:        "spock_local_node",
		ArchivePath: "spock/%s/local_node.tsv",
		Query:       "SELECT * FROM spock.local_node ORDER BY node_id",
	},
	{
		Name:        "spock_local_sync_status",
		ArchivePath: "spock/%s/local_sync_status.tsv",
		Query:       "SELECT * FROM spock.local_sync_status ORDER BY sync_subid, sync_nspname, sync_relname",
	},
	{
		// info is excluded deliberately: it is an operator-supplied jsonb
		// argument to spock.node_create, defaulting to NULL, so what it holds
		// depends on the deployment rather than on Spock.
		Name:        "spock_node",
		ArchivePath: "spock/%s/node.tsv",
		Query:       "SELECT node_id, node_name, location, country FROM spock.node ORDER BY node_name",
	},
	{
		// Column names only, no values: this table records which columns hold
		// personally identifiable information.
		Name:        "spock_pii",
		ArchivePath: "spock/%s/pii.tsv",
		Query:       "SELECT * FROM spock.pii ORDER BY pii_schema, pii_table, pii_column",
	},
	{
		Name:        "spock_progress",
		ArchivePath: "spock/%s/progress.tsv",
		Query:       "SELECT * FROM spock.progress ORDER BY node_id, remote_node_id",
	},
	{
		Name:        "spock_replication_set",
		ArchivePath: "spock/%s/replication_set.tsv",
		Query:       "SELECT * FROM spock.replication_set ORDER BY set_name",
	},
	{
		Name:        "spock_replication_set_table",
		ArchivePath: "spock/%s/replication_set_table.tsv",
		Query:       "SELECT * FROM spock.replication_set_table ORDER BY set_id, set_reloid",
	},
	{
		// local_tuple and remote_tuple are excluded deliberately: they are text
		// images of the conflicting rows.
		Name:        "spock_resolutions",
		ArchivePath: "spock/%s/resolutions.tsv",
		Query: `SELECT id, node_name, log_time, relname, idxname,
       conflict_type, conflict_resolution, local_origin, local_xid,
       local_timestamp, remote_origin, remote_xid, remote_timestamp,
       remote_lsn
FROM spock.resolutions
ORDER BY log_time DESC
LIMIT 1000`,
	},
	{
		Name:        "spock_subscription",
		ArchivePath: "spock/%s/subscription.tsv",
		Query:       "SELECT * FROM spock.subscription ORDER BY sub_name",
	},
	{
		Name:        "spock_tables",
		ArchivePath: "spock/%s/tables.tsv",
		Query:       "SELECT * FROM spock.tables ORDER BY nspname, relname",
	},
}

// buildQueryTasks converts SimpleQueryTask registry to CollectionTask slice
func buildQueryTasks(category string, tasks []SimpleQueryTask, db *sql.DB) []CollectionTask {
	result := make([]CollectionTask, len(tasks))
	for i, t := range tasks {
		result[i] = CollectionTask{
			Category:    category,
			Name:        t.Name,
			ArchivePath: t.ArchivePath,
			Collector:   pgQueryCollector(db, t.Query),
		}
	}
	return result
}

// buildConfigFileTasks converts SimpleConfigFileTask registry to CollectionTask slice
func buildConfigFileTasks(category string, tasks []SimpleConfigFileTask, db *sql.DB) []CollectionTask {
	result := make([]CollectionTask, len(tasks))
	for i, t := range tasks {
		filename := t.Filename // Capture loop variable
		result[i] = CollectionTask{
			Category:    category,
			Name:        t.Name,
			ArchivePath: t.ArchivePath,
			Collector: func(cfg *Config, w io.Writer) error {
				return collectPGConfigFile(db, cfg, filename, w)
			},
		}
	}
	return result
}
