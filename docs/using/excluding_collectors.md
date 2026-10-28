# Excluding Collectors

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
