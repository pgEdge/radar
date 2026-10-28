# Running a Collection

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
