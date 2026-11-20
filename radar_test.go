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
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgproto3"
)

// TestExecCommand checks what execCommand writes for each outcome of a
// command, and how it classifies a failure.
func TestExecCommand(t *testing.T) {
	const skip = "skip"
	tests := []struct {
		name    string
		command string
		args    []string
		want    string // what reaches the writer
		wantErr string // "" for success, skip for a SkipError, else part of the error
	}{
		{"stdout", "echo", []string{"test"}, "test\n", ""},
		{"stderr in order with stdout", "sh", []string{"-c", "echo out; echo err >&2; echo out2"}, "out\nerr\nout2\n", ""},
		{"command not found", "nonexistentcommand12345", nil, "", skip},
		{"no data", "sh", []string{"-c", "echo 'No such file or directory' >&2; exit 1"}, "", skip},
		{"failure before any output is written", "sh", []string{"-c", "echo boom >&2; exit 2"}, "", "boom"},
		{"failure after output is written", "sh", []string{"-c", fmt.Sprintf("head -c %d /dev/zero; exit 2", 2*commandHeadSize)},
			strings.Repeat("\x00", 2*commandHeadSize), "exit status 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := execCommand(&buf, tt.command, tt.args...)
			var skipErr SkipError
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("unexpected error: %v", err)
			case tt.wantErr == skip && !errors.As(err, &skipErr):
				t.Errorf("want a SkipError, got %v", err)
			case tt.wantErr != "" && tt.wantErr != skip && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("want an error containing %q, got %v", tt.wantErr, err)
			}
			if buf.String() != tt.want {
				t.Errorf("wrote %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

// TestExecCommandNotVirtualized checks that the "none" systemd-detect-virt
// prints, exiting 1, on a host that is not virtualised is collected as output.
func TestExecCommandNotVirtualized(t *testing.T) {
	dir := t.TempDir()
	stub := "#!/bin/sh\necho none\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "systemd-detect-virt"), []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var buf bytes.Buffer
	if err := execCommand(&buf, "systemd-detect-virt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "none\n" {
		t.Errorf("wrote %q, want %q", buf.String(), "none\n")
	}
}

// TestCopyFile checks that copyFile writes a file's contents, and that a
// missing file is a skip.
func TestCopyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "radar_test.txt")
	content := "test content\nline 2\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := copyFile(path, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != content {
		t.Errorf("wrote %q, want %q", buf.String(), content)
	}

	var skipErr SkipError
	if err := copyFile("/nonexistent/file/path/12345", &buf); !errors.As(err, &skipErr) {
		t.Errorf("want a SkipError for a missing file, got %v", err)
	}
}

// streamedSize is large enough that a collector holding its output whole
// allocates many times maxCollectAlloc.
const streamedSize = 64 << 20

// maxCollectAlloc bounds the heap that collect may allocate while it streams
// streamedSize bytes into the archive. About 0.85 MiB of it is the zip
// writer's DEFLATE compressor, whose size is fixed.
const maxCollectAlloc = 1 << 20

// collectAlloc runs task through collect into a zip archive on disk. It returns
// the bytes allocated on the heap meanwhile and the uncompressed size of the
// entry the task wrote.
func collectAlloc(t *testing.T, task CollectionTask) (alloc, size uint64) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "radar.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeErrCheck(f, "archive")

	// Two collections empty the zip package's pool of compressors, so the one
	// this archive uses is allocated, and counted, inside the measurement.
	runtime.GC()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	zw := zip.NewWriter(f)
	collected := collect(&Config{}, zw, []CollectionTask{task})
	err = zw.Close()
	runtime.ReadMemStats(&after)
	if err != nil {
		t.Fatal(err)
	}
	if collected != 1 {
		t.Fatalf("collected %d tasks, want 1", collected)
	}

	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeErrCheck(r, "archive reader")
	alloc, size = after.TotalAlloc-before.TotalAlloc, r.File[0].UncompressedSize64
	t.Logf("collecting %d bytes allocated %d bytes", size, alloc)
	return alloc, size
}

// TestCollectStreamsCommandOutput checks that a command's output flows into the
// archive through fixed-size buffers rather than being held in memory whole.
func TestCollectStreamsCommandOutput(t *testing.T) {
	alloc, size := collectAlloc(t, CollectionTask{
		Name:        "large_command",
		ArchivePath: "test/large_command.out",
		Collector:   execCommandCollector("head", "-c", strconv.Itoa(streamedSize), "/dev/zero"),
	})
	if size != streamedSize {
		t.Errorf("archive entry holds %d bytes, want %d", size, streamedSize)
	}
	if alloc > maxCollectAlloc {
		t.Errorf("collecting %d bytes allocated %d bytes, want at most %d", size, alloc, maxCollectAlloc)
	}
}

// TestCollectStreamsFile checks that a file flows into the archive through
// fixed-size buffers rather than being held in memory whole.
func TestCollectStreamsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large_file")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(streamedSize); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	alloc, size := collectAlloc(t, CollectionTask{
		Name:        "large_file",
		ArchivePath: "test/large_file.out",
		Collector:   readFileCollector(path),
	})
	if size != streamedSize {
		t.Errorf("archive entry holds %d bytes, want %d", size, streamedSize)
	}
	if alloc > maxCollectAlloc {
		t.Errorf("collecting %d bytes allocated %d bytes, want at most %d", size, alloc, maxCollectAlloc)
	}
}

// Test rowsToTSV with mock data
func TestRowsToTSV(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		addRows  func(*sqlmock.Rows) *sqlmock.Rows
		expected string
	}{
		{
			name:     "simple values",
			columns:  []string{"id", "name"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(1, "Alice").AddRow(2, "Bob") },
			expected: "id\tname\n1\tAlice\n2\tBob\n",
		},
		{
			name:     "tab escaping",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("foo\tbar") },
			expected: "value\n\"foo\tbar\"\n",
		},
		{
			name:     "newline escaping",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("foo\nbar") },
			expected: "value\n\"foo\nbar\"\n",
		},
		{
			name:     "carriage return escaping",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("foo\rbar") },
			expected: "value\n\"foo\rbar\"\n",
		},
		{
			name:     "double quote escaping",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(`foo"bar`) },
			expected: "value\n\"foo\"\"bar\"\n",
		},
		{
			name:     "single quote - no escaping",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("foo'bar") },
			expected: "value\nfoo'bar\n",
		},
		{
			name:     "mixed quotes",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(`foo"bar'baz`) },
			expected: "value\n\"foo\"\"bar'baz\"\n",
		},
		{
			name:     "mixed special characters",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("foo\t\"bar\n") },
			expected: "value\n\"foo\t\"\"bar\n\"\n",
		},
		{
			name:     "null value",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(nil) },
			expected: "value\n\n",
		},
		{
			name:     "empty string",
			columns:  []string{"value"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow("") },
			expected: "value\n\n",
		},
		{
			name:     "byte array",
			columns:  []string{"data"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow([]byte("hello")) },
			expected: "data\nhello\n",
		},
		{
			name:     "byte array escaping",
			columns:  []string{"data"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow([]byte("{\"a\":\t1}")) },
			expected: "data\n\"{\"\"a\"\":\t1}\"\n",
		},
		{
			name:     "integer",
			columns:  []string{"num"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(42) },
			expected: "num\n42\n",
		},
		{
			name:     "float",
			columns:  []string{"val"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(3.14) },
			expected: "val\n3.14\n",
		},
		{
			name:     "empty result",
			columns:  []string{"col"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r },
			expected: "col\n",
		},
		{
			name:     "multiple columns",
			columns:  []string{"id", "name", "score"},
			addRows:  func(r *sqlmock.Rows) *sqlmock.Rows { return r.AddRow(1, "Alice", 95.5).AddRow(2, "Bob", 87.3) },
			expected: "id\tname\tscore\n1\tAlice\t95.5\n2\tBob\t87.3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock database
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock: %v", err)
			}

			// Configure mock to return our test data
			mock.ExpectQuery("SELECT").WillReturnRows(tt.addRows(sqlmock.NewRows(tt.columns)))
			mock.ExpectClose()

			// Execute query to get *sql.Rows
			rows, err := db.Query("SELECT *")
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}

			// Test rowsToTSV
			var buf bytes.Buffer
			if err := rowsToTSV(rows, &buf); err != nil {
				t.Fatalf("rowsToTSV failed: %v", err)
			}

			// Manually close rows and db to satisfy expectations before verification
			if err := rows.Close(); err != nil {
				t.Errorf("failed to close rows: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Errorf("failed to close mock db: %v", err)
			}

			// Verify output and mock expectations
			if buf.String() != tt.expected {
				t.Errorf("output mismatch\nexpected: %q\ngot:      %q", tt.expected, buf.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// Test Config.ConnectionString
func TestConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		contains []string
	}{
		{
			name: "basic config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				SSLMode:  "prefer",
			},
			contains: []string{
				"host=localhost",
				"port=5432",
				"dbname=testdb",
				"user=testuser",
				"sslmode=prefer",
			},
		},
		{
			name: "config with password",
			config: Config{
				Host:     "dbhost",
				Port:     5433,
				Database: "mydb",
				Username: "admin",
				Password: "secret",
				SSLMode:  "prefer",
			},
			contains: []string{
				"host=dbhost",
				"port=5433",
				"dbname=mydb",
				"user=admin",
				"password=secret",
				"sslmode=prefer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := tt.config.ConnectionString(tt.config.Database)
			for _, expected := range tt.contains {
				if !strings.Contains(connStr, expected) {
					t.Errorf("connection string missing %q: %s", expected, connStr)
				}
			}
		})
	}
}

// Test getSystemTasks returns valid tasks
func TestGetSystemTasks(t *testing.T) {
	tasks := getSystemTasks()

	if len(tasks) == 0 {
		t.Fatal("getSystemTasks returned no tasks")
	}

	// Verify all tasks have required fields
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
	}

	// Verify all system tasks have category "system"
	for i, task := range tasks {
		if task.Category != "system" {
			t.Errorf("task %d has category %q, expected \"system\"", i, task.Category)
		}
	}

	// Verify archive paths don't start with slash
	for i, task := range tasks {
		if strings.HasPrefix(task.ArchivePath, "/") {
			t.Errorf("task %d ArchivePath starts with /: %s", i, task.ArchivePath)
		}
	}
}

// Test getPostgreSQLTasks returns valid tasks
func TestGetPostgreSQLTasks(t *testing.T) {
	tasks := getPostgreSQLTasks(nil)

	if len(tasks) == 0 {
		t.Fatal("getPostgreSQLTasks returned no tasks")
	}

	// Verify all tasks have required fields
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
	}

	// Verify all postgresql tasks have category "postgresql"
	for i, task := range tasks {
		if task.Category != "postgresql" {
			t.Errorf("task %d has category %q, expected \"postgresql\"", i, task.Category)
		}
	}
}

// Test collect with mock tasks
func TestCollect(t *testing.T) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer closeErrCheck(zipWriter, "zip writer")

	cfg := &Config{Verbose: false}

	tasks := []CollectionTask{
		{
			Category:    "test",
			Name:        "seq_task1",
			ArchivePath: "test/seq1.out",
			Collector: func(cfg *Config, w io.Writer) error {
				_, err := w.Write([]byte("sequential 1"))
				return err
			},
		},
		{
			Category:    "test",
			Name:        "seq_task2",
			ArchivePath: "test/seq2.out",
			Collector: func(cfg *Config, w io.Writer) error {
				_, err := w.Write([]byte("sequential 2"))
				return err
			},
		},
	}

	collected := collect(cfg, zipWriter, tasks)

	if collected != 2 {
		t.Errorf("expected 2 collected, got %d", collected)
	}
}

// TestCollectWithFailures tests collect error handling
func TestCollectWithFailures(t *testing.T) {
	t.Run("skip error is silent", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		defer closeErrCheck(zipWriter, "zip writer")

		var errBuf bytes.Buffer
		errorLog.SetOutput(&errBuf)
		defer errorLog.SetOutput(os.Stderr)

		cfg := &Config{Verbose: false}
		tasks := []CollectionTask{
			{Category: "test", Name: "skip_task", ArchivePath: "test/skip.out",
				Collector: func(cfg *Config, w io.Writer) error {
					return NewSkipError("command not found: fake")
				}},
		}

		collected := collect(cfg, zipWriter, tasks)
		if collected != 0 {
			t.Errorf("expected 0 collected, got %d", collected)
		}
		if errBuf.Len() > 0 {
			t.Errorf("SkipError should not be logged, got: %s", errBuf.String())
		}
	})

	t.Run("real error is logged", func(t *testing.T) {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		defer closeErrCheck(zipWriter, "zip writer")

		var errBuf bytes.Buffer
		errorLog.SetOutput(&errBuf)
		defer errorLog.SetOutput(os.Stderr)

		cfg := &Config{Verbose: false}
		tasks := []CollectionTask{
			{Category: "test", Name: "real_error", ArchivePath: "test/fail.out",
				Collector: func(cfg *Config, w io.Writer) error {
					return bytes.ErrTooLarge
				}},
			{Category: "test", Name: "success", ArchivePath: "test/ok.out",
				Collector: func(cfg *Config, w io.Writer) error {
					_, err := w.Write([]byte("ok"))
					return err
				}},
		}

		collected := collect(cfg, zipWriter, tasks)
		if collected != 1 {
			t.Errorf("expected 1 collected, got %d", collected)
		}
		if !strings.Contains(errBuf.String(), "real_error") {
			t.Errorf("real error should be logged, got: %s", errBuf.String())
		}
	})
}

// TestNoDuplicateSystemArchivePaths verifies no duplicate archive paths in system tasks
func TestNoDuplicateSystemArchivePaths(t *testing.T) {
	tasks := getSystemTasks()
	seen := make(map[string]string)
	for _, task := range tasks {
		if prev, exists := seen[task.ArchivePath]; exists {
			t.Errorf("duplicate archive path %q: %q and %q", task.ArchivePath, prev, task.Name)
		}
		seen[task.ArchivePath] = task.Name
	}
}

// TestNoDuplicatePostgreSQLArchivePaths verifies no duplicate archive paths in PostgreSQL tasks
func TestNoDuplicatePostgreSQLArchivePaths(t *testing.T) {
	tasks := getPostgreSQLTasks(nil)
	seen := make(map[string]string)
	for _, task := range tasks {
		if prev, exists := seen[task.ArchivePath]; exists {
			t.Errorf("duplicate archive path %q: %q and %q", task.ArchivePath, prev, task.Name)
		}
		seen[task.ArchivePath] = task.Name
	}
}

// TestShouldRunTask covers -include / -exclude / default-disabled precedence.
func TestShouldRunTask(t *testing.T) {
	tests := []struct {
		name    string
		task    string
		exclude []string
		include []string
		want    bool
	}{
		{"default task runs", "stat_ssl", nil, nil, true},
		{"default-disabled task is skipped", "pgstattuple", nil, nil, false},
		{"per-db default-disabled task is skipped", "mydb/pgstattuple", nil, nil, false},
		{"include overrides default-disabled", "pgstattuple", nil, []string{"pgstattuple"}, true},
		{"include overrides per-db default-disabled", "mydb/pgstattuple", nil, []string{"pgstattuple"}, true},
		{"exclude skips a default-enabled task", "stat_ssl", []string{"stat_ssl"}, nil, false},
		{"exclude skips a per-db task", "mydb/bloat", []string{"bloat"}, nil, false},
		{"include wins over exclude", "pgstattuple", []string{"pgstattuple"}, []string{"pgstattuple"}, true},
		{"unrelated exclude does not affect task", "stat_ssl", []string{"bloat"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Exclude: tt.exclude, Include: tt.include}
			if got := shouldRunTask(tt.task, cfg); got != tt.want {
				t.Errorf("shouldRunTask(%q, exclude=%v, include=%v) = %v, want %v",
					tt.task, tt.exclude, tt.include, got, tt.want)
			}
		})
	}
}

// TestPerDatabaseTasksStructure verifies all per-database tasks have required fields
func TestPerDatabaseTasksStructure(t *testing.T) {
	for i, task := range perDatabaseQueryTasks {
		if task.Name == "" {
			t.Errorf("perDatabaseQueryTasks[%d] missing Name", i)
		}
		if task.ArchivePath == "" {
			t.Errorf("perDatabaseQueryTasks[%d] (%s) missing ArchivePath", i, task.Name)
		}
		if task.Query == "" {
			t.Errorf("perDatabaseQueryTasks[%d] (%s) missing Query", i, task.Name)
		}
		if !strings.Contains(task.ArchivePath, "%s") {
			t.Errorf("perDatabaseQueryTasks[%d] (%s) ArchivePath missing %%s placeholder: %s", i, task.Name, task.ArchivePath)
		}
	}
}

// queryColumnChecks lists the column tokens each expanded query must keep
// referencing, so an accidental column removal is caught.
var queryColumnChecks = []struct {
	tasks       []SimpleQueryTask
	taskName    string
	mustContain []string
}{
	{postgresQueryTasks, "databases", []string{
		"datfrozenxid", "datminmxid", "datconnlimit",
		"datistemplate", "datallowconn",
	}},
	{perDatabaseQueryTasks, "table_freeze_age", []string{
		// Ranked by freeze age rather than size, so a small table near the
		// wraparound threshold is included. Reaches pg_catalog and pg_toast,
		// which tables.tsv excludes.
		"age(c.relfrozenxid)", "mxid_age(c.relminmxid)",
		"GREATEST", "LIMIT 1000", "'r', 'm', 't'",
	}},
	{perDatabaseQueryTasks, "tables", []string{
		"n_live_tup", "n_dead_tup", "last_autovacuum", "last_analyze",
		"reltuples", "reloptions", "reltoastrelid", "relpersistence",
		"relpages", "pg_relation_size", "pg_table_size", "LIMIT 1000",
		"many_tables", "current_setting('block_size')",
		"last_vacuum_age_seconds", "last_autovacuum_age_seconds",
		"last_analyze_age_seconds", "last_autoanalyze_age_seconds",
		// Freeze age, with the guards that keep a partitioned table's
		// relfrozenxid = 0 from reading as an imminent wraparound.
		"relfrozenxid_age", "relminmxid_age",
		"age(c.relfrozenxid)", "mxid_age(c.relminmxid)",
		"c.relfrozenxid <> '0'::xid", "c.relminmxid <> '0'::xid",
	}},
	{perDatabaseQueryTasks, "indexes", []string{
		"indrelid", "indclass", "indkey", "indisvalid",
		"idx_scan", "pg_relation_size", "LIMIT 1000",
		"many_indexes", "current_setting('block_size')",
	}},
	{perDatabaseQueryTasks, "sequences", []string{
		"pg_sequences", "last_value", "max_value", "min_value", "increment_by",
	}},
	{postgresQueryTasks, "log_directory", []string{
		// pg_ls_logdir() is executable by pg_monitor and needs no filesystem
		// access. The pseudo-constant guard keeps it uncalled when the logging
		// collector is off, where the directory need not exist.
		"pg_ls_logdir()", "logging_collector",
	}},
	{postgresQueryTasks, "stat_ssl", []string{
		"pg_stat_ssl", "ssl", "cipher",
	}},
	{postgresQueryTasks, "stat_replication_slots", []string{
		"pg_stat_replication_slots",
	}},
	{perDatabaseQueryTasks, "bloat", []string{
		"table_bloat_ratio", "wastedbytes",
	}},
	{perDatabaseQueryTasks, "pgstattuple", []string{
		"pgstattuple_approx",
	}},
}

// TestQueryTaskColumnsCoverage asserts that the queries we expanded continue to
// reference the column tokens we rely on.
func TestQueryTaskColumnsCoverage(t *testing.T) {
	for _, c := range queryColumnChecks {
		task := findQueryTask(c.tasks, c.taskName)
		if task == nil {
			t.Errorf("task %q not found", c.taskName)
			continue
		}
		assertQueryContains(t, c.taskName, task.Query, c.mustContain)
	}
}

// findQueryTask returns the task named name, or nil when the registry has none.
func findQueryTask(list []SimpleQueryTask, name string) *SimpleQueryTask {
	for i := range list {
		if list[i].Name == name {
			return &list[i]
		}
	}
	return nil
}

// assertQueryContains checks that query mentions every token in want.
func assertQueryContains(t *testing.T, taskName, query string, want []string) {
	t.Helper()
	for _, token := range want {
		if !strings.Contains(query, token) {
			t.Errorf("task %q query missing expected token %q", taskName, token)
		}
	}
}

// TestPgStatvizCollectorsRegistered pins the set of pg_statviz tables radar
// collects, and the archive path each one lands on.
func TestPgStatvizCollectorsRegistered(t *testing.T) {
	expected := map[string]string{
		"pg_statviz_blocking":  "pg_statviz/%s/blocking.tsv",
		"pg_statviz_buf":       "pg_statviz/%s/buf.tsv",
		"pg_statviz_conf":      "pg_statviz/%s/conf.tsv",
		"pg_statviz_conn":      "pg_statviz/%s/conn.tsv",
		"pg_statviz_db":        "pg_statviz/%s/db.tsv",
		"pg_statviz_io":        "pg_statviz/%s/io.tsv",
		"pg_statviz_lock":      "pg_statviz/%s/lock.tsv",
		"pg_statviz_repl":      "pg_statviz/%s/repl.tsv",
		"pg_statviz_slru":      "pg_statviz/%s/slru.tsv",
		"pg_statviz_snapshots": "pg_statviz/%s/snapshots.tsv",
		"pg_statviz_wait":      "pg_statviz/%s/wait.tsv",
		"pg_statviz_wal":       "pg_statviz/%s/wal.tsv",
	}

	got := make(map[string]string, len(pgStatvizQueryTasks))
	for _, task := range pgStatvizQueryTasks {
		got[task.Name] = task.ArchivePath
	}

	for name, path := range expected {
		if got[name] != path {
			t.Errorf("pg_statviz task %q archive path = %q, want %q", name, got[name], path)
		}
	}
	if len(got) != len(expected) {
		t.Errorf("got %d pg_statviz tasks, want %d", len(got), len(expected))
	}
}

// TestPgStatvizTasksStructure verifies all pg_statviz tasks have required fields
func TestPgStatvizTasksStructure(t *testing.T) {
	for i, task := range pgStatvizQueryTasks {
		if task.Name == "" {
			t.Errorf("pgStatvizQueryTasks[%d] missing Name", i)
		}
		if task.ArchivePath == "" {
			t.Errorf("pgStatvizQueryTasks[%d] (%s) missing ArchivePath", i, task.Name)
		}
		if task.Query == "" {
			t.Errorf("pgStatvizQueryTasks[%d] (%s) missing Query", i, task.Name)
		}
		if !strings.Contains(task.ArchivePath, "%s") {
			t.Errorf("pgStatvizQueryTasks[%d] (%s) ArchivePath missing %%s placeholder: %s", i, task.Name, task.ArchivePath)
		}
	}
}

// TestSpockCollectorsRegistered pins the set of Spock catalogue relations radar
// collects, and the archive path each one lands on.
func TestSpockCollectorsRegistered(t *testing.T) {
	expected := map[string]string{
		"spock_channel_summary_stats": "spock/%s/channel_summary_stats.tsv",
		"spock_exception_log":         "spock/%s/exception_log.tsv",
		"spock_lag_tracker":           "spock/%s/lag_tracker.tsv",
		"spock_local_node":            "spock/%s/local_node.tsv",
		"spock_local_sync_status":     "spock/%s/local_sync_status.tsv",
		"spock_node":                  "spock/%s/node.tsv",
		"spock_pii":                   "spock/%s/pii.tsv",
		"spock_progress":              "spock/%s/progress.tsv",
		"spock_replication_set":       "spock/%s/replication_set.tsv",
		"spock_replication_set_table": "spock/%s/replication_set_table.tsv",
		"spock_resolutions":           "spock/%s/resolutions.tsv",
		"spock_subscription":          "spock/%s/subscription.tsv",
		"spock_tables":                "spock/%s/tables.tsv",
	}

	got := make(map[string]string, len(spockQueryTasks))
	for _, task := range spockQueryTasks {
		got[task.Name] = task.ArchivePath
	}

	for name, path := range expected {
		if got[name] != path {
			t.Errorf("Spock task %q archive path = %q, want %q", name, got[name], path)
		}
	}
	if len(got) != len(expected) {
		t.Errorf("got %d Spock tasks, want %d", len(got), len(expected))
	}
}

// TestSpockTasksStructure verifies all Spock tasks have required fields.
func TestSpockTasksStructure(t *testing.T) {
	for i, task := range spockQueryTasks {
		if task.Name == "" {
			t.Errorf("spockQueryTasks[%d] missing Name", i)
		}
		if task.ArchivePath == "" {
			t.Errorf("spockQueryTasks[%d] (%s) missing ArchivePath", i, task.Name)
		}
		if task.Query == "" {
			t.Errorf("spockQueryTasks[%d] (%s) missing Query", i, task.Name)
		}
		if !strings.Contains(task.ArchivePath, "%s") {
			t.Errorf("spockQueryTasks[%d] (%s) ArchivePath missing %%s placeholder: %s", i, task.Name, task.ArchivePath)
		}
	}
}

// TestSpockQueriesExcludeSensitiveColumns guards the three Spock relations that
// carry conflicting-row images or node connection strings. Those columns must
// never reach an archive, so no Spock query may name them, and the relations
// holding them may not be read with SELECT *.
func TestSpockQueriesExcludeSensitiveColumns(t *testing.T) {
	// Every column named, matched on a word boundary. None of these names is a
	// substring of another, so each one has to be listed: local_tup and
	// local_tuple are columns of different relations.
	forbidden := []string{
		"local_tup", "remote_old_tup", "remote_new_tup", // exception_log row images
		"local_tuple", "remote_tuple", "resolution_details", // resolutions row images
		"if_dsn", "node_interface", // node connection string, password included
		"info", // spock.node, deployment-defined jsonb
	}
	sensitive := regexp.MustCompile(`\b(` + strings.Join(forbidden, "|") + `)\b`)

	for _, task := range spockQueryTasks {
		if found := sensitive.FindString(task.Query); found != "" {
			t.Errorf("Spock task %q query names sensitive column %q", task.Name, found)
		}
	}

	// The relations carrying those columns must use explicit column lists.
	starRelations := []struct {
		task     string
		relation string
	}{
		{"spock_exception_log", "spock.exception_log"},
		{"spock_resolutions", "spock.resolutions"},
		{"spock_node", "spock.node"},
	}
	byName := make(map[string]SimpleQueryTask, len(spockQueryTasks))
	for _, task := range spockQueryTasks {
		byName[task.Name] = task
	}
	for _, r := range starRelations {
		task, found := byName[r.task]
		if !found {
			t.Errorf("Spock task %q not registered", r.task)
			continue
		}
		if strings.Contains(task.Query, "SELECT * FROM "+r.relation) {
			t.Errorf("Spock task %q reads %s with SELECT *; an explicit column list is required", r.task, r.relation)
		}
	}
}

// archivePathsByName maps task name to archive path, asserting every task is in
// the "database" category on the way through.
func archivePathsByName(t *testing.T, tasks []CollectionTask) map[string]string {
	t.Helper()
	paths := make(map[string]string, len(tasks))
	for _, task := range tasks {
		if task.Category != "database" {
			t.Errorf("task %q has category %q, want \"database\"", task.Name, task.Category)
		}
		paths[task.Name] = task.ArchivePath
	}
	return paths
}

// TestGenerateDatabaseTasksRegistersAllRegistries traces the full path from the
// three per-database registries to the archive paths they produce, and confirms
// the template databases are left out.
func TestGenerateDatabaseTasksRegistersAllRegistries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer closeErrCheck(db, "mock database")

	mock.ExpectQuery("SELECT datname, has_database_privilege").
		WillReturnRows(sqlmock.NewRows([]string{"datname", "has_database_privilege"}).
			AddRow("mydb", true).AddRow("template0", true).AddRow("template1", true))

	tasks, conns, err := generateDatabaseTasks(db)
	if err != nil {
		t.Fatalf("generateDatabaseTasks: %v", err)
	}
	defer closeErrCheck(conns, "database connection")

	paths := archivePathsByName(t, tasks)

	// One representative per registry: per-database, pg_statviz, Spock. The
	// template databases must contribute nothing, so their entries expect "".
	expected := map[string]string{
		"mydb/tables":              "databases/mydb/tables.tsv",
		"mydb/pg_statviz_blocking": "pg_statviz/mydb/blocking.tsv",
		"mydb/spock_lag_tracker":   "spock/mydb/lag_tracker.tsv",
		"template0/tables":         "",
		"template1/tables":         "",
	}
	for name, want := range expected {
		if paths[name] != want {
			t.Errorf("task %q archive path = %q, want %q", name, paths[name], want)
		}
	}

	wantCount := len(perDatabaseQueryTasks) + len(pgStatvizQueryTasks) + len(spockQueryTasks)
	if len(tasks) != wantCount {
		t.Errorf("got %d tasks for one database, want %d", len(tasks), wantCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestLazyZipWriter verifies the lazy ZIP writer prevents empty entries
func TestLazyZipWriter(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	header := &zip.FileHeader{
		Name:   "test.txt",
		Method: zip.Deflate,
	}

	// Test: no writes should not create an entry
	lazy := &lazyZipWriter{zipWriter: zw, header: header}
	if lazy.WroteAny() {
		t.Error("WroteAny() should be false before any writes")
	}

	// Empty write should not create entry
	n, err := lazy.Write([]byte{})
	if err != nil {
		t.Errorf("empty Write returned error: %v", err)
	}
	if n != 0 {
		t.Errorf("empty Write returned n=%d, want 0", n)
	}
	if lazy.WroteAny() {
		t.Error("WroteAny() should be false after empty write")
	}

	// Real write should create entry
	n, err = lazy.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write returned n=%d, want 5", n)
	}
	if !lazy.WroteAny() {
		t.Error("WroteAny() should be true after write")
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	// Verify the ZIP contains the entry
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("failed to open zip reader: %v", err)
	}
	if len(reader.File) != 1 {
		t.Errorf("expected 1 file in zip, got %d", len(reader.File))
	}
	if len(reader.File) > 0 && reader.File[0].Name != "test.txt" {
		t.Errorf("expected file name 'test.txt', got %q", reader.File[0].Name)
	}
}

// TestLazyZipWriterNoWrite verifies no entry is created when nothing is written
func TestLazyZipWriterNoWrite(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	header := &zip.FileHeader{
		Name:   "should_not_exist.txt",
		Method: zip.Deflate,
	}

	lazy := &lazyZipWriter{zipWriter: zw, header: header}
	_ = lazy // not used, no writes

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("failed to open zip reader: %v", err)
	}
	if len(reader.File) != 0 {
		t.Errorf("expected 0 files in zip, got %d", len(reader.File))
	}
}

// TestGlobalOptions tests that -help and -version are answered whatever else
// is on the command line, that no other flag takes effect when they are, and
// that -help outranks -version when both are given.
func TestGlobalOptions(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name  string
		args  []string
		want  error
		check func(*testing.T, *Config)
	}{
		{name: "--help", args: []string{"radar", "--help"}, want: errHelpRequested},
		{name: "-help", args: []string{"radar", "-help"}, want: errHelpRequested},
		{name: "--version", args: []string{"radar", "--version"}, want: errVersionRequested},
		{name: "-version", args: []string{"radar", "-version"}, want: errVersionRequested},
		{name: "-V", args: []string{"radar", "-V"}, want: errVersionRequested},
		{name: "help outranks version", args: []string{"radar", "--version", "--help"}, want: errHelpRequested},
		{name: "help outranks version whatever the order", args: []string{"radar", "--help", "--version"}, want: errHelpRequested},
		{name: "help outranks an unparseable flag", args: []string{"radar", "-nosuchflag", "--help"}, want: errHelpRequested},
		{name: "version outranks an unparseable flag", args: []string{"radar", "-nosuchflag", "-V"}, want: errVersionRequested},
		{name: "help outranks a collection flag", args: []string{"radar", "-d", "testdb", "--help"}, want: errHelpRequested},
		{name: "neither given", args: []string{"radar", "-d", "testdb"}},
		{
			name: "-h is the database host",
			args: []string{"radar", "-h", "localhost"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Host != "localhost" {
					t.Errorf("host = %q, want %q", cfg.Host, "localhost")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
			os.Args = tt.args

			cfg, err := parseConfig()
			if !errors.Is(err, tt.want) {
				t.Fatalf("parseConfig() error = %v, want %v", err, tt.want)
			}
			if tt.want != nil {
				if cfg != nil {
					t.Errorf("expected no config alongside %v, got %+v", tt.want, cfg)
				}
				return
			}
			if cfg == nil {
				t.Fatal("expected a config when neither option was given")
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

// TestPrintUsage tests that the help text lists the global options above the
// collection flags. It must carry nothing build-dependent, because it is
// pasted verbatim into README.md and docs/index.md.
func TestPrintUsage(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
	os.Args = []string{"radar", "--help"}
	if _, err := parseConfig(); !errors.Is(err, errHelpRequested) {
		t.Fatalf("parseConfig() error = %v, want errHelpRequested", err)
	}

	var buf bytes.Buffer
	printUsage(&buf)
	got := buf.String()

	want := []string{
		"Usage: radar [options]",
		"GLOBAL OPTIONS:",
		"--help         show help",
		"--version, -V  print the version",
		"Options:",
		"-sslmode string",
		"database host",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("usage output missing %q\ngot:\n%s", w, got)
		}
	}

	if strings.Contains(got, version) {
		t.Errorf("usage output must not carry the build version\ngot:\n%s", got)
	}

	// The global options are answered before parsing, so registering them
	// would list a second, unreachable copy among the collection flags.
	for _, name := range []string{"help", "version", "V"} {
		if flag.CommandLine.Lookup(name) != nil {
			t.Errorf("-%s must not be a registered flag", name)
		}
	}
	if f := flag.CommandLine.Lookup("h"); f == nil || f.Usage != "database host" {
		t.Error("-h must stay bound to the database host")
	}
}

// TestPGEnvFallbacks tests PGPORT and PGDATABASE environment variable fallbacks.
func TestPGEnvFallbacks(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("PGPORT fallback", func(t *testing.T) {
		t.Setenv("PGPORT", "5433")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar", "--skip-system", "-d", "testdb"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 5433 {
			t.Errorf("expected port 5433, got %d", cfg.Port)
		}
	})

	t.Run("PGPORT flag takes precedence", func(t *testing.T) {
		t.Setenv("PGPORT", "5433")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar", "--skip-system", "-d", "testdb", "-p", "5434"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 5434 {
			t.Errorf("expected port 5434, got %d", cfg.Port)
		}
	})

	t.Run("PGDATABASE fallback", func(t *testing.T) {
		t.Setenv("PGDATABASE", "envdb")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Database != "envdb" {
			t.Errorf("expected database 'envdb', got %q", cfg.Database)
		}
	})

	t.Run("PGDATABASE flag takes precedence", func(t *testing.T) {
		t.Setenv("PGDATABASE", "envdb")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar", "-d", "flagdb"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Database != "flagdb" {
			t.Errorf("expected database 'flagdb', got %q", cfg.Database)
		}
	})
}

// TestSkipFlagValidation tests validation of skip flag combinations
func TestSkipFlagValidation(t *testing.T) {
	// Save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "both skip flags",
			args:        []string{"radar", "--skip-system", "--skip-postgres"},
			expectError: true,
			errorMsg:    "cannot use --skip-system and --skip-postgres together",
		},
		{
			name:        "skip-system without database",
			args:        []string{"radar", "--skip-system"},
			expectError: true,
			errorMsg:    "--skip-system requires PostgreSQL database",
		},
		{
			name:        "skip-system with database",
			args:        []string{"radar", "--skip-system", "-d", "testdb"},
			expectError: false,
		},
		{
			name:        "skip-postgres only",
			args:        []string{"radar", "--skip-postgres"},
			expectError: false,
		},
		{
			name:        "no skip flags",
			args:        []string{"radar"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package for each test
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ContinueOnError)
			os.Args = tt.args

			cfg, err := parseConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if cfg == nil {
					t.Error("expected config, got nil")
				}
			}
		})
	}
}

// TestSSLModeDefault verifies sslmode defaults to prefer.
func TestSSLModeDefault(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
	os.Args = []string{"radar"}

	cfg, err := parseConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SSLMode != "prefer" {
		t.Errorf("expected SSLMode=prefer, got %q", cfg.SSLMode)
	}
	if !strings.Contains(cfg.ConnectionString(cfg.Database), "sslmode=prefer") {
		t.Errorf("expected sslmode=prefer in connection string: %s", cfg.ConnectionString(cfg.Database))
	}
}

// TestSSLModeValidation verifies invalid sslmode values are rejected.
func TestSSLModeValidation(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name        string
		sslmode     string
		expectError bool
	}{
		{"prefer is valid", "prefer", false},
		{"disable is valid", "disable", false},
		{"require is valid", "require", false},
		{"invalid value rejected", "bogus", true},
		{"allow is not supported", "allow", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
			os.Args = []string{"radar", "--sslmode", tt.sslmode}

			_, err := parseConfig()
			if tt.expectError && err == nil {
				t.Errorf("expected error for sslmode=%q, got nil", tt.sslmode)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for sslmode=%q: %v", tt.sslmode, err)
			}
		})
	}
}

// TestSSLModeEnvFallback verifies PGSSLMODE env var is respected.
func TestSSLModeEnvFallback(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("PGSSLMODE fallback", func(t *testing.T) {
		t.Setenv("PGSSLMODE", "require")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.SSLMode != "require" {
			t.Errorf("expected SSLMode=require, got %q", cfg.SSLMode)
		}
	})

	t.Run("flag takes precedence over PGSSLMODE", func(t *testing.T) {
		t.Setenv("PGSSLMODE", "require")
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar", "--sslmode", "disable"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.SSLMode != "disable" {
			t.Errorf("expected SSLMode=disable, got %q", cfg.SSLMode)
		}
	})
}

// TestSSLCertValidation verifies cert/key pairing and rootcert requirements.
func TestSSLCertValidation(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	certFile, err := os.CreateTemp("", "cert*.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErrCheck(certFile, "certFile"); os.Remove(certFile.Name()) }) //nolint:errcheck

	keyFile, err := os.CreateTemp("", "key*.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErrCheck(keyFile, "keyFile"); os.Remove(keyFile.Name()) }) //nolint:errcheck

	rootCertFile, err := os.CreateTemp("", "rootcert*.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErrCheck(rootCertFile, "rootCertFile"); os.Remove(rootCertFile.Name()) }) //nolint:errcheck

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"sslcert without sslkey rejected", []string{"radar", "--sslcert", certFile.Name()}, true},
		{"sslkey without sslcert rejected", []string{"radar", "--sslkey", keyFile.Name()}, true},
		{"verify-ca without sslrootcert rejected", []string{"radar", "--sslmode", "verify-ca"}, true},
		{"verify-full without sslrootcert rejected", []string{"radar", "--sslmode", "verify-full"}, true},
		{"sslcert with sslkey valid", []string{"radar", "--sslcert", certFile.Name(), "--sslkey", keyFile.Name()}, false},
		{"verify-ca with sslrootcert valid", []string{"radar", "--sslmode", "verify-ca", "--sslrootcert", rootCertFile.Name()}, false},
		{"nonexistent sslcert rejected", []string{"radar", "--sslcert", "/nonexistent/cert.pem", "--sslkey", keyFile.Name()}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
			os.Args = tt.args
			_, err := parseConfig()
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestSSLCertEnvFallbacks verifies PGSSLCERT/PGSSLKEY/PGSSLROOTCERT env vars.
func TestSSLCertEnvFallbacks(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	certFile, err := os.CreateTemp("", "cert*.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErrCheck(certFile, "certFile"); os.Remove(certFile.Name()) }) //nolint:errcheck

	keyFile, err := os.CreateTemp("", "key*.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeErrCheck(keyFile, "keyFile"); os.Remove(keyFile.Name()) }) //nolint:errcheck

	t.Run("PGSSLCERT and PGSSLKEY env vars", func(t *testing.T) {
		t.Setenv("PGSSLCERT", certFile.Name())
		t.Setenv("PGSSLKEY", keyFile.Name())
		flag.CommandLine = flag.NewFlagSet("radar", flag.ContinueOnError)
		os.Args = []string{"radar"}

		cfg, err := parseConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.SSLCert != certFile.Name() {
			t.Errorf("expected SSLCert=%q, got %q", certFile.Name(), cfg.SSLCert)
		}
		if cfg.SSLKey != keyFile.Name() {
			t.Errorf("expected SSLKey=%q, got %q", keyFile.Name(), cfg.SSLKey)
		}
	})
}

// TestConnectionStringWithCerts verifies cert params appear in connection string.
func TestConnectionStringWithCerts(t *testing.T) {
	cfg := Config{
		Host:        "localhost",
		Port:        5432,
		Database:    "testdb",
		Username:    "testuser",
		SSLMode:     "verify-full",
		SSLCert:     "/path/to/cert.pem",
		SSLKey:      "/path/to/key.pem",
		SSLRootCert: "/path/to/root.pem",
	}

	s := cfg.ConnectionString(cfg.Database)
	for _, want := range []string{
		"sslmode=verify-full",
		"sslcert=/path/to/cert.pem",
		"sslkey=/path/to/key.pem",
		"sslrootcert=/path/to/root.pem",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in connection string: %s", want, s)
		}
	}
}

// TestConnectionStringQuoting verifies values with special characters are quoted.
func TestConnectionStringQuoting(t *testing.T) {
	cfg := Config{
		Host: "localhost", Port: 5432, Database: "testdb",
		Username: "testuser", Password: "pass word", SSLMode: "prefer",
		SSLCert: "/path/with spaces/cert.pem", SSLKey: "/path/with spaces/key.pem",
	}
	s := cfg.ConnectionString(cfg.Database)
	for _, want := range []string{
		"password='pass word'",
		"sslcert='/path/with spaces/cert.pem'",
		"sslkey='/path/with spaces/key.pem'",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in connection string: %s", want, s)
		}
	}

	cfg2 := Config{
		Host: "localhost", Port: 5432, Database: "db",
		Username: "user", Password: "it's", SSLMode: "prefer",
	}
	s2 := cfg2.ConnectionString(cfg2.Database)
	if !strings.Contains(s2, `password='it\'s'`) {
		t.Errorf("expected escaped single quote in: %s", s2)
	}
}

// A server that asks for GSSAPI authentication must reach the Kerberos
// provider, which then reads the krb5.conf that KRB5_CONFIG names.
func TestGSSAPIAuthUsesKerberosProvider(t *testing.T) {
	krb5Conf := filepath.Join(t.TempDir(), "absent-krb5.conf")
	t.Setenv("KRB5_CONFIG", krb5Conf)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer closeErrCheck(ln, "listener")
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer closeErrCheck(conn, "server connection")
		be := pgproto3.NewBackend(conn, conn)
		if _, err := be.ReceiveStartupMessage(); err != nil {
			return
		}
		be.Send(&pgproto3.AuthenticationGSS{})
		_ = be.Flush()
	}()

	cfg := Config{
		Host: "127.0.0.1", Port: ln.Addr().(*net.TCPAddr).Port,
		Database: "db", Username: "user", SSLMode: "disable",
	}
	db, err := sql.Open("pgx", cfg.ConnectionString(cfg.Database))
	if err != nil {
		t.Fatal(err)
	}
	defer closeErrCheck(db, "database")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err == nil {
		t.Fatal("expected the GSSAPI handshake to fail without a krb5.conf")
	}
	if !strings.Contains(err.Error(), krb5Conf) {
		t.Errorf("expected the Kerberos provider to read %s, got: %v", krb5Conf, err)
	}
}
