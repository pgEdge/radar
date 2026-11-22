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
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// version is stamped at build time via -ldflags "-X main.version=<version>".
// Release builds set it to the tag without its leading "v", e.g. "0.5.0";
// unstamped dev builds report "dev".
var version = "dev"

// errHelpRequested and errVersionRequested are returned by parseConfig when
// -help or -version was given instead of a collection request. main prints the
// corresponding output and exits successfully.
var (
	errHelpRequested    = errors.New("help requested")
	errVersionRequested = errors.New("version requested")
)

// errCommandLine is returned by parseConfig when the flag package rejects the
// command line. The flag package has by then printed the reason and the usage.
var errCommandLine = errors.New("the command line cannot be parsed")

// printUsage writes the help text: the usage line, the global options, then
// the collection options from the registered flag set. It carries nothing
// build-dependent, because it is pasted verbatim into README.md and
// docs/index.md. The write error is discarded because flag.PrintDefaults,
// which emits the rest of the same output, discards its own.
func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `Usage: radar [options]

GLOBAL OPTIONS:
   --help         show help
   --version, -V  print the version

Options:
`)
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

// defaultDisabledTasks lists task names not run unless -include lists them.
// pgstattuple_approx() reads heap pages of every user table.
var defaultDisabledTasks = []string{"pgstattuple"}

// shouldRunTask applies -include / -exclude / default-disabled to a task name.
// -include wins over -exclude wins over the default-disabled set.
// Per-database tasks are named "dbname/taskname"; the trailing segment is
// what users type on the command line.
func shouldRunTask(name string, cfg *Config) bool {
	short := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		short = name[i+1:]
	}
	if slices.Contains(cfg.Include, short) {
		return true
	}
	if slices.Contains(cfg.Exclude, short) {
		return false
	}
	return !slices.Contains(defaultDisabledTasks, short)
}

// writeRadarMeta writes a radar.out entry at the archive root identifying the
// radar binary that produced the archive. Version comes from the build-time
// -X main.version stamp; commit comes from Go's embedded VCS info
// (debug.ReadBuildInfo).
func writeRadarMeta(zw *zip.Writer) error {
	commit := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) >= 7 {
					commit = s.Value[:7]
				} else if s.Value != "" {
					commit = s.Value
				}
				break
			}
		}
	}

	w, err := zw.CreateHeader(archiveHeader("radar.out"))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "version: %s\ncommit: %s\n", version, commit)
	return err
}

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
	NoDataPatterns             = []string{"No such file or directory", "no match", "doesn't exist", "not been booted with systemd"}
)

// Exit codes
const (
	// ExitError is a failure that none of the other codes describes: the
	// archive cannot be created or closed.
	ExitError = 1
	// ExitInvalidArg is an argument that is invalid or missing.
	ExitInvalidArg = 3
	// ExitNoData is a run that collected nothing.
	ExitNoData = 4
	// ExitUsage is a command line that cannot be parsed.
	ExitUsage = 64
	// ExitUnavailable is a PostgreSQL instance that cannot be reached.
	ExitUnavailable = 69
)

// Config holds connection parameters and collection settings
type Config struct {
	// PostgreSQL connection
	Host          string
	Port          int
	Database      string
	Username      string
	Password      string
	DataDir       string
	PgBouncerConf string
	SSLMode       string
	SSLCert       string
	SSLKey        string
	SSLRootCert   string

	// Database connection (injected)
	DB *sql.DB

	// Collection control
	SkipSystem   bool
	SkipPostgres bool
	Exclude      []string
	Include      []string
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

func archiveHeader(name string) *zip.FileHeader {
	return &zip.FileHeader{
		Name:     name,
		Method:   DefaultCompressionMethod,
		Modified: time.Now(),
	}
}

// lazyZipWriter defers ZIP entry creation until first Write()
// This prevents empty files in the archive when collectors produce no output
type lazyZipWriter struct {
	zipWriter *zip.Writer
	header    *zip.FileHeader
	writer    io.Writer // nil until first Write()
}

// Write defers ZIP entry creation until first non-empty write.
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

// WroteAny returns true if any data was written.
func (w *lazyZipWriter) WroteAny() bool {
	return w.writer != nil
}

// SkipError indicates a collector was skipped (tool missing, no data, not applicable)
type SkipError struct {
	Reason string
}

// Error returns the skip reason.
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

// main is the radar entry point.
func main() {
	os.Exit(run())
}

// run collects one archive and returns radar's exit status. It returns rather
// than exits, so that the deferred closes run on every path.
func run() int {
	cfg, err := parseConfig()
	switch {
	case errors.Is(err, errHelpRequested):
		printUsage(os.Stdout)
		return 0
	case errors.Is(err, errVersionRequested):
		fmt.Printf("radar version %s\n", version)
		return 0
	case errors.Is(err, errCommandLine):
		return ExitUsage
	case err != nil:
		errorLog.Println(err)
		flag.Usage()
		return ExitInvalidArg
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
	pgUnavailable := false
	if !cfg.SkipPostgres {
		if cfg.Verbose {
			infoLog.Printf("Connecting to PostgreSQL at %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
		}
		if err := initPostgreSQL(cfg); err != nil {
			errorLog.Printf("Could not connect to PostgreSQL: %v", err)
			if cfg.SkipSystem {
				errorLog.Println("Nothing to collect: --skip-system is set and PostgreSQL is unreachable")
				return ExitUnavailable
			}
			errorLog.Println("Continuing with system data collection only...")
			cfg.SkipPostgres = true
			pgUnavailable = true
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
	outFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		errorLog.Printf("Failed to create output file: %v", err)
		return ExitError
	}
	defer closeErrCheck(outFile, "output file")

	// Create ZIP writer
	zipWriter := zip.NewWriter(outFile)

	// Write radar build identity (version, commit) to the archive root
	if err := writeRadarMeta(zipWriter); err != nil {
		errorLog.Printf("Failed to write radar.out: %v", err)
	}

	// Collect all data
	if cfg.Verbose {
		infoLog.Println("Starting data collection...")
	}
	totalCollected := collectAll(cfg, zipWriter)

	// Close ZIP writer
	if err := zipWriter.Close(); err != nil {
		errorLog.Printf("Failed to close archive: %v", err)
		return ExitError
	}

	if totalCollected == 0 {
		errorLog.Println("No data collected - this may indicate a problem")
		return ExitNoData
	}

	// Print summary
	printSummary(totalCollected, outputFile, cfg)
	if pgUnavailable {
		return ExitUnavailable
	}
	return 0
}

// parseConfig parses command-line flags into a Config.
func parseConfig() (*Config, error) {
	cfg := &Config{}

	flag.Usage = func() { printUsage(os.Stderr) }

	flag.StringVar(&cfg.Host, "h", "", "database host")
	flag.IntVar(&cfg.Port, "p", DefaultPostgresPort, "database port")
	flag.StringVar(&cfg.Database, "d", "", "database name")
	flag.StringVar(&cfg.Username, "U", "", "database user")
	flag.StringVar(&cfg.DataDir, "data-dir", "", "PostgreSQL data directory")
	flag.StringVar(&cfg.PgBouncerConf, "pgbouncer-conf", "", "PgBouncer configuration directory (default: /etc/pgbouncer, /usr/local/etc/pgbouncer)")
	flag.StringVar(&cfg.SSLMode, "sslmode", "prefer", "SSL mode (prefer, disable, require, verify-ca, verify-full)")
	flag.StringVar(&cfg.SSLCert, "sslcert", "", "client SSL certificate file")
	flag.StringVar(&cfg.SSLKey, "sslkey", "", "client SSL key file")
	flag.StringVar(&cfg.SSLRootCert, "sslrootcert", "", "SSL root certificate file")
	flag.BoolVar(&cfg.SkipSystem, "skip-system", false, "skip system data collection")
	flag.BoolVar(&cfg.SkipPostgres, "skip-postgres", false, "skip PostgreSQL data collection")
	var excludeRaw, includeRaw string
	flag.StringVar(&excludeRaw, "exclude", "", "comma-separated task names to skip (e.g. -exclude bloat,stat_ssl)")
	flag.StringVar(&includeRaw, "include", "", "comma-separated default-disabled task names to enable (e.g. pgstattuple, disabled by default)")
	flag.BoolVar(&cfg.Verbose, "v", false, "verbose output (summary)")
	flag.BoolVar(&cfg.VeryVerbose, "vv", false, "very verbose output (detailed)")

	// Read after registration so printUsage can list the options above, and
	// before parsing so that neither a collection flag nor a typo elsewhere
	// stops radar answering. Help outranks version when both are given.
	switch args := os.Args[1:]; {
	case slices.Contains(args, "-help"), slices.Contains(args, "--help"):
		return nil, errHelpRequested
	case slices.Contains(args, "-version"), slices.Contains(args, "--version"), slices.Contains(args, "-V"):
		return nil, errVersionRequested
	}

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, errCommandLine
	}

	for _, raw := range strings.Split(excludeRaw, ",") {
		if t := strings.TrimSpace(raw); t != "" {
			cfg.Exclude = append(cfg.Exclude, t)
		}
	}
	for _, raw := range strings.Split(includeRaw, ",") {
		if t := strings.TrimSpace(raw); t != "" {
			cfg.Include = append(cfg.Include, t)
		}
	}

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

	portFlagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "p" {
			portFlagSet = true
		}
	})
	if !portFlagSet && !cfg.SkipPostgres {
		if pgport := os.Getenv("PGPORT"); pgport != "" {
			p, err := strconv.Atoi(pgport)
			if err != nil {
				return nil, fmt.Errorf("invalid PGPORT %q: must be a port number", pgport)
			}
			cfg.Port = p
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

	if cfg.Database == "" {
		cfg.Database = os.Getenv("PGDATABASE")
	}

	if !cfg.SkipPostgres {
		sslmodeFlagSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "sslmode" {
				sslmodeFlagSet = true
			}
		})
		if !sslmodeFlagSet {
			if v := os.Getenv("PGSSLMODE"); v != "" {
				cfg.SSLMode = v
			}
		}

		validSSLModes := map[string]bool{
			"prefer": true, "disable": true, "require": true,
			"verify-ca": true, "verify-full": true,
		}
		if !validSSLModes[cfg.SSLMode] {
			return nil, fmt.Errorf("invalid sslmode %q: must be one of prefer, disable, require, verify-ca, verify-full", cfg.SSLMode)
		}

		if cfg.SSLCert == "" {
			cfg.SSLCert = os.Getenv("PGSSLCERT")
		}
		if cfg.SSLKey == "" {
			cfg.SSLKey = os.Getenv("PGSSLKEY")
		}
		if cfg.SSLRootCert == "" {
			cfg.SSLRootCert = os.Getenv("PGSSLROOTCERT")
		}

		// Cert and key must be specified together
		if (cfg.SSLCert == "") != (cfg.SSLKey == "") {
			return nil, fmt.Errorf("--sslcert and --sslkey must be specified together")
		}
		// verify-ca and verify-full require a root certificate
		if (cfg.SSLMode == "verify-ca" || cfg.SSLMode == "verify-full") && cfg.SSLRootCert == "" {
			return nil, fmt.Errorf("--sslrootcert is required for sslmode=%s", cfg.SSLMode)
		}
		// Validate cert files exist
		for _, f := range []struct{ flag, path string }{
			{"--sslcert", cfg.SSLCert}, {"--sslkey", cfg.SSLKey}, {"--sslrootcert", cfg.SSLRootCert},
		} {
			if f.path != "" {
				if _, err := os.Stat(f.path); err != nil {
					return nil, fmt.Errorf("%s: %w", f.flag, err)
				}
			}
		}
	}

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

// ConnectionString builds a PostgreSQL connection string for the given database.
// Values containing spaces, quotes, or backslashes are escaped per libpq conventions.
func (c *Config) ConnectionString(dbname string) string {
	q := func(v string) string {
		if !strings.ContainsAny(v, " '\\") {
			return v
		}
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `'`, `\'`)
		return "'" + v + "'"
	}

	params := []string{
		"host=" + q(c.Host),
		fmt.Sprintf("port=%d", c.Port),
		"dbname=" + q(dbname),
		"user=" + q(c.Username),
		"sslmode=" + q(c.SSLMode),
	}

	// pgx applies PGCONNECT_TIMEOUT itself, and a value given here would
	// override it.
	if os.Getenv("PGCONNECT_TIMEOUT") == "" {
		params = append(params, fmt.Sprintf("connect_timeout=%d", connectTimeoutSeconds))
	}
	if c.Password != "" {
		params = append(params, "password="+q(c.Password))
	}
	if c.SSLCert != "" {
		params = append(params, "sslcert="+q(c.SSLCert))
	}
	if c.SSLKey != "" {
		params = append(params, "sslkey="+q(c.SSLKey))
	}
	if c.SSLRootCert != "" {
		params = append(params, "sslrootcert="+q(c.SSLRootCert))
	}

	return strings.Join(params, " ")
}

// initPostgreSQL opens and verifies the PostgreSQL connection.
func initPostgreSQL(cfg *Config) error {
	db, err := sql.Open("pgx", cfg.ConnectionString(cfg.Database))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		closeErrCheck(db, "database connection")
		return err
	}

	cfg.DB = db
	return nil
}

// collectAll runs all collection tasks and writes results to the ZIP archive.
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
		dbTasks, conns, err := generateDatabaseTasks(cfg.DB)
		if err != nil {
			errorLog.Printf("Failed to generate database tasks: %v", err)
		} else {
			defer closeErrCheck(conns, "database connection")
			pgTasks = append(pgTasks, dbTasks...)
		}
	}

	// Filter both phases up-front; log the skipped set once, in verbose mode.
	systemTasks, systemSkipped := filterTasks(systemTasks, cfg)
	pgTasks, pgSkipped := filterTasks(pgTasks, cfg)
	skipped := append(systemSkipped, pgSkipped...)
	if len(skipped) > 0 && cfg.Verbose {
		infoLog.Printf("Skipping %d task(s): %s", len(skipped), strings.Join(skipped, ", "))
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

// filterTasks partitions tasks into the runnable set and the names of skipped
// tasks based on -include / -exclude / default-disabled rules.
func filterTasks(tasks []CollectionTask, cfg *Config) (runnable []CollectionTask, skipped []string) {
	for _, t := range tasks {
		if shouldRunTask(t.Name, cfg) {
			runnable = append(runnable, t)
		} else {
			skipped = append(skipped, t.Name)
		}
	}
	return
}

// collect executes tasks sequentially and streams directly to ZIP
// Returns: collected count only
func collect(cfg *Config, zipWriter *zip.Writer, tasks []CollectionTask) int {
	collected := 0

	for _, task := range tasks {
		// Use lazy writer - only creates ZIP entry on first Write()
		lazy := &lazyZipWriter{zipWriter: zipWriter, header: archiveHeader(task.ArchivePath)}

		err := task.Collector(cfg, lazy)
		if err != nil {
			var skipErr SkipError
			if errors.As(err, &skipErr) {
				// Unavailable (command not found, file missing, no data)
				if cfg.VeryVerbose {
					infoLog.Printf("⊘ %s (unavailable)", task.Name)
				}
			} else {
				// Error (I/O, permission, SQL)
				errorLog.Printf("✗ %s: %v", task.Name, err)
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

// commandTimeout is how long a command may run before radar stops it, so that
// one hung command cannot stall the run.
var commandTimeout = 30 * time.Second

// queryTimeout is how long a query may run, rows included, before radar
// cancels it, so that one query waiting on a lock cannot stall the run.
var queryTimeout = 30 * time.Second

// connectTimeoutSeconds limits each PostgreSQL connection attempt when the
// operator has not set PGCONNECT_TIMEOUT.
var connectTimeoutSeconds = 10

// commandHeadSize is how much of a command's output is held back from the
// archive until the command either succeeds or writes more. A command that
// fails within it is classified by its error message and leaves no entry.
const commandHeadSize = 4096

// headWriter holds back the first commandHeadSize bytes written to it, then
// streams everything to w.
type headWriter struct {
	w         io.Writer
	head      []byte
	streaming bool
}

// Write holds p back while the output still fits in the head, and otherwise
// streams the head and p to the underlying writer.
func (h *headWriter) Write(p []byte) (int, error) {
	if !h.streaming && len(h.head)+len(p) <= commandHeadSize {
		h.head = append(h.head, p...)
		return len(p), nil
	}
	if err := h.flush(); err != nil {
		return 0, err
	}
	return h.w.Write(p)
}

// flush writes the held-back head to the underlying writer, once.
func (h *headWriter) flush() error {
	if h.streaming {
		return nil
	}
	h.streaming = true
	_, err := h.w.Write(h.head)
	return err
}

// execCommand streams a command's stdout and stderr into w. Both go through
// the same writer, so os/exec gives the command a single pipe for the two and
// they arrive in the order the command wrote them.
func execCommand(w io.Writer, name string, args ...string) error {
	out := &headWriter{w: w, head: make([]byte, 0, commandHeadSize)}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	// A stopped command can be stuck in the kernel, or leave a child holding
	// the output pipe open: WaitDelay makes Run return regardless.
	cmd.WaitDelay = time.Second
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err == nil {
		return out.flush()
	}

	// What the command wrote before the deadline is data, so it is kept.
	if ctx.Err() != nil {
		if err := out.flush(); err != nil {
			return err
		}
		return fmt.Errorf("command '%s %v' timed out after %s", name, args, commandTimeout)
	}

	// The output is in the archive already, so only the failure is left to report.
	if out.streaming {
		return fmt.Errorf("command '%s %v' failed: %w (output kept in archive)", name, args, err)
	}

	if special, ok := handleSpecialCases(name, out.head); ok {
		_, err := w.Write(special)
		return err
	}

	if isCommandNotFoundError(err) {
		return NewSkipError(fmt.Sprintf("command not found: %s", name))
	}

	outputStr := string(out.head)
	if isNoDataAvailable(outputStr) {
		return NewSkipError(fmt.Sprintf("data not available: %s", name))
	}

	return fmt.Errorf("command '%s %v' failed: %w (output: %s)", name, args, err, outputStr)
}

// openFile opens path for reading. A missing file is a skip.
func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, NewSkipError(fmt.Sprintf("file not found: %s", path))
	}
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return f, nil
}

// copyFile streams the file at path into w.
func copyFile(path string, w io.Writer) error {
	f, err := openFile(path)
	if err != nil {
		return err
	}
	defer closeErrCheck(f, path)
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}
	return nil
}

// writeDirListingTSV writes a TSV listing of the regular files in dir: one row
// per file with its size and modification time. File contents are never read.
// Directories, symlinks and other irregular entries are left out, having no
// meaningful size to report. A missing directory is a skip.
func writeDirListingTSV(dir string, w io.Writer) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return NewSkipError(fmt.Sprintf("directory not found: %s", dir))
	}
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	if _, err := io.WriteString(w, "directory\tfilename\tsize_bytes\tmodified\n"); err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// The entry went away between the listing and the stat.
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			dir, entry.Name(), info.Size(), info.ModTime().Format(time.RFC3339)); err != nil {
			return err
		}
	}

	return nil
}

// closeErrCheck safely closes a resource and logs any error
func closeErrCheck(closer io.Closer, resourceName string) {
	if err := closer.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to close %s: %v\n", resourceName, err)
	}
}

// execCommandCollector creates a collector that streams a command's output
func execCommandCollector(name string, args ...string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		return execCommand(w, name, args...)
	}
}

// readFileCollector creates a collector that streams a file's contents
func readFileCollector(path string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		return copyFile(path, w)
	}
}

// pgQueryCollector creates a collector that executes a PostgreSQL query and streams results as TSV
func pgQueryCollector(db *sql.DB, relation, query string) func(*Config, io.Writer) error {
	return func(cfg *Config, w io.Writer) error {
		if db == nil {
			return fmt.Errorf("PostgreSQL not initialized")
		}
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()
		if err := skipIfRelationMissing(ctx, db, relation); err != nil {
			return err
		}
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			if isPGUnavailableError(err) {
				return NewSkipError(err.Error())
			}
			return timedOut(ctx, err)
		}
		defer closeErrCheck(rows, "query rows")
		return timedOut(ctx, rowsToTSV(rows, w))
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

			var field []byte
			switch v := val.(type) {
			case nil:
				// NULL → empty string
				continue
			case []byte:
				field = v
			case string:
				field = []byte(v)
			default:
				field = fmt.Append(nil, v)
			}
			if err := writeTSVField(w, field); err != nil {
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

// writeTSVField writes field, quoted when it holds a tab, newline, carriage
// return or double quote, with each double quote doubled. It writes the field
// in pieces rather than building an escaped copy of it.
func writeTSVField(w io.Writer, field []byte) error {
	if !bytes.ContainsAny(field, "\t\n\r\"") {
		_, err := w.Write(field)
		return err
	}
	if _, err := io.WriteString(w, `"`); err != nil {
		return err
	}
	for {
		i := bytes.IndexByte(field, '"')
		if i < 0 {
			break
		}
		// field[:i+1] ends with the quote, and the write after it doubles it.
		if _, err := w.Write(field[:i+1]); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `"`); err != nil {
			return err
		}
		field = field[i+1:]
	}
	if _, err := w.Write(field); err != nil {
		return err
	}
	_, err := io.WriteString(w, `"`)
	return err
}
