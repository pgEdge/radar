# Collecting pg_statviz Data

The pg_statviz collectors are optional; these collectors are available when a
database has the [pg_statviz](https://github.com/vyruss/pg_statviz) extension
installed. radar stores their files in `pg_statviz/{dbname}/`. The following
table describes the pg_statviz files and the table that each file comes from:

| File | Table | Description |
|------|-------|-------------|
| `blocking.tsv` | `pgstatviz.blocking` | Contains the blocking lock history with blocked and blocker counts, plus per-lock-type counts (JSONB, pg_statviz 1.2 and later). |
| `buf.tsv` | `pgstatviz.buf` | Contains buffer and checkpoint statistics. |
| `conf.tsv` | `pgstatviz.conf` | Contains configuration snapshots (JSONB). |
| `conn.tsv` | `pgstatviz.conn` | Contains connection statistics (JSONB). |
| `db.tsv` | `pgstatviz.db` | Contains database statistics. |
| `io.tsv` | `pgstatviz.io` | Contains I/O statistics (JSONB, PostgreSQL 16 and later). |
| `lock.tsv` | `pgstatviz.lock` | Contains lock statistics (JSONB). |
| `repl.tsv` | `pgstatviz.repl` | Contains replication statistics (JSONB). |
| `slru.tsv` | `pgstatviz.slru` | Contains SLRU cache statistics (JSONB). |
| `snapshots.tsv` | `pgstatviz.snapshots` | Contains the snapshot timestamps. |
| `wait.tsv` | `pgstatviz.wait` | Contains wait event statistics (JSONB). |
| `wal.tsv` | `pgstatviz.wal` | Contains WAL statistics (PostgreSQL 14 and later). |
