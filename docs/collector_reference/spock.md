# Collecting Spock Data

The Spock collectors are optional; these collectors are available when a
database has the [Spock](https://github.com/pgEdge/spock) extension installed.
radar stores their files in `spock/{dbname}/`.

Three Spock relations contain data that must never enter an archive, because an
archive routinely accompanies a support ticket. The `exception_log` relation
holds jsonb images of the conflicting rows, and the `resolutions` relation
holds the same row images as text. The `node_interface.if_dsn` column holds the
node connection string, including its password. The two tasks that read
`exception_log` and `resolutions` name their columns explicitly rather than
using `SELECT *`. radar does not collect `node_interface` at all.

The following table describes the Spock files and the relation that each file
comes from:

| File | Relation | Description |
|------|----------|-------------|
| `channel_summary_stats.tsv` | `spock.channel_summary_stats` | Reports per-subscription counters for rows applied, conflicts and delta-apply conflicts. The file contains aggregates only and no row data. |
| `exception_log.tsv` | `spock.exception_log` | Lists the last 1000 apply exceptions by retry time. The file excludes `local_tup`, `remote_old_tup` and `remote_new_tup`, which hold jsonb row images. The file keeps `error_message`, the primary error message with its SQLSTATE. Spock stores no `DETAIL`, so the file does not include the offending key of a constraint violation; however, an error whose primary message embeds a value (a failed type conversion, for example) would include that value. |
| `lag_tracker.tsv` | `spock.lag_tracker` | Reports the replication lag in bytes and time for each node pair. |
| `local_node.tsv` | `spock.local_node` | Identifies which node this database is. |
| `local_sync_status.tsv` | `spock.local_sync_status` | Reports the initial synchronization state of each relation. |
| `node.tsv` | `spock.node` | Lists the cluster nodes with their id, name, location and country. The file excludes `info`, which holds deployment-defined jsonb. |
| `pii.tsv` | `spock.pii` | Lists the columns registered as holding personally identifiable information. The file contains column names only and no values. |
| `progress.tsv` | `spock.progress` | Reports the apply progress of each origin, with the commit timestamp and LSNs. |
| `replication_set.tsv` | `spock.replication_set` | Lists the replication sets and the operations that each set replicates. |
| `replication_set_table.tsv` | `spock.replication_set_table` | Lists the tables in each replication set, with the column list and row filter. |
| `resolutions.tsv` | `spock.resolutions` | Lists the last 1000 conflict resolutions by log time. The file excludes `local_tuple` and `remote_tuple`, which hold text row images. |
| `subscription.tsv` | `spock.subscription` | Lists the subscriptions with their enabled state, slot name, replication sets and apply delay. |
| `tables.tsv` | `spock.tables` | Lists the user tables and the replication set, if any, that each table belongs to. |
