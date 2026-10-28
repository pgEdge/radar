# Setting Environment Variables

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

The pgx driver also applies the other libpq environment variables that radar
does not set itself, such as `PGAPPNAME` and `PGCONNECT_TIMEOUT`.
