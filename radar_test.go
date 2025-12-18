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
			},
			contains: []string{
				"host=localhost",
				"port=5432",
				"dbname=testdb",
				"user=testuser",
				"sslmode=disable",
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
			},
			contains: []string{
				"host=dbhost",
				"port=5433",
				"dbname=mydb",
				"user=admin",
				"password=secret",
				"sslmode=disable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := tt.config.ConnectionString()
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

// Test collect with failures
func TestCollectWithFailures(t *testing.T) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer closeErrCheck(zipWriter, "zip writer")

	cfg := &Config{Verbose: false}

	tasks := []CollectionTask{
		{
			Category:    "test",
			Name:        "success",
			ArchivePath: "test/success.out",
			Collector: func(cfg *Config, w io.Writer) error {
				_, err := w.Write([]byte("ok"))
				return err
			},
		},
		{
			Category:    "test",
			Name:        "failure",
			ArchivePath: "test/fail.out",
			Collector: func(cfg *Config, w io.Writer) error {
				return bytes.ErrTooLarge
			},
		},
	}

	collected := collect(cfg, zipWriter, tasks)

	if collected != 1 {
		t.Errorf("expected 1 collected, got %d", collected)
	}
	// Note: We no longer track failed count separately - tasks that fail are simply not collected
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
