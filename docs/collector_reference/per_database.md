# Collecting Per-Database Data

radar runs the per-database collectors for each accessible database and stores
their files in `databases/{dbname}/`. The following table describes the files
that these collectors produce for each database:

| File | Source | Description |
|------|--------|-------------|
| `bloat.tsv` | `pg_stats` heuristic | Estimates table bloat with a heuristic. |
| `extensions.tsv` | `pg_extension` | Lists the installed extensions. |
| `funcs.tsv` | `pg_proc WHERE prokind='f'` | Lists the functions. |
| `indexes.tsv` | `pg_index` + `pg_stat_all_indexes` | Lists the top 1000 indexes by size with the semantic key, validity, scan counters and size. |
| `languages.tsv` | `pg_language` | Lists the procedural languages. |
| `operators.tsv` | `pg_operator` | Lists the operators. |
| `partitioned_tables.tsv` | `pg_partitioned_table` | Lists the partitioned tables (PostgreSQL 10 and later). |
| `partitions.tsv` | `pg_inherits` | Lists the partition relationships. |
| `pgstattuple.tsv` | `pgstattuple_approx()` | Reports bloat when the `pgstattuple` extension is installed, with exact dead-tuple figures and approximate live-tuple and free-space figures. The collector is disabled by default because it reads every heap page that the visibility map does not mark all-visible, in every user table; run radar with `--include pgstattuple` to enable the collector. |
| `procs.tsv` | `pg_proc WHERE prokind='p'` | Lists the procedures (PostgreSQL 11 and later). |
| `publication_tables.tsv` | `pg_publication_tables` | Lists the tables in publications. |
| `publications.tsv` | `pg_publication` | Lists the logical replication publications. |
| `schemas.tsv` | `pg_namespace` | Lists the schemas. |
| `sequences.tsv` | `pg_sequences` | Lists the sequences with their last, minimum and maximum values (PostgreSQL 10 and later). |
| `stat_database.tsv` | `pg_stat_database` | Reports per-database statistics. |
| `statistics.tsv` | `pg_statistic_ext` | Lists the extended statistics (PostgreSQL 10 and later). |
| `subscription_tables.tsv` | `pg_subscription_rel` | Reports the subscription relation states. |
| `table_freeze_age.tsv` | `pg_class` | Lists the top 1000 relations by transaction-ID and multixact freeze age. The ranking uses age rather than size, so the file includes a small table that holds the wraparound horizon back where `tables.tsv` would miss that table. The file also covers the `pg_catalog` and `pg_toast` relations that `tables.tsv` excludes. The query selects only relations whose `relkind` is `r`, `m` or `t` (ordinary tables, materialized views and TOAST tables). |
| `tables.tsv` | `pg_class` + `pg_stat_all_tables` | Lists the top 1000 tables by size with persistence, options, heap and table sizes, toast mapping, dead-tup counters, vacuum and analyze timestamps, per-table vacuum and analyze ages in seconds, and transaction-ID and multixact freeze ages. The freeze ages are empty for partitioned tables, which hold no tuples of their own. The query requires PostgreSQL 13 or later. |
| `triggers.tsv` | `pg_trigger` | Lists the triggers. |
| `types.tsv` | `pg_type` | Lists the data types. |
