# Changelog

All notable changes to the pgEdge radar project will be
documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-12-23

### Added
- `pg_stat_checkpointer` collector (PostgreSQL 17+) - checkpoint operations
- `pg_stat_io` collector (PostgreSQL 16+) - I/O statistics by backend type
- `pg_stat_wal` collector (PostgreSQL 14+) - WAL generation statistics
- `pg_postmaster_start_time()` collector - server startup timestamp
- `pg_stat_database` per-database collector - conflicts, deadlocks, temp files

## [0.1.0] - 2025-12-18

### Added
- Initial release
- Static binaries for Linux (amd64, arm64) and macOS (amd64, arm64)
- Comprehensive system and PostgreSQL metrics collection
