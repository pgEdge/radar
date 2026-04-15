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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgconn"
)

// isPGUnavailableError reports whether err indicates that the queried object
// is not installed/available (missing extension, table, function, or schema).
// These are treated as skips rather than failures.
func isPGUnavailableError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case "42P01", // undefined_table
		"42704", // undefined_object
		"42883", // undefined_function
		"3F000": // invalid_schema_name
		return true
	}
	return false
}

// postgresConfigFileTasks defines tasks for collecting PostgreSQL configuration files (sorted alphabetically by name)
var postgresConfigFileTasks = []SimpleConfigFileTask{
	{
		Name:        "pg_hba.conf",
		ArchivePath: "postgresql/pg_hba.conf",
		Filename:    "pg_hba.conf",
	},
	{
		Name:        "pg_ident.conf",
		ArchivePath: "postgresql/pg_ident.conf",
		Filename:    "pg_ident.conf",
	},
	{
		Name:        "postgresql.auto.conf",
		ArchivePath: "postgresql/postgresql.auto.conf",
		Filename:    "postgresql.auto.conf",
	},
	{
		Name:        "postgresql.conf",
		ArchivePath: "postgresql/postgresql.conf",
		Filename:    "postgresql.conf",
	},
	{
		Name:        "recovery.conf",
		ArchivePath: "postgresql/recovery.conf",
		Filename:    "recovery.conf",
	},
	{
		Name:        "recovery.done",
		ArchivePath: "postgresql/recovery.done",
		Filename:    "recovery.done",
	},
}

// getPostgreSQLTasks returns PostgreSQL instance-level collection tasks
func getPostgreSQLTasks(db *sql.DB) []CollectionTask {
	// Build simple query tasks from registry
	tasks := buildQueryTasks("postgresql", postgresQueryTasks, db)

	// Build config file tasks
	tasks = append(tasks, buildConfigFileTasks("postgresql", postgresConfigFileTasks, db)...)

	return tasks
}

// collectPGConfigFile reads a PostgreSQL config file
func collectPGConfigFile(db *sql.DB, cfg *Config, filename string, w io.Writer) error {
	if db == nil {
		return fmt.Errorf("PostgreSQL not initialized")
	}

	// Auto-detect data directory if not provided
	if cfg.DataDir == "" {
		var dataDir string
		err := db.QueryRow("SHOW data_directory").Scan(&dataDir)
		if err != nil {
			return fmt.Errorf("detecting data directory: %w", err)
		}
		cfg.DataDir = dataDir
	}

	path := filepath.Join(cfg.DataDir, filename)
	data, err := readFile(path)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// generateDatabaseTasks creates per-database collection tasks
func generateDatabaseTasks(db *sql.DB) ([]CollectionTask, error) {
	if db == nil {
		return nil, fmt.Errorf("PostgreSQL not initialized")
	}

	// Get list of databases
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datallowconn ORDER BY datname")
	if err != nil {
		return nil, fmt.Errorf("querying databases: %w", err)
	}
	defer closeErrCheck(rows, "database list query rows")

	var databases []string
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			return nil, fmt.Errorf("scanning database name: %w", err)
		}
		// Always skip template0 and template1
		if dbname == "template0" || dbname == "template1" {
			continue
		}
		databases = append(databases, dbname)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating databases: %w", err)
	}

	// Generate tasks for each database
	var tasks []CollectionTask
	allDBTasks := append(perDatabaseQueryTasks, pgStatvizQueryTasks...)

	for _, dbname := range databases {
		// Capture loop variables for closure
		dbName := dbname

		for _, taskDef := range allDBTasks {
			// Capture loop variable for closure
			td := taskDef

			tasks = append(tasks, CollectionTask{
				Category:    "database",
				Name:        fmt.Sprintf("%s/%s", dbName, td.Name),
				ArchivePath: fmt.Sprintf(td.ArchivePath, dbName),
				Collector: func(cfg *Config, w io.Writer) error {
					return execPGQueryOnDB(dbName, cfg, td.Query, w)
				},
			})
		}
	}

	return tasks, nil
}

// execPGQueryOnDB executes a query on a specific database
func execPGQueryOnDB(dbname string, cfg *Config, query string, w io.Writer) error {
	db, err := sql.Open("pgx", cfg.ConnectionString(dbname))
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", dbname, err)
	}
	defer closeErrCheck(db, "database connection")

	rows, err := db.Query(query)
	if err != nil {
		if isPGUnavailableError(err) {
			return NewSkipError(err.Error())
		}
		return fmt.Errorf("query failed: %w", err)
	}
	defer closeErrCheck(rows, "query rows")

	return rowsToTSV(rows, w)
}

// printSummary logs the archive filename, size, and collector count.
func printSummary(totalCollected int, outputFile string, cfg *Config) {
	stat, err := os.Stat(outputFile)
	if err != nil {
		errorLog.Printf("Failed to stat archive: %v", err)
		return
	}

	// Format file size nicely (KB)
	sizeKB := stat.Size() / 1024

	if cfg.Verbose {
		// Verbose mode: show total collected
		infoLog.Printf("\n✓ Archive created: %s (%d KB)", outputFile, sizeKB)
		infoLog.Printf("  Total collectors: %d", totalCollected)
	} else {
		// Simple success message for default mode
		infoLog.Printf("✓ Archive created: %s (%d KB)", outputFile, sizeKB)
	}
}
