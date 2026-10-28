# radar

radar is an agentless diagnostic data collection tool for PostgreSQL and for
the system that hosts it.

radar collects a snapshot of system and PostgreSQL diagnostic data for
troubleshooting and analysis. radar runs system commands and PostgreSQL
queries, and streams the output into a timestamped ZIP archive. A PostgreSQL
connection is optional; without one, radar collects system data only. When a
command, file, or view is unavailable, radar skips that collector and continues
the run.

pgEdge publishes radar as RPM and DEB packages and as pre-built binaries for
Linux and macOS, so you do not need to build radar yourself. The
[Installing radar](installing/index.md) section describes each method, starting
with the simplest.
