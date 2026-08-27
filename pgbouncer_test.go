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
	for _, secret := range []string{"s3cr3t", "another-s3cr3t", "retired-s3cr3t", "password="} {
		if strings.Contains(got, secret) {
			t.Errorf("filtered output contains %q:\n%s", secret, got)
		}
	}

	// The [pgbouncer] section is what answers whether a pooler is deployed and
	// how it is tuned, so its values must survive verbatim.
	for _, want := range []string{
		"[pgbouncer]",
		"listen_addr = 127.0.0.1",
		"listen_port = 6432",
		"auth_type = scram-sha-256",
		"auth_file = /etc/pgbouncer/userlist.txt",
		"pool_mode = transaction",
		"max_client_conn = 500",
		"default_pool_size = 20",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("filtered output lost %q:\n%s", want, got)
		}
	}

	// Keys elsewhere survive so the reader can still see the shape of the
	// configuration, and every section header is kept.
	for _, want := range []string{"[databases]", "[peers]", "[users]", "mydb", "pooler"} {
		if !strings.Contains(got, want) {
			t.Errorf("filtered output lost %q:\n%s", want, got)
		}
	}

	// An %include names a file radar does not follow; keeping the line tells
	// the reader that more configuration exists.
	if !strings.Contains(got, "%include /etc/pgbouncer/extra.ini") {
		t.Errorf("filtered output lost the %%include line:\n%s", got)
	}

	// pool_mode appears in both [users] and [pgbouncer]. The [users] one must
	// be redacted, so exactly one unredacted occurrence may remain.
	if n := strings.Count(got, "pool_mode = transaction"); n != 1 {
		t.Errorf("got %d unredacted pool_mode values, want 1:\n%s", n, got)
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

	_, err := findPgBouncerDir()

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

	got, err := findPgBouncerDir()
	if err != nil {
		t.Fatalf("findPgBouncerDir: %v", err)
	}
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
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
