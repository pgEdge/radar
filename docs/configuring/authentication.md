# Choosing an Authentication Method

The following table describes how to configure radar for each authentication
method:

| Method | Configuration |
|--------|---------------|
| Password | Supply the user with `-U` and the password with `PGPASSWORD`, or rely on the operating system user defaults. |
| LDAP | The server performs LDAP authentication, so the client needs no changes; supply the user name and password as for password authentication. |
| Certificate | Run radar with `--sslmode verify-full --sslcert client.crt --sslkey client.key --sslrootcert ca.crt`. |
| GSSAPI (Kerberos) | Obtain a valid Kerberos ticket with `kinit`; radar uses the ticket automatically. |
