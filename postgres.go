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
// These are treated as skips rather than failures. A permission denial is
// excluded deliberately: the object exists and holds data, so being refused is
// reported per collector rather than skipped.
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

// generateDatabaseTasks creates per-database collection tasks, returning the
// closer for the connection they share.
func generateDatabaseTasks(db *sql.DB) ([]CollectionTask, io.Closer, error) {
	if db == nil {
		return nil, nil, fmt.Errorf("PostgreSQL not initialized")
	}

	// Get list of databases
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datallowconn ORDER BY datname")
	if err != nil {
		return nil, nil, fmt.Errorf("querying databases: %w", err)
	}
	defer closeErrCheck(rows, "database list query rows")

	var databases []string
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			return nil, nil, fmt.Errorf("scanning database name: %w", err)
		}
		// Always skip template0 and template1
		if dbname == "template0" || dbname == "template1" {
			continue
		}
		databases = append(databases, dbname)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating databases: %w", err)
	}

	// Generate tasks for each database
	var tasks []CollectionTask
	conns := &dbConns{}
	allDBTasks := make([]SimpleQueryTask, 0,
		len(perDatabaseQueryTasks)+len(pgStatvizQueryTasks)+len(spockQueryTasks))
	allDBTasks = append(allDBTasks, perDatabaseQueryTasks...)
	allDBTasks = append(allDBTasks, pgStatvizQueryTasks...)
	allDBTasks = append(allDBTasks, spockQueryTasks...)

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
					return conns.exec(cfg, dbName, td.Query, w)
				},
			})
		}
	}

	return tasks, conns, nil
}

// dbConns holds the connection the per-database tasks share. They are generated
// and run grouped by database, so holding one connection and closing it when
// collection moves on costs one connect and one authentication per database
// rather than one per task.
type dbConns struct {
	db   *sql.DB
	name string
}

// conn returns the connection for dbname, opening one on first use and closing
// the connection to the database collection has left behind. The database radar
// was invoked against is served by the connection opened at startup.
func (c *dbConns) conn(cfg *Config, dbname string) (*sql.DB, error) {
	if dbname == cfg.Database && cfg.DB != nil {
		return cfg.DB, nil
	}
	if c.db != nil {
		if dbname == c.name {
			return c.db, nil
		}
		closeErrCheck(c, "database connection")
	}

	db, err := sql.Open("pgx", cfg.ConnectionString(dbname))
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", dbname, err)
	}
	// Collectors run one at a time, so one connection is all the pool needs.
	db.SetMaxOpenConns(1)
	c.db, c.name = db, dbname
	return db, nil
}

// Close closes the connection the per-database tasks were sharing.
func (c *dbConns) Close() error {
	if c.db == nil {
		return nil
	}
	db := c.db
	c.db, c.name = nil, ""
	return db.Close()
}

// exec executes a query on a specific database
func (c *dbConns) exec(cfg *Config, dbname, query string, w io.Writer) error {
	db, err := c.conn(cfg, dbname)
	if err != nil {
		return err
	}

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
