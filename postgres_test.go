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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
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
		{"control_checkpoint", "postgresql/control_checkpoint.tsv"},
		{"control_init", "postgresql/control_init.tsv"},
		{"control_recovery", "postgresql/control_recovery.tsv"},
		{"control_system", "postgresql/control_system.tsv"},
		{"database_conflicts", "postgresql/database_conflicts.tsv"},
		{"database_sizes", "postgresql/database_sizes.tsv"},
		{"databases", "postgresql/databases.tsv"},
		{"databases_blk", "postgresql/databases_blk.tsv"},
		{"databases_checksums", "postgresql/databases_checksums.tsv"},
		{"databases_tup", "postgresql/databases_tup.tsv"},
		{"databases_xact", "postgresql/databases_xact.tsv"},
		{"db_role_setting", "postgresql/db_role_setting.tsv"},
		{"file_settings", "postgresql/file_settings.tsv"},
		{"log_directory", "postgresql/log_directory.tsv"},
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
		{"stat_replication_slots", "postgresql/stat_replication_slots.tsv"},
		{"stat_slru", "postgresql/stat_slru.tsv"},
		{"stat_ssl", "postgresql/stat_ssl.tsv"},
		{"stat_statements_calls", "postgresql/stat_statements_calls.tsv"},
		{"stat_statements_max_time", "postgresql/stat_statements_max_time.tsv"},
		{"stat_statements_total_time", "postgresql/stat_statements_total_time.tsv"},
		{"stat_subscription_stats", "postgresql/stat_subscription_stats.tsv"},
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

// TestPGQueryCollectorUnavailableAsSkip verifies that PG errors for objects
// radar cannot reach (missing extension, or an existing object the collecting
// role has no rights on) are returned as SkipError, while real errors are
// returned as-is.
func TestPGQueryCollectorUnavailableAsSkip(t *testing.T) {
	tests := []struct {
		name     string
		pgErr    *pgconn.PgError
		wantSkip bool
	}{
		{"undefined_table is skip", &pgconn.PgError{Code: "42P01", Message: "relation \"pg_stat_statements\" does not exist"}, true},
		{"undefined_function is skip", &pgconn.PgError{Code: "42883", Message: "function does not exist"}, true},
		{"undefined_object is skip", &pgconn.PgError{Code: "42704", Message: "type does not exist"}, true},
		{"invalid_schema is skip", &pgconn.PgError{Code: "3F000", Message: "schema \"pgstatviz\" does not exist"}, true},
		{"insufficient_privilege is skip", &pgconn.PgError{Code: "42501", Message: "permission denied for schema spock"}, true},
		{"syntax_error is real error", &pgconn.PgError{Code: "42601", Message: "syntax error"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock: %v", err)
			}
			mock.ExpectQuery("SELECT").WillReturnError(tt.pgErr)

			collector := pgQueryCollector(db, "SELECT 1")
			err = collector(&Config{}, &bytes.Buffer{})

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var skipErr SkipError
			isSkip := errors.As(err, &skipErr)
			if isSkip != tt.wantSkip {
				t.Errorf("errors.As SkipError = %v, want %v (err: %v)", isSkip, tt.wantSkip, err)
			}
		})
	}
}

// TestWriteDirListingTSV verifies the directory listing shape: a header, one row
// per regular file with its size, and no rows for subdirectories.
func TestWriteDirListingTSV(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.log"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.log"), []byte("hi"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatalf("creating fixture subdirectory: %v", err)
	}

	var buf bytes.Buffer
	if err := writeDirListingTSV(dir, &buf); err != nil {
		t.Fatalf("writeDirListingTSV: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if lines[0] != "directory\tfilename\tsize_bytes\tmodified" {
		t.Errorf("header = %q", lines[0])
	}
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header plus two files): %q", len(lines), lines)
	}

	for i, want := range []struct {
		name string
		size string
	}{{"a.log", "5"}, {"b.log", "2"}} {
		fields := strings.Split(lines[i+1], "\t")
		if len(fields) != 4 {
			t.Fatalf("row %d has %d fields, want 4: %q", i, len(fields), lines[i+1])
		}
		if fields[0] != dir {
			t.Errorf("row %d directory = %q, want %q", i, fields[0], dir)
		}
		if fields[1] != want.name {
			t.Errorf("row %d filename = %q, want %q", i, fields[1], want.name)
		}
		if fields[2] != want.size {
			t.Errorf("row %d size_bytes = %q, want %q", i, fields[2], want.size)
		}
		if _, err := time.Parse(time.RFC3339, fields[3]); err != nil {
			t.Errorf("row %d modified = %q is not RFC3339: %v", i, fields[3], err)
		}
	}

	if strings.Contains(buf.String(), "sub") {
		t.Error("listing includes the subdirectory")
	}
}

// TestWriteDirListingTSVUnreadableIsSkip verifies that a directory radar cannot
// list costs it that one file rather than failing the run. That is the normal
// case when radar runs as an OS user without rights on the path.
func TestWriteDirListingTSVUnreadableIsSkip(t *testing.T) {
	var buf bytes.Buffer
	err := writeDirListingTSV(filepath.Join(t.TempDir(), "does-not-exist"), &buf)

	var skipErr SkipError
	if !errors.As(err, &skipErr) {
		t.Fatalf("err = %v, want SkipError", err)
	}
	if buf.Len() != 0 {
		t.Errorf("wrote %d bytes for an unreadable directory, want 0", buf.Len())
	}
}

// TestCollectLogDirectory verifies that log_directory is resolved against the
// data directory when relative, and used as given when absolute.
func TestCollectLogDirectory(t *testing.T) {
	dataDir := t.TempDir()
	logDir := filepath.Join(dataDir, "log")
	if err := os.Mkdir(logDir, 0o700); err != nil {
		t.Fatalf("creating fixture log directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "postgresql.log"), []byte("x"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	tests := []struct {
		name    string
		setting string
		cfg     *Config
		// showDataDir is true when the collector is expected to ask the server
		// for the data directory.
		showDataDir bool
	}{
		{"relative setting resolves against the data directory", "log", &Config{}, true},
		{"relative setting uses the configured data directory", "log", &Config{DataDir: dataDir}, false},
		{"absolute setting is used as given", logDir, &Config{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock: %v", err)
			}
			defer closeErrCheck(db, "mock database")

			mock.ExpectQuery("SHOW log_directory").
				WillReturnRows(sqlmock.NewRows([]string{"log_directory"}).AddRow(tt.setting))
			if tt.showDataDir {
				mock.ExpectQuery("SHOW data_directory").
					WillReturnRows(sqlmock.NewRows([]string{"data_directory"}).AddRow(dataDir))
			}

			var buf bytes.Buffer
			if err := collectLogDirectory(db, tt.cfg, &buf); err != nil {
				t.Fatalf("collectLogDirectory: %v", err)
			}

			if !strings.Contains(buf.String(), "postgresql.log") {
				t.Errorf("listing does not mention the log file: %q", buf.String())
			}
			if !strings.Contains(buf.String(), logDir) {
				t.Errorf("listing does not resolve to %q: %q", logDir, buf.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}

// TestCollectLogDirectoryNeverReadsContents verifies the listing carries file
// names and sizes only. Log bodies are far larger than anything else in an
// archive and carry query text and error detail beyond what the rest holds.
func TestCollectLogDirectoryNeverReadsContents(t *testing.T) {
	dataDir := t.TempDir()
	logDir := filepath.Join(dataDir, "log")
	if err := os.Mkdir(logDir, 0o700); err != nil {
		t.Fatalf("creating fixture log directory: %v", err)
	}
	const secret = "FATAL: password authentication failed for user hunter2"
	if err := os.WriteFile(filepath.Join(logDir, "postgresql.log"), []byte(secret), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer closeErrCheck(db, "mock database")
	mock.ExpectQuery("SHOW log_directory").
		WillReturnRows(sqlmock.NewRows([]string{"log_directory"}).AddRow(logDir))

	var buf bytes.Buffer
	if err := collectLogDirectory(db, &Config{}, &buf); err != nil {
		t.Fatalf("collectLogDirectory: %v", err)
	}

	if strings.Contains(buf.String(), "hunter2") {
		t.Errorf("listing leaked log file contents: %q", buf.String())
	}
}
