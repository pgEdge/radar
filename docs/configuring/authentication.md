# Choosing an Authentication Method

The following table describes how to configure radar for each authentication
method:

| Method | Configuration |
|--------|---------------|
| Password | Supply the password with `PGPASSWORD`, or in a password file: `~/.pgpass`, or the file that `PGPASSFILE` names. The user name comes from `-U` or its defaults in [Setting the Connection Defaults](index.md#setting-the-connection-defaults). |
| LDAP | The server performs LDAP authentication, so the client needs no changes; supply the user name and password as for password authentication. |
| Certificate | Run radar with `--sslmode verify-full --sslcert client.crt --sslkey client.key --sslrootcert ca.crt`. |
| GSSAPI (Kerberos) | Obtain a Kerberos ticket with `kinit` as the user that runs radar; radar reads the ticket from a file credential cache. With a KCM or KEYRING cache, run `kinit -c FILE:/tmp/krb5cc_$(id -u)` instead. radar supports GSSAPI authentication but not GSSAPI encryption, so a `hostgssenc` line in `pg_hba.conf` does not match radar's connections. |
