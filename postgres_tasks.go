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
		Query:       "SELECT * FROM pg_available_extensions ORDER BY name",
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
		Name:        "databases",
		ArchivePath: "postgresql/databases.tsv",
		Query:       "SELECT oid, datname, datdba, encoding, datcollate, datctype FROM pg_database ORDER BY datname",
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
    max(clock_timestamp() - backend_start) AS max_backend_age
FROM pg_stat_activity
WHERE state != 'idle'`,
	},
	{
		Name:        "running_locks",
		ArchivePath: "postgresql/running_locks.tsv",
		Query:       "SELECT * FROM pg_locks WHERE granted ORDER BY pid, locktype",
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
		Name:        "stat_progress_copy",
		ArchivePath: "postgresql/stat_progress_copy.tsv",
		Query:       "SELECT * FROM pg_stat_progress_copy",
	},
	{
		Name:        "stat_progress_vacuum",
		ArchivePath: "postgresql/stat_progress_vacuum.tsv",
		Query:       "SELECT * FROM pg_stat_progress_vacuum",
	},
	{
		Name:        "stat_slru",
		ArchivePath: "postgresql/stat_slru.tsv",
		Query:       "SELECT * FROM pg_stat_slru ORDER BY name",
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
}

// Per-database query tasks (sorted alphabetically by name)
// These are per-database tasks - ArchivePath will be formatted with dbname
var perDatabaseQueryTasks = []SimpleQueryTask{
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
			SELECT schemaname, tablename, indexname, indexdef
			FROM pg_indexes
			ORDER BY schemaname, tablename, indexname
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
		Name:        "tables",
		ArchivePath: "databases/%s/tables.tsv",
		Query: `
			SELECT schemaname, tablename, tableowner, tablespace, hasindexes, hasrules, hastriggers
			FROM pg_tables
			ORDER BY schemaname, tablename
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
