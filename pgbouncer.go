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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// pgBouncerConfigDirs are the usual locations of a PgBouncer configuration,
// tried in order.
var pgBouncerConfigDirs = []string{
	"/etc/pgbouncer",
	"/usr/local/etc/pgbouncer",
}

// redactedValue stands in for every configuration value radar declines to ship.
const redactedValue = "<redacted by radar>"

// getPgBouncerTasks returns the PgBouncer collection tasks. They sit with the
// system tasks rather than the PostgreSQL ones because they need no database
// connection: whether a pooler is deployed in front of the instance is worth
// knowing even when the instance itself is unreachable.
func getPgBouncerTasks() []CollectionTask {
	return []CollectionTask{
		{
			Category:    "system",
			Name:        "pgbouncer.ini",
			ArchivePath: "pgbouncer/pgbouncer.ini",
			Collector:   collectPgBouncerINI,
		},
		{
			Category:    "system",
			Name:        "pgbouncer_files",
			ArchivePath: "pgbouncer/files.tsv",
			Collector:   collectPgBouncerFiles,
		},
	}
}

// findPgBouncerDir returns the configuration directory named by -pgbouncer-conf,
// or the first of the usual locations that exists. A named directory is used as
// given, and a wrong one is a failure rather than a skip, because the operator
// said where to look. Absence from the usual locations means PgBouncer is not
// installed, which is most hosts, and is a skip.
func findPgBouncerDir(cfg *Config) (string, error) {
	dirs := pgBouncerConfigDirs
	if cfg.PgBouncerConf != "" {
		dirs = []string{cfg.PgBouncerConf}
	}

	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	if cfg.PgBouncerConf != "" {
		return "", fmt.Errorf("no PgBouncer configuration directory at %s", cfg.PgBouncerConf)
	}
	return "", NewSkipError(fmt.Sprintf("no PgBouncer configuration directory in %v", dirs))
}

// collectPgBouncerINI writes the PgBouncer configuration with every value
// outside the [pgbouncer] section removed.
func collectPgBouncerINI(cfg *Config, w io.Writer) error {
	dir, err := findPgBouncerDir(cfg)
	if err != nil {
		return err
	}

	data, err := readFile(filepath.Join(dir, "pgbouncer.ini"))
	if err != nil {
		return err
	}

	return filterPgBouncerINI(data, w)
}

// filterPgBouncerINI ships the [pgbouncer] section verbatim and keeps only the
// key names everywhere else. [databases] and [peers] hold connection strings
// that routinely contain password=, so their keys alone go into the archive:
// enough to say which databases are pooled, without the credentials to reach
// them. Every other section is treated the same way, so a section radar has not
// vetted is not shipped verbatim just because it is new.
//
// Comments are dropped entirely, wherever they appear. A commented-out
// connection string carries a password as readily as a live one, and no comment
// is worth the trouble of deciding which ones are safe. An %include line is not
// a comment and is kept, naming a file radar does not follow.
func filterPgBouncerINI(data []byte, w io.Writer) error {
	shipValues := false

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case isINIComment(trimmed):
			continue
		case strings.HasPrefix(trimmed, "["):
			name, _, closed := strings.Cut(strings.TrimPrefix(trimmed, "["), "]")
			shipValues = closed && strings.EqualFold(strings.TrimSpace(name), "pgbouncer")
		case !shipValues:
			if key, _, found := strings.Cut(line, "="); found {
				line = key + "= " + redactedValue
			}
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// isINIComment reports whether a trimmed line is a PgBouncer configuration
// comment, which starts with either ; or #.
func isINIComment(trimmed string) bool {
	return strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#")
}

// collectPgBouncerFiles lists the PgBouncer configuration directory. This is how
// userlist.txt is recorded: that it is there, how large it is and when it last
// changed, never its contents, which are credentials.
func collectPgBouncerFiles(cfg *Config, w io.Writer) error {
	dir, err := findPgBouncerDir(cfg)
	if err != nil {
		return err
	}

	return writeDirListingTSV(dir, w)
}
