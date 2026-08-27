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
			Collector: func(cfg *Config, w io.Writer) error {
				return collectPgBouncerINI(w)
			},
		},
		{
			Category:    "system",
			Name:        "pgbouncer_files",
			ArchivePath: "pgbouncer/files.tsv",
			Collector: func(cfg *Config, w io.Writer) error {
				return collectPgBouncerFiles(w)
			},
		},
	}
}

// findPgBouncerDir returns the first candidate configuration directory that
// exists. Absence means PgBouncer is not installed, which is most hosts, and is
// a skip rather than a failure.
func findPgBouncerDir() (string, error) {
	for _, dir := range pgBouncerConfigDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", NewSkipError("PgBouncer is not installed")
}

// collectPgBouncerINI writes the PgBouncer configuration with every value
// outside the [pgbouncer] section removed.
func collectPgBouncerINI(w io.Writer) error {
	dir, err := findPgBouncerDir()
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
// A commented-out setting is redacted wherever it appears, [pgbouncer] included,
// because a commented-out connection string carries a password as readily as a
// live one. A prose comment has no value to redact and survives, as does an
// %include line, which names a file radar does not follow.
func filterPgBouncerINI(data []byte, w io.Writer) error {
	shipValues := false

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "["):
			name, _, closed := strings.Cut(strings.TrimPrefix(trimmed, "["), "]")
			shipValues = closed && strings.EqualFold(strings.TrimSpace(name), "pgbouncer")
		case !shipValues || isINIComment(trimmed):
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
func collectPgBouncerFiles(w io.Writer) error {
	dir, err := findPgBouncerDir()
	if err != nil {
		return err
	}

	return writeDirListingTSV(dir, w)
}
