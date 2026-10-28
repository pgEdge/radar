# Reading the Archive

radar writes all data to a single ZIP file named
`radar-{hostname}-{timestamp}.zip` in the current directory. Only the owner can
read the archive, and radar refuses to overwrite an existing file. The
following tree shows the layout of the archive:

```text
radar-hostname-20260115-133700.zip
├── radar.out            (radar version and commit)
├── system/              (Linux system data)
│   ├── lsblk.out
│   ├── mount.out
│   ├── diskspace.out
│   ├── proc/
│   │   ├── meminfo.out
│   │   └── cpuinfo.out
│   └── ...
├── postgresql/          (PostgreSQL instance data)
│   ├── version.tsv
│   ├── databases.tsv
│   ├── configuration.tsv
│   ├── postgresql.conf
│   ├── pg_hba.conf
│   └── ...
├── databases/           (Per-database data)
│   ├── postgres/
│   │   ├── extensions.tsv
│   │   ├── tables.tsv
│   │   └── ...
│   └── mydb/
│       ├── extensions.tsv
│       └── ...
├── pg_statviz/          (Per-database pg_statviz time series, if installed)
│   └── mydb/
│       ├── snapshots.tsv
│       └── ...
├── spock/               (Per-database Spock replication state, if installed)
│   └── mydb/
│       ├── lag_tracker.tsv
│       └── ...
└── pgbouncer/           (PgBouncer configuration, if installed)
    ├── pgbouncer.ini
    └── files.tsv
```

The following table describes the file formats in the archive:

| Extension | Format |
|-----------|--------|
| `.tsv` | Tab-separated values with a header row; query results and directory listings use this format. |
| `.out` | Command output, with `stdout` and `stderr` combined, or a copy of a system file; `radar.out` records the radar version and commit. |
| `.conf` | The raw contents of a configuration file, such as `postgresql.conf`. |
| `.ini` | The `pgbouncer.ini` file, with the values outside the `[pgbouncer]` section redacted. |
| `.plist` | A macOS property list, such as the system version. |
