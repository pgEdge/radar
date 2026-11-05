# Collecting PgBouncer Configuration

radar collects the [PgBouncer](https://www.pgbouncer.org/) configuration from
the first of `/etc/pgbouncer/` and `/usr/local/etc/pgbouncer/` that exists.
radar does not fall back to the second directory when the first directory lacks
a file. When neither directory exists, radar treats PgBouncer as not installed
and skips these collectors. radar uses a directory named with the
`--pgbouncer-conf` option as given, and treats a wrong path as a failure rather
than a skip. These collectors need no database connection, so radar collects
the files even when the PostgreSQL instance is unreachable.

radar never collects the contents of `userlist.txt`. The file holds
credentials, so `files.tsv` records only that `userlist.txt` is present as a
regular file.

The following table describes the PgBouncer files:

| File | Source | Description |
|------|--------|-------------|
| `pgbouncer/pgbouncer.ini` | `pgbouncer.ini` | Contains the `[pgbouncer]` section without its comments, including the pool mode, connection limits, listen address, auth type and auth file. In every other section, radar keeps each key name and replaces its value with `<redacted by radar>`, because `[databases]` and `[peers]` connection strings routinely contain `password=`. radar drops comments entirely, wherever they appear, because a commented-out entry holds a password as readily as a live entry. radar keeps an `%include` line but does not follow the file that the line names. radar reads the file from the `--pgbouncer-conf` directory when given, and otherwise from the first of `/etc/pgbouncer/` and `/usr/local/etc/pgbouncer/` that exists, without falling back to the second. |
| `pgbouncer/files.tsv` | filesystem | Lists the regular files in the configuration directory with the name, size and modification time of each file. This listing is how radar records `userlist.txt`. radar leaves out symbolic links and subdirectories, so the listing does not record a `userlist.txt` that is a symbolic link. |
