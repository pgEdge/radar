# Reviewing the Command-Line Options

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
