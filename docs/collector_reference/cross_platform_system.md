# Collecting Cross-Platform System Data

The cross-platform system collectors work on both Linux and macOS. The
following table describes the files that these collectors produce:

| File | Source | Description |
|------|--------|-------------|
| `system/diskspace.out` | `df -h` | Reports disk space usage. |
| `system/dmesg.out` | `dmesg` | Contains the kernel ring buffer. |
| `system/hosts.out` | `/etc/hosts` | Contains the host name resolution entries. |
| `system/locale.out` | `locale` | Reports the current locale settings. |
| `system/locale_all.out` | `locale -a` | Lists the available locales. |
| `system/mount.out` | `mount` | Lists the mounted filesystems. |
| `system/openssl/ciphers.out` | `openssl ciphers` | Lists the available SSL/TLS ciphers. |
| `system/openssl/engines.out` | `openssl engine` | Lists the OpenSSL engines. |
| `system/openssl/version.out` | `openssl version -a` | Reports the OpenSSL version details. |
| `system/ps.out` | `ps auxww` | Contains the process list. |
| `system/resolv_conf.out` | `/etc/resolv.conf` | Contains the DNS resolver configuration. |
| `system/sysctl.out` | `sysctl -a` | Lists the kernel parameters. |
| `system/uname.out` | `uname -a` | Reports the kernel version and system information. |
