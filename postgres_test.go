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
	"strings"
	"testing"
)

// TestPostgreSQLCollectors verifies all expected PostgreSQL collectors are registered
// with correct metadata using a table-driven approach
func TestPostgreSQLCollectors(t *testing.T) {
	// Expected collectors with their metadata
	expected := []struct {
		name        string
		archivePath string
	}{
		// Instance-level collectors (alphabetically ordered)
		{"activity", "postgresql/running_activity.tsv"},
		{"archiver", "postgresql/archiver.tsv"},
		{"available_extensions", "postgresql/available_extensions.tsv"},
		{"bgwriter", "postgresql/bgwriter.tsv"},
		{"blocking_locks", "postgresql/blocking_locks.tsv"},
		{"checkpointer", "postgresql/checkpointer.tsv"},
		{"configuration", "postgresql/configuration.tsv"},
		{"connection_summary", "postgresql/connection_summary.tsv"},
		{"database_conflicts", "postgresql/database_conflicts.tsv"},
		{"database_sizes", "postgresql/database_sizes.tsv"},
		{"databases", "postgresql/databases.tsv"},
		{"databases_blk", "postgresql/databases_blk.tsv"},
		{"databases_checksums", "postgresql/databases_checksums.tsv"},
		{"databases_tup", "postgresql/databases_tup.tsv"},
		{"databases_xact", "postgresql/databases_xact.tsv"},
		{"db_role_setting", "postgresql/db_role_setting.tsv"},
		{"file_settings", "postgresql/file_settings.tsv"},
		{"pg_hba.conf", "postgresql/pg_hba.conf"},
		{"pg_hba_file_rules", "postgresql/pg_hba_file_rules.tsv"},
		{"pg_ident.conf", "postgresql/pg_ident.conf"},
		{"postmaster_start_time", "postgresql/postmaster_start_time.tsv"},
		{"postgresql.auto.conf", "postgresql/postgresql.auto.conf"},
		{"postgresql.conf", "postgresql/postgresql.conf"},
		{"prepared_xacts", "postgresql/prepared_xacts.tsv"},
		{"recovery.conf", "postgresql/recovery.conf"},
		{"recovery.done", "postgresql/recovery.done"},
		{"replication", "postgresql/replication.tsv"},
		{"replication_origin", "postgresql/replication_origin.tsv"},
		{"replication_slots", "postgresql/replication_slots.tsv"},
		{"roles", "postgresql/roles.tsv"},
		{"running_activity_maxage", "postgresql/running_activity_maxage.tsv"},
		{"running_locks", "postgresql/running_locks.tsv"},
		{"shmem_allocations", "postgresql/shmem_allocations.tsv"},
		{"stat_io", "postgresql/stat_io.tsv"},
		{"stat_progress_analyze", "postgresql/stat_progress_analyze.tsv"},
		{"stat_progress_basebackup", "postgresql/stat_progress_basebackup.tsv"},
		{"stat_progress_cluster", "postgresql/stat_progress_cluster.tsv"},
		{"stat_progress_copy", "postgresql/stat_progress_copy.tsv"},
		{"stat_progress_create_index", "postgresql/stat_progress_create_index.tsv"},
		{"stat_progress_vacuum", "postgresql/stat_progress_vacuum.tsv"},
		{"stat_slru", "postgresql/stat_slru.tsv"},
		{"stat_statements_calls", "postgresql/stat_statements_calls.tsv"},
		{"stat_statements_max_time", "postgresql/stat_statements_max_time.tsv"},
		{"stat_statements_total_time", "postgresql/stat_statements_total_time.tsv"},
		{"stat_wal", "postgresql/stat_wal.tsv"},
		{"subscriptions", "postgresql/subscriptions.tsv"},
		{"tablespace_sizes", "postgresql/tablespace_sizes.tsv"},
		{"tablespaces", "postgresql/tablespaces.tsv"},
		{"version", "postgresql/version.tsv"},
		{"waits_sample", "postgresql/waits_sample.tsv"},
		{"wal_position", "postgresql/wal_position.tsv"},
		{"wal_receiver", "postgresql/wal_receiver.tsv"},
	}

	tasks := getPostgreSQLTasks(nil)

	if len(tasks) == 0 {
		t.Fatal("getPostgreSQLTasks returned no tasks")
	}

	// Build a map for easy lookup
	taskMap := make(map[string]*CollectionTask)
	for i := range tasks {
		taskMap[tasks[i].Name] = &tasks[i]
	}

	// Verify each expected collector exists with correct metadata
	for _, exp := range expected {
		t.Run(exp.name, func(t *testing.T) {
			task, found := taskMap[exp.name]
			if !found {
				t.Fatalf("collector %q not found in getPostgreSQLTasks()", exp.name)
			}

			if task.Category != "postgresql" {
				t.Errorf("expected category 'postgresql', got %q", task.Category)
			}

			if task.ArchivePath != exp.archivePath {
				t.Errorf("expected archive path %q, got %q", exp.archivePath, task.ArchivePath)
			}

			if task.Collector == nil {
				t.Fatal("collector function is nil")
			}
		})
	}

	// Verify we have the expected number of collectors
	if len(tasks) != len(expected) {
		t.Errorf("expected %d collectors, got %d", len(expected), len(tasks))
	}
}

// TestPostgreSQLTasksStructure verifies all PostgreSQL tasks have required fields
func TestPostgreSQLTasksStructure(t *testing.T) {
	tasks := getPostgreSQLTasks(nil)

	if len(tasks) == 0 {
		t.Fatal("getPostgreSQLTasks returned no tasks")
	}

	for i, task := range tasks {
		if task.Category == "" {
			t.Errorf("task %d missing Category", i)
		}
		if task.Name == "" {
			t.Errorf("task %d missing Name", i)
		}
		if task.ArchivePath == "" {
			t.Errorf("task %d missing ArchivePath", i)
		}
		if task.Collector == nil {
			t.Errorf("task %d missing Collector function", i)
		}

		// Verify all PostgreSQL tasks have category "postgresql"
		if task.Category != "postgresql" {
			t.Errorf("task %d (%s) has category %q, expected \"postgresql\"", i, task.Name, task.Category)
		}

		// Verify archive paths don't start with slash
		if strings.HasPrefix(task.ArchivePath, "/") {
			t.Errorf("task %d (%s) ArchivePath starts with /: %s", i, task.Name, task.ArchivePath)
		}
	}
}
