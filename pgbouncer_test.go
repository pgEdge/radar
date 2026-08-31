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
	"path/filepath"
	"strings"
	"testing"
)

// sampleINI mirrors a real pgbouncer.ini: credentials in the [databases]
// connection strings, and a commented-out one carrying a password too.
const sampleINI = `; PgBouncer configuration
%include /etc/pgbouncer/extra.ini

[databases]
mydb = host=db1.internal port=5432 dbname=mydb user=pooler password=s3cr3t
* = host=db2.internal password=another-s3cr3t
; olddb = host=db3.internal password=retired-s3cr3t

[peers]
1 = host=peer1.internal port=6432

[users]
pooler = pool_mode=transaction

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 500
default_pool_size = 20
`

// TestFilterPgBouncerINI verifies the [pgbouncer] section is shipped intact
// while every other section keeps its keys and loses its values, so an archive
// says which databases are pooled without carrying the credentials to reach
// them.
func TestFilterPgBouncerINI(t *testing.T) {
	var buf bytes.Buffer
	if err := filterPgBouncerINI([]byte(sampleINI), &buf); err != nil {
		t.Fatalf("filterPgBouncerINI: %v", err)
	}
	got := buf.String()

	// No secret from any section may survive, including the commented-out one.
	assertAbsent(t, got, []string{"s3cr3t", "another-s3cr3t", "retired-s3cr3t", "password="})

	// The [pgbouncer] section is what answers whether a pooler is deployed and
	// how it is tuned, so its values must survive verbatim. Keys elsewhere
	// survive too, along with every section header, so the reader can still see
	// the shape of the configuration. An %include names a file radar does not
	// follow, and keeping the line says more configuration exists.
	assertPresent(t, got, []string{
		"[pgbouncer]",
		"listen_addr = 127.0.0.1",
		"listen_port = 6432",
		"auth_type = scram-sha-256",
		"auth_file = /etc/pgbouncer/userlist.txt",
		"pool_mode = transaction",
		"max_client_conn = 500",
		"default_pool_size = 20",
		"[databases]", "[peers]", "[users]", "mydb", "pooler",
		"%include /etc/pgbouncer/extra.ini",
	})

	// No comment survives, so neither the prose header nor the commented-out
	// [databases] entry appears at all.
	assertAbsent(t, got, []string{"; PgBouncer configuration", "olddb"})

	// pool_mode appears in both [users] and [pgbouncer]. The [users] one must
	// be redacted, so exactly one unredacted occurrence may remain.
	if n := strings.Count(got, "pool_mode = transaction"); n != 1 {
		t.Errorf("got %d unredacted pool_mode values, want 1:\n%s", n, got)
	}
}

// assertAbsent fails the test for every item present in got.
func assertAbsent(t *testing.T, got string, items []string) {
	t.Helper()
	for _, item := range items {
		if strings.Contains(got, item) {
			t.Errorf("filtered output contains %q:\n%s", item, got)
		}
	}
}

// assertPresent fails the test for every item missing from got.
func assertPresent(t *testing.T, got string, items []string) {
	t.Helper()
	for _, item := range items {
		if !strings.Contains(got, item) {
			t.Errorf("filtered output lost %q:\n%s", item, got)
		}
	}
}

// TestFilterPgBouncerINIRedactsUnknownSections verifies the filter fails closed:
// a section radar has not vetted is not shipped verbatim just because it is new.
func TestFilterPgBouncerINIRedactsUnknownSections(t *testing.T) {
	var buf bytes.Buffer
	ini := "[something_new]\nsecret_key = do-not-ship\n"
	if err := filterPgBouncerINI([]byte(ini), &buf); err != nil {
		t.Fatalf("filterPgBouncerINI: %v", err)
	}

	if strings.Contains(buf.String(), "do-not-ship") {
		t.Errorf("unvetted section was shipped verbatim:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "secret_key") {
		t.Errorf("unvetted section lost its key names:\n%s", buf.String())
	}
}

// TestFindPgBouncerDirSkipsWhenAbsent verifies that a host without PgBouncer,
// which is most hosts, is a skip rather than a failure.
func TestFindPgBouncerDirSkipsWhenAbsent(t *testing.T) {
	original := pgBouncerConfigDirs
	t.Cleanup(func() { pgBouncerConfigDirs = original })
	pgBouncerConfigDirs = []string{filepath.Join(t.TempDir(), "no-pgbouncer-here")}

	_, err := findPgBouncerDir(&Config{})

	var skipErr SkipError
	if !errors.As(err, &skipErr) {
		t.Fatalf("err = %v, want SkipError", err)
	}
}

// TestFindPgBouncerDirPrefersFirstPresent verifies the candidate locations are
// tried in order.
func TestFindPgBouncerDirPrefersFirstPresent(t *testing.T) {
	dir := t.TempDir()
	original := pgBouncerConfigDirs
	t.Cleanup(func() { pgBouncerConfigDirs = original })
	pgBouncerConfigDirs = []string{filepath.Join(dir, "absent"), dir}

	got, err := findPgBouncerDir(&Config{})
	if err != nil {
		t.Fatalf("findPgBouncerDir: %v", err)
	}
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
	}
}

// TestFindPgBouncerDirHonoursFlag verifies that -pgbouncer-conf replaces the
// usual locations rather than adding to them, so a named path that is absent
// reports itself instead of silently falling back to /etc.
func TestFindPgBouncerDirHonoursFlag(t *testing.T) {
	present := t.TempDir()
	original := pgBouncerConfigDirs
	t.Cleanup(func() { pgBouncerConfigDirs = original })
	pgBouncerConfigDirs = []string{present}

	named := t.TempDir()
	got, err := findPgBouncerDir(&Config{PgBouncerConf: named})
	if err != nil {
		t.Fatalf("findPgBouncerDir with a named directory: %v", err)
	}
	if got != named {
		t.Errorf("got %q, want the named directory %q", got, named)
	}

	// A named directory that does not exist is a failure, not a skip: the
	// operator said where to look, so falling quiet would hide the typo.
	_, err = findPgBouncerDir(&Config{PgBouncerConf: filepath.Join(named, "absent")})
	var skipErr SkipError
	if err == nil || errors.As(err, &skipErr) {
		t.Fatalf("err = %v, want a plain error for an absent named directory", err)
	}
}

// TestPgBouncerCollectorsRegistered verifies the PgBouncer tasks are wired into
// the system task set, which is collected even when the database is unreachable.
func TestPgBouncerCollectorsRegistered(t *testing.T) {
	expected := map[string]string{
		"pgbouncer.ini":   "pgbouncer/pgbouncer.ini",
		"pgbouncer_files": "pgbouncer/files.tsv",
	}

	got := make(map[string]string)
	for _, task := range getSystemTasks() {
		if _, wanted := expected[task.Name]; wanted {
			got[task.Name] = task.ArchivePath
			if task.Collector == nil {
				t.Errorf("task %q has a nil collector", task.Name)
			}
		}
	}

	for name, path := range expected {
		if got[name] != path {
			t.Errorf("task %q archive path = %q, want %q", name, got[name], path)
		}
	}
}

// TestFilterPgBouncerINIDropsAllComments verifies that no comment reaches the
// archive, wherever it appears. A commented-out connection string carries a
// password as readily as a live one, so none is kept.
func TestFilterPgBouncerINIDropsAllComments(t *testing.T) {
	ini := `; PgBouncer configuration
[pgbouncer]
listen_port = 6432
; password=comment-secret-value
# olddb = host=db9.internal password=hash-secret-value
`
	var buf bytes.Buffer
	if err := filterPgBouncerINI([]byte(ini), &buf); err != nil {
		t.Fatalf("filterPgBouncerINI: %v", err)
	}
	got := buf.String()

	// Every comment is gone, secret or not, so nothing beginning ; or # remains.
	assertAbsent(t, got, []string{"comment-secret-value", "hash-secret-value", ";", "#"})
	assertPresent(t, got, []string{"listen_port = 6432"})
}
