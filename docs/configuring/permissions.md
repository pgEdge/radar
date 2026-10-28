# Choosing Permissions

The privileges of the operating system user and of the PostgreSQL role
determine how much data radar can collect.

## Running as Root with a Superuser

For a complete collection, run radar as root with a PostgreSQL superuser
account; this is the recommended combination. The following command runs a
collection as root with the `postgres` superuser:

```bash
PGPASSWORD='postgres_password' radar -d mydatabase -U postgres
```

This combination gives radar:

- full system-level access to operating system metrics, logs, and configuration
  files.
- complete access to the PostgreSQL catalog, including all views, tables, and
  system information.
- the most diagnostic data for troubleshooting.

## Running as Root with the pg_monitor Role

Where your policy restricts superuser access in production, grant the
`pg_monitor` role to a dedicated user instead. The `pg_monitor` role provides
most of the diagnostic data without superuser privileges.

The following commands create a monitoring user as the `postgres` operating
system user, and then run a collection as root:

```bash
# As the postgres user, create a monitoring user with the pg_monitor role
psql -c "CREATE USER radaruser WITH PASSWORD 'secure_password';"
psql -c "GRANT pg_monitor TO radaruser;"

# As root, run a collection
PGPASSWORD='secure_password' radar -d mydatabase -U radaruser
```

The `pg_monitor` role can read nearly every source that radar queries. A few
replication catalog views and `pg_hba_file_rules` remain unreadable, and radar
reports a permission denial for each affected collector. The run still
completes and writes an archive, and radar still reads `pg_hba.conf` from the
data directory.

## Running with Limited Permissions

radar can run as a non-root user with limited PostgreSQL permissions. radar
then reports an error for each system file that it cannot read and each command
that fails, and some PostgreSQL catalog queries fail; the run continues and
writes an archive.
