# Configuring radar

radar takes its settings from command-line options and from the standard
PostgreSQL environment variables; radar has no configuration file.

## Setting the Connection Defaults

The `-h`, `-U`, and `-d` options are optional, and radar resolves each value as
follows:

- `-h` falls back to `PGHOST`, and then to `localhost` over TCP; radar does not
  default to a Unix-domain socket.
- `-U` falls back to `PGUSER`, then to the name of the current operating system
  user, and then to `postgres` when radar cannot determine that user.
- `-d` falls back to `PGDATABASE`, and then to `postgres`; with
  `--skip-system`, radar requires a database name from `-d` or `PGDATABASE`.

## Setting Environment Variables

The following table describes the environment variables that radar reads:

| Variable | Description |
|----------|-------------|
| `PGHOST` | Specifies the PostgreSQL host. |
| `PGPORT` | Specifies the PostgreSQL port. |
| `PGUSER` | Specifies the PostgreSQL user name. |
| `PGPASSWORD` | Specifies the PostgreSQL password. |
| `PGPASSFILE` | Names the password file that the driver reads when no password is supplied; the default is `~/.pgpass`. |
| `PGDATABASE` | Specifies the PostgreSQL database name. |
| `PGSSLMODE` | Specifies the SSL mode: `prefer`, `disable`, `require`, `verify-ca`, or `verify-full`. radar rejects any other value. |
| `PGSSLCERT` | Specifies the client SSL certificate file. |
| `PGSSLKEY` | Specifies the client SSL private key file. |
| `PGSSLROOTCERT` | Specifies the SSL root (CA) certificate file. |

The pgx driver also applies some other libpq environment variables, such as
`PGAPPNAME` and `PGCONNECT_TIMEOUT`. pgx does not support `PGSSLCRL` or
`PGSSLCRLDIR`, so radar does not check certificate revocation lists.
