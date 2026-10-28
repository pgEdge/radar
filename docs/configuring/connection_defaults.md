# Setting the Connection Defaults

The `-h`, `-U`, and `-d` options are optional, and radar resolves each value as
follows:

- `-h` falls back to `PGHOST`, and then to `localhost` over TCP; radar does not
  default to a Unix-domain socket.
- `-U` falls back to `PGUSER`, then to the name of the current operating system
  user, and then to `postgres` when radar cannot determine that user.
- `-d` falls back to `PGDATABASE`, and then to `postgres`; with
  `--skip-system`, radar requires a database name from `-d` or `PGDATABASE`.
