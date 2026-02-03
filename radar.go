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
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Date/Time Formats
const TimestampFormat = "20060102-150405"

// PostgreSQL Defaults
const (
	DefaultPostgresPort = 5432
	DefaultDatabase     = "postgres"
)

// Archive Settings
const DefaultCompressionMethod = zip.Deflate

// Error message patterns for skip detection
var (
	ExecutableNotFoundPatterns = []string{"executable file not found", "command not found"}
	NoDataPatterns             = []string{"No such file or directory", "no match", "doesn't exist"}
)

// Exit codes
const (
	ExitUsageError   = 1
	ExitCollectError = 3
	ExitNoData       = 4
)

// Config holds connection parameters and collection settings
type Config struct {
	// PostgreSQL connection
	Host     string
	Port     int
	Database string
	Username string
	Password string
	DataDir  string

	// Database connection (injected)
	DB *sql.DB

	// Collection control
	SkipSystem   bool
	SkipPostgres bool
	Verbose      bool
	VeryVerbose  bool
}

// CollectionTask defines a single data collection task
type CollectionTask struct {
	Category    string // "system", "postgresql", "database"
	Name        string // Descriptive name for logging
	ArchivePath string // Path within ZIP archive
	Collector   func(*Config, io.Writer) error
}

// lazyZipWriter defers ZIP entry creation until first Write()
// This prevents empty files in the archive when collectors produce no output
type lazyZipWriter struct {
	zipWriter *zip.Writer
	header    *zip.FileHeader
	writer    io.Writer // nil until first Write()
}

func (w *lazyZipWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil // Ignore empty writes
	}
	if w.writer == nil {
		var err error
		w.writer, err = w.zipWriter.CreateHeader(w.header)
		if err != nil {
			return 0, err
		}
	}
	return w.writer.Write(p)
}

func (w *lazyZipWriter) WroteAny() bool {
	return w.writer != nil
}

// SkipError indicates a collector was skipped (tool missing, no data, not applicable)
type SkipError struct {
	Reason string
}

func (e SkipError) Error() string {
	return e.Reason
}

// NewSkipError creates a new skip error
func NewSkipError(reason string) error {
	return SkipError{Reason: reason}
}

// isCommandNotFoundError checks if error indicates missing executable
func isCommandNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Check error message
	for _, pattern := range ExecutableNotFoundPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Check exit code
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 127 {
			return true
		}
	}

	return false
}

// isNoDataAvailable checks if output indicates data unavailable
func isNoDataAvailable(output string) bool {
	outputStr := strings.TrimSpace(output)

	if outputStr == "" {
		return true
	}

	// Check for common "no data" patterns, case-insensitively
	lowerOutput := strings.ToLower(outputStr)
	for _, pattern := range NoDataPatterns {
		if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// handleSpecialCases handles known special case outputs
func handleSpecialCases(name string, output []byte) ([]byte, bool) {
	// systemd-detect-virt returns "none" with exit 1 when not virtualized
	if name == "systemd-detect-virt" && strings.TrimSpace(string(output)) == "none" {
		return []byte("none\n"), true
	}

	return nil, false
}

var (
	infoLog  = log.New(os.Stderr, "", 0)
	errorLog = log.New(os.Stderr, "ERROR: ", 0)
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		errorLog.Println(err)
		flag.Usage()
		os.Exit(ExitUsageError)
	}

	// Generate output filename
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	timestamp := time.Now().Format(TimestampFormat)
	outputFile := fmt.Sprintf("radar-%s-%s.zip", hostname, timestamp)

	// Simplified output for non-verbose mode
	if !cfg.Verbose {
		infoLog.Println("Collecting diagnostic data...")
	}

	// Connect to PostgreSQL if not skipped
	if !cfg.SkipPostgres {
		if cfg.Verbose {
			infoLog.Printf("Connecting to PostgreSQL at %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
		}
		if err := initPostgreSQL(cfg); err != nil {
			errorLog.Printf("Could not connect to PostgreSQL: %v", err)
			errorLog.Println("Continuing with system data collection only...")
			cfg.SkipPostgres = true
		} else {
			defer closeErrCheck(cfg.DB, "database connection")
			if cfg.Verbose {
				infoLog.Println("PostgreSQL connected")
			}
		}
	}

	// Create output file (don't announce it unless verbose)
	if cfg.Verbose {
		infoLog.Printf("Creating archive: %s", outputFile)
	}
	outFile, err := os.Create(outputFile)
	if err != nil {
		errorLog.Printf("Failed to create output file: %v", err)
		os.Exit(ExitCollectError)
	}
	defer closeErrCheck(outFile, "output file")

	// Create ZIP writer
	zipWriter := zip.NewWriter(outFile)

	// Collect all data
	if cfg.Verbose {
		infoLog.Println("Starting data collection...")
	}
	totalCollected := collectAll(cfg, zipWriter)

	// Close ZIP writer
	if err := zipWriter.Close(); err != nil {
		errorLog.Printf("Failed to close archive: %v", err)
		os.Exit(ExitCollectError)
	}

	// Print professional summary
	printSummary(totalCollected, outputFile, cfg)

	if totalCollected == 0 {
		errorLog.Println("No data collected - this may indicate a problem")
		os.Exit(ExitNoData)
	}
}

func parseConfig() (*Config, error) {
	cfg := &Config{}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: radar [options]\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&cfg.Host, "h", "", "database host")
	flag.IntVar(&cfg.Port, "p", DefaultPostgresPort, "database port")
	flag.StringVar(&cfg.Database, "d", "", "database name")
	flag.StringVar(&cfg.Username, "U", "", "database user")
	flag.StringVar(&cfg.DataDir, "data-dir", "", "PostgreSQL data directory")
	flag.BoolVar(&cfg.SkipSystem, "skip-system", false, "skip system data collection")
	flag.BoolVar(&cfg.SkipPostgres, "skip-postgres", false, "skip PostgreSQL data collection")
	flag.BoolVar(&cfg.Verbose, "v", false, "verbose output (summary)")
	flag.BoolVar(&cfg.VeryVerbose, "vv", false, "very verbose output (detailed)")
	flag.Parse()

	// If -vv is set, also enable -v
	if cfg.VeryVerbose {
		cfg.Verbose = true
	}

	// Environment variable fallbacks
	if cfg.Host == "" {
		cfg.Host = os.Getenv("PGHOST")
		if cfg.Host == "" {
			cfg.Host = "localhost"
		}
	}

	if cfg.Username == "" {
		cfg.Username = os.Getenv("PGUSER")
		if cfg.Username == "" {
			if u, err := user.Current(); err == nil {
				cfg.Username = u.Username
			} else {
				cfg.Username = "postgres"
			}
		}
	}

	cfg.Password = os.Getenv("PGPASSWORD")

	// Validate skip flag combinations
	if cfg.SkipSystem && cfg.SkipPostgres {
		return nil, fmt.Errorf("cannot use --skip-system and --skip-postgres together (nothing would be collected)")
	}

	// If skipping system data, PostgreSQL connection is mandatory
	if cfg.SkipSystem && cfg.Database == "" {
		return nil, fmt.Errorf("--skip-system requires PostgreSQL database (-d flag)")
	}

	// Default to "postgres" database if not specified and not skipping PostgreSQL
	if !cfg.SkipPostgres && cfg.Database == "" {
		cfg.Database = DefaultDatabase
	}

	return cfg, nil
}

func (c *Config) ConnectionString() string {
	params := []string{
		fmt.Sprintf("host=%s", c.Host),
		fmt.Sprintf("port=%d", c.Port),
		fmt.Sprintf("dbname=%s", c.Database),
		fmt.Sprintf("user=%s", c.Username),
		"sslmode=disable",
	}

	if c.Password != "" {
		params = append(params, fmt.Sprintf("password=%s", c.Password))
	}

	return strings.Join(params, " ")
}

func initPostgreSQL(cfg *Config) error {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	cfg.DB = db
	return nil
}
func collectAll(cfg *Config, zipWriter *zip.Writer) int {
	collected := 0

	// Build list of system and PostgreSQL tasks separately
	var systemTasks []CollectionTask
	var pgTasks []CollectionTask

	if !cfg.SkipSystem {
		systemTasks = getSystemTasks()
	}

	if !cfg.SkipPostgres {
		// Pass cfg.DB to PostgreSQL task generators
		pgTasks = append(pgTasks, getPostgreSQLTasks(cfg.DB)...)
		dbTasks, err := generateDatabaseTasks(cfg.DB)
		if err != nil {
			errorLog.Printf("Failed to generate database tasks: %v", err)
		} else {
			pgTasks = append(pgTasks, dbTasks...)
		}
	}

	// PHASE 1: Collect system tasks
	if len(systemTasks) > 0 {
		collected += collect(cfg, zipWriter, systemTasks)
	}

	// PHASE 2: Collect PostgreSQL tasks
	if len(pgTasks) > 0 {
		collected += collect(cfg, zipWriter, pgTasks)
	}

	return collected
}

// collect executes tasks sequentially and streams directly to ZIP
// Returns: collected count only
func collect(cfg *Config, zipWriter *zip.Writer, tasks []CollectionTask) int {
	collected := 0

	for _, task := range tasks {
		header := &zip.FileHeader{
			Name:     task.ArchivePath,
			Method:   DefaultCompressionMethod,
			Modified: time.Now(),
		}

		// Use lazy writer - only creates ZIP entry on first Write()
		lazy := &lazyZipWriter{zipWriter: zipWriter, header: header}

		err := task.Collector(cfg, lazy)
		if err != nil {
			// Silently skip - don't spam user with errors
			if cfg.VeryVerbose {
				infoLog.Printf("⊘ %s (unavailable)", task.Name)
			}
			continue
		}

		// Only count if something was actually written
		if !lazy.WroteAny() {
			if cfg.VeryVerbose {
				infoLog.Printf("⊘ %s (empty)", task.Name)
			}
			continue
		}

		if cfg.VeryVerbose {
			infoLog.Printf("✓ %s", task.Name)
		}
		collected++
	}

	return collected
}

// execCommand executes a command and returns its output
func execCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		outputStr := string(output)

		// Check for special cases first
		if special, ok := handleSpecialCases(name, output); ok {
			return special, nil
		}

		// Check if command not found
		if isCommandNotFoundError(err) {
			return nil, NewSkipError(fmt.Sprintf("command not found: %s", name))
		}

		// Check if no data available
		if isNoDataAvailable(outputStr) {
			return nil, NewSkipError(fmt.Sprintf("data not available: %s", name))
		}

		// Real error
		return nil, fmt.Errorf("command '%s %v' failed: %w (output: %s)", name, args, err, outputStr)
	}

	return output, nil
}

// readFile reads a file from the filesystem
func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewSkipError(fmt.Sprintf("file not found: %s", path))
		}
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return data, nil
}

// closeErrCheck safely closes a resource and logs any error
func closeErrCheck(closer io.Closer, resourceName string) {
	if err := closer.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to close %s: %v\n", resourceName, err)
	}
}

// execCommandCollector creates a collector that executes a command and writes output
func execCommandCollector(name string, args ...string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		data, err := execCommand(name, args...)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
}

// readFileCollector creates a collector that reads a file and writes its contents
func readFileCollector(path string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		data, err := readFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
}

// pgQueryCollector creates a collector that executes a PostgreSQL query and streams results as TSV
func pgQueryCollector(db *sql.DB, query string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		if db == nil {
			return fmt.Errorf("PostgreSQL not initialized")
		}
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		defer closeErrCheck(rows, "query rows")
		return rowsToTSV(rows, w)
	}
}

// rowsToTSV streams SQL rows to TSV format directly to writer
func rowsToTSV(rows *sql.Rows, w io.Writer) error {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("getting columns: %w", err)
	}

	// Write TSV header
	for i, col := range columns {
		if i > 0 {
			if _, err := w.Write([]byte{'\t'}); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, col); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}

	// Prepare scan destinations
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Write rows
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}

		for i, val := range values {
			if i > 0 {
				if _, err := w.Write([]byte{'\t'}); err != nil {
					return err
				}
			}

			if val == nil {
				// NULL → empty string
				continue
			}

			// Convert to string
			var str string
			switch v := val.(type) {
			case []byte:
				str = string(v)
			default:
				str = fmt.Sprintf("%v", v)
			}

			// TSV escaping: quote if contains tab, newline, or quote
			if strings.ContainsAny(str, "\t\n\r\"") {
				str = `"` + strings.ReplaceAll(str, `"`, `""`) + `"`
			}

			if _, err := io.WriteString(w, str); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating rows: %w", err)
	}

	return nil
}
