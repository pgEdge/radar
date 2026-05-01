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
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Test execCommand helper
func TestExecCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		shouldError bool
	}{
		{
			name:        "valid command - echo",
			command:     "echo",
			args:        []string{"test"},
			shouldError: false,
		},
		{
			name:        "invalid command",
			command:     "nonexistentcommand12345",
			args:        []string{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := execCommand(tt.command, tt.args...)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test readFile helper
func TestReadFile(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "radar_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	testContent := "test content\nline 2\n"
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "valid file",
			path:        tmpfile.Name(),
			shouldError: false,
		},
		{
			name:        "nonexistent file",
			path:        "/nonexistent/file/path/12345",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := readFile(tt.path)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && string(data) != testContent {
				t.Errorf("expected %q, got %q", testContent, string(data))
			}
		})
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

// TestQueryTaskColumnsCoverage asserts that the queries we expanded continue
// to reference the column tokens we rely on. Catches accidental column
// removal during future edits to these query strings.
func TestQueryTaskColumnsCoverage(t *testing.T) {
	checks := []struct {
		taskList    string
		taskName    string
		mustContain []string
	}{
		{"postgres", "databases", []string{
			"datfrozenxid", "datminmxid", "datconnlimit",
			"datistemplate", "datallowconn",
		}},
		{"perDB", "tables", []string{
			"n_live_tup", "n_dead_tup", "last_autovacuum", "last_analyze",
			"reltuples", "reloptions", "reltoastrelid", "relpersistence",
		}},
		{"perDB", "indexes", []string{
			"indrelid", "indclass", "indkey", "indisvalid",
			"idx_scan", "pg_relation_size",
		}},
		{"perDB", "sequences", []string{
			"pg_sequences", "last_value", "max_value", "min_value", "increment_by",
		}},
		{"postgres", "stat_ssl", []string{
			"pg_stat_ssl", "ssl", "cipher",
		}},
		{"postgres", "stat_replication_slots", []string{
			"pg_stat_replication_slots",
		}},
		{"perDB", "bloat", []string{
			"table_bloat_ratio", "wastedbytes",
		}},
		{"perDB", "pgstattuple", []string{
			"pgstattuple_approx",
		}},
	}

	taskByName := func(list []SimpleQueryTask, name string) *SimpleQueryTask {
		for i := range list {
			if list[i].Name == name {
				return &list[i]
			}
		}
		return nil
	}

	for _, c := range checks {
		var task *SimpleQueryTask
		switch c.taskList {
		case "postgres":
			task = taskByName(postgresQueryTasks, c.taskName)
		case "perDB":
			task = taskByName(perDatabaseQueryTasks, c.taskName)
		}
		if task == nil {
			t.Errorf("task %q not found in %s tasks", c.taskName, c.taskList)
			continue
		}
		for _, want := range c.mustContain {
			if !strings.Contains(task.Query, want) {
				t.Errorf("task %q query missing expected token %q", c.taskName, want)
			}
		}
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
