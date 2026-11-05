# Choosing an Authentication Method

The following table describes how to configure radar for each authentication
method:

| Method | Configuration |
|--------|---------------|
| Password | Supply the password with `PGPASSWORD`, or in a password file: `~/.pgpass`, or the file that `PGPASSFILE` names. The user name comes from `-U` or its defaults in [Setting the Connection Defaults](index.md#setting-the-connection-defaults). |
| LDAP | The server performs LDAP authentication, so the client needs no changes; supply the user name and password as for password authentication. |
| Certificate | Run radar with `--sslmode verify-full --sslcert client.crt --sslkey client.key --sslrootcert ca.crt`. |

radar does not support GSSAPI (Kerberos) authentication.
