# Using radar

Each radar invocation runs one collection, writes a single ZIP archive, and
exits.

## Running a Collection

The following commands collect both kinds of data, system data only, and
PostgreSQL data only:

```bash
# Collect system and PostgreSQL data (the default)
radar -d mydatabase -U databaseuser

# Collect system data only
radar --skip-postgres

# Collect PostgreSQL data only
radar -d mydatabase -U databaseuser --skip-system
```

The `--skip-system` option also skips the PgBouncer collectors, which radar
treats as system collectors.

radar prints the archive name and its size in kilobytes when the collection
finishes:

```text
$ radar -d mydatabase
Collecting diagnostic data...
✓ Archive created: radar-hostname-20260115-133700.zip (1234 KB)
```

## Reviewing the Command-Line Options

The `radar --help` command prints the following usage summary:

```text
Usage: radar [options]

GLOBAL OPTIONS:
   --help         show help
   --version, -V  print the version

Options:
  -U string
    	database user
  -d string
    	database name
  -data-dir string
    	PostgreSQL data directory
  -exclude string
    	comma-separated task names to skip (e.g. -exclude bloat,stat_ssl)
  -h string
    	database host
  -include string
    	comma-separated default-disabled task names to enable (e.g. pgstattuple, disabled by default)
  -p int
    	database port (default 5432)
  -pgbouncer-conf string
    	PgBouncer configuration directory (default: /etc/pgbouncer, /usr/local/etc/pgbouncer)
  -skip-postgres
    	skip PostgreSQL data collection
  -skip-system
    	skip system data collection
  -sslcert string
    	client SSL certificate file
  -sslkey string
    	client SSL key file
  -sslmode string
    	SSL mode (prefer, disable, require, verify-ca, verify-full) (default "prefer")
  -sslrootcert string
    	SSL root certificate file
  -v	verbose output (summary)
  -vv
    	very verbose output (detailed)
```

## Excluding Collectors

The `--exclude` option skips the collectors that you name, and the `--include`
option enables the collectors that are disabled by default; `--include` takes
precedence over `--exclude`.

Run radar with `-vv` to list each collector by name; for a per-database
collector, the name after the database prefix is the name that `--exclude`
accepts:

```text
✓ stat_ssl
✓ postgres/tables
```

The following command skips the `bloat` and `stat_ssl` collectors:

```bash
radar -d mydatabase -U databaseuser --exclude bloat,stat_ssl
```
