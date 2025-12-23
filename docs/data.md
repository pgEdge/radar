# radar - Data Collection Reference

Complete reference of all data collected by radar.

## Summary

- **Cross-Platform System**: Collectors that work on both Linux and macOS
- **Linux-Specific System**: Linux-only collectors (systemd, /proc, /sys, etc.)
- **macOS-Specific System**: macOS-only collectors (sysctl, system_profiler, diskutil, etc.)
- **PostgreSQL Instance**: Instance-level PostgreSQL collectors
- **Per-Database**: Schema and object collectors (per database)
- **pg_statviz (optional)**: Time-series statistics collectors (per database)

**Output**: All data collected in a single ZIP file named `radar-{hostname}-{timestamp}.zip`

---

## Cross-Platform System Collectors

These collectors work on both Linux and macOS.

| File | Source | Description |
|------|--------|-------------|
| `system/diskspace.out` | `df -h` | Disk space usage |
| `system/dmesg.out` | `dmesg` | Kernel ring buffer |
| `system/fstab.out` | `/etc/fstab` | Filesystem table |
| `system/hostname.out` | `hostname -f` | Fully qualified hostname |
| `system/hosts.out` | `/etc/hosts` | Host name resolution |
| `system/locale.out` | `locale` | Current locale settings |
| `system/locale_all.out` | `locale -a` | Available locales |
| `system/mount.out` | `mount` | Mounted filesystems |
| `system/openssl/ciphers.out` | `openssl ciphers` | Available SSL/TLS ciphers |
| `system/openssl/engines.out` | `openssl engine` | OpenSSL engines |
| `system/openssl/version.out` | `openssl version -a` | OpenSSL version details |
| `system/ps.out` | `ps auxww` | Process list |
| `system/sysctl.out` | `sysctl -a` | Kernel parameters |
| `system/top.out` | `top -l 1` | Process snapshot |
| `system/uname.out` | `uname -a` | Kernel version and system info |

---

## Linux-Specific System Collectors

These collectors only run on Linux systems.

| File | Source | Description |
|------|--------|-------------|
| `system/dmesg_t.out` | `dmesg -T` | Kernel ring buffer with timestamps |
| `system/hypervisor.out` | `systemd-detect-virt` | Hypervisor detection |
| `system/ifconfig.out` | `ifconfig -a` | Network interfaces (legacy) |
| `system/interfaces.out` | `ip -o address` | Network interfaces (one-line) |
| `system/io_schedulers.out` | `/sys/block/*/queue/scheduler` | I/O scheduler settings |
| `system/iostat.out` | `iostat -x 1 5` | I/O statistics (5 samples) |
| `system/ip_addr.out` | `ip address list` | IP addresses |
| `system/ipcs.out` | `ipcs -a` | IPC resources |
| `system/limits.out` | `/etc/security/limits.conf` | System resource limits |
| `system/locale_conf.out` | `/etc/locale.conf` | Locale configuration |
| `system/localectl.out` | `localectl status` | Locale and keymap settings |
| `system/lsblk.out` | `lsblk` | Block device layout |
| `system/lsdevmapper.out` | `ls -la /dev/mapper` | Device mapper devices |
| `system/lsmod.out` | `lsmod` | Loaded kernel modules |
| `system/lspci.out` | `lspci` | PCI devices |
| `system/machine_id.out` | `/etc/machine-id` | Machine identifier |
| `system/mpstat.out` | `mpstat -P ALL 1 5` | Per-CPU statistics |
| `system/nfsiostat.out` | `nfsiostat` | NFS I/O statistics |
| `system/openssl/crypto-policies-isapplied.out` | `update-crypto-policies --is-applied` | Crypto policy status |
| `system/openssl/crypto-policies-show.out` | `update-crypto-policies --show` | Active crypto policy |
| `system/openssl/fips-mode-setup.out` | `fips-mode-setup --check` | FIPS mode status |
| `system/os_release.out` | `/etc/os-release` | OS distribution info |
| `system/packages-apt-list-installed.out` | `apt list --installed *postgres*` | APT packages (Debian/Ubuntu) |
| `system/packages-dnf-list-installed.out` | `dnf list installed *postgres*` | DNF packages (Fedora/RHEL 8+) |
| `system/packages-dpkg.out` | `dpkg -l *postgres*` | Debian packages |
| `system/packages-rpm.out` | `rpm -qa *postgres*` | RPM packages |
| `system/packages-yum-list-installed.out` | `yum list installed *postgres*` | YUM packages (RHEL/CentOS) |
| `system/proc/cpuinfo.out` | `/proc/cpuinfo` | CPU information |
| `system/proc/loadavg.out` | `/proc/loadavg` | Load average |
| `system/proc/meminfo.out` | `/proc/meminfo` | Memory information |
| `system/proc/mounts.out` | `/proc/mounts` | Mounted filesystems |
| `system/proc/pressure_cpu.out` | `/proc/pressure/cpu` | CPU pressure stall information |
| `system/proc/pressure_io.out` | `/proc/pressure/io` | I/O pressure stall information |
| `system/proc/pressure_memory.out` | `/proc/pressure/memory` | Memory pressure stall information |
| `system/proc/swaps.out` | `/proc/swaps` | Swap space usage |
| `system/proc/uptime.out` | `/proc/uptime` | System uptime |
| `system/proc/vmstat.out` | `/proc/vmstat` | Virtual memory statistics |
| `system/read_ahead.out` | `blockdev --getra /dev/*` | Block device read-ahead settings |
| `system/sar.out` | `sar -A` | System activity report |
| `system/sestatus.out` | `sestatus` | SELinux status |
| `system/sys/cpu_scaling_available_governors.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_available_governors` | Available CPU governors |
| `system/sys/cpu_scaling_driver.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_driver` | CPU frequency scaling driver |
| `system/sys/cpu_scaling_governor.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_governor` | Active CPU governor |
| `system/sys/energy_perf_bias.out` | `/sys/devices/system/cpu/cpu*/power/energy_perf_bias` | CPU energy performance bias |
| `system/sys/intel_pstate.out` | `/sys/devices/system/cpu/intel_pstate/*` | Intel P-state settings |
| `system/sys/kernel_mm_transparent_hugepage.out` | `/sys/kernel/mm/transparent_hugepage/*` | Transparent hugepage settings |
| `system/system_release.out` | `/etc/system-release` | System release info |
| `system/systemd/list-units.out` | `systemctl list-units --all` | Systemd units |
| `system/tuned/tuned-active.out` | `tuned-adm active` | Active tuned profile |
| `system/tuned/tuned-list.out` | `tuned-adm list` | Available tuned profiles |
| `system/vmstat-command.out` | `vmstat 1 10` | Virtual memory statistics (10 samples) |

---

## macOS-Specific System Collectors

These collectors only run on macOS systems.

| File | Source | Description |
|------|--------|-------------|
| `system/diskutil_info_all.out` | `diskutil info` (all disks) | Detailed disk information |
| `system/diskutil_list.out` | `diskutil list` | Disk layout |
| `system/hypervisor.out` | `sysctl kern.hv_vmm_present` | Hypervisor detection |
| `system/ifconfig.out` | `ifconfig -a` | Network interfaces |
| `system/iostat.out` | `iostat -c 5 -w 1` | I/O statistics (5 samples) |
| `system/ipcs.out` | `ipcs -a` | IPC resources |
| `system/kextstat.out` | `kextstat` | Loaded kernel extensions |
| `system/launchctl_list.out` | `launchctl list` | Launch daemons and agents |
| `system/netstat_interfaces.out` | `netstat -i` | Network interface statistics |
| `system/netstat_routing.out` | `netstat -r` | Routing table |
| `system/packages_brew.out` | `brew list --versions` | Homebrew packages |
| `system/packages_brew_postgres.out` | `brew list \| grep postgres` | PostgreSQL Homebrew packages |
| `system/pmset_assertions.out` | `pmset -g assertions` | Power management assertions |
| `system/pmset_settings.out` | `pmset -g` | Power management settings |
| `system/sysctl.conf` | `/etc/sysctl.conf` | Kernel parameter configuration |
| `system/sysctl_cpu.out` | `sysctl -a machdep.cpu` | CPU information |
| `system/sysctl_hw.out` | `sysctl -a hw` | Hardware information |
| `system/sysctl_kern.out` | `sysctl -a kern` | Kernel information |
| `system/sysctl_vm.out` | `sysctl -a vm` | Virtual memory settings |
| `system/system_log_boot.out` | `log show --last boot` | System log since boot |
| `system/system_profiler_hardware.out` | `system_profiler SPHardwareDataType` | Hardware overview |
| `system/system_profiler_network.out` | `system_profiler SPNetworkDataType` | Network configuration |
| `system/system_profiler_pci.out` | `system_profiler SPPCIDataType` | PCI devices |
| `system/system_profiler_software.out` | `system_profiler SPSoftwareDataType` | Software overview |
| `system/system_profiler_storage.out` | `system_profiler SPStorageDataType` | Storage devices |
| `system/system_version.plist` | `/System/Library/CoreServices/SystemVersion.plist` | macOS version |
| `system/vm_stat.out` | `vm_stat` | Virtual memory statistics |
| `system/vm_stat_interval.out` | `vm_stat -c 10 1` | Virtual memory statistics (10 samples) |

---

## PostgreSQL Instance Collectors

Instance-level PostgreSQL collectors. Files stored in `postgresql/`.

| File | Source | Description |
|------|--------|-------------|
| `postgresql/archiver.tsv` | `pg_stat_archiver` | WAL archiver statistics |
| `postgresql/available_extensions.tsv` | `pg_available_extensions` | Available extensions |
| `postgresql/bgwriter.tsv` | `pg_stat_bgwriter` | Background writer statistics |
| `postgresql/blocking_locks.tsv` | Complex query | Blocking/blocked lock pairs |
| `postgresql/checkpointer.tsv` | `pg_stat_checkpointer` | Checkpoint statistics (PG17+) |
| `postgresql/configuration.tsv` | `pg_settings` | Configuration parameters |
| `postgresql/databases.tsv` | `pg_database` | Database list |
| `postgresql/databases_checksums.tsv` | `pg_stat_database` | Checksum failure counts |
| `postgresql/db_role_setting.tsv` | `pg_db_role_setting` | Per-database/role settings |
| `postgresql/pg_hba.conf` | Data directory | Host-based authentication config |
| `postgresql/pg_hba_file_rules.tsv` | `pg_hba_file_rules` | Parsed pg_hba.conf rules (PG10+) |
| `postgresql/pg_ident.conf` | Data directory | User name mapping config |
| `postgresql/postmaster_start_time.tsv` | `pg_postmaster_start_time()` | Server startup timestamp |
| `postgresql/postgresql.auto.conf` | Data directory | Auto-generated configuration |
| `postgresql/postgresql.conf` | Data directory | Main configuration file |
| `postgresql/prepared_xacts.tsv` | `pg_prepared_xacts` | Prepared transactions |
| `postgresql/recovery.conf` | Data directory | Recovery configuration (PG11-) |
| `postgresql/recovery.done` | Data directory | Recovery completion marker |
| `postgresql/replication.tsv` | `pg_stat_replication` | Replication status |
| `postgresql/replication_origin.tsv` | `pg_replication_origin_status` | Replication origin status |
| `postgresql/replication_slots.tsv` | `pg_replication_slots` | Replication slots |
| `postgresql/roles.tsv` | `pg_roles` | Database roles |
| `postgresql/running_activity.tsv` | `pg_stat_activity` | Active connections and queries |
| `postgresql/running_activity_maxage.tsv` | Complex query | Oldest queries/transactions |
| `postgresql/running_locks.tsv` | `pg_locks WHERE granted` | Held locks |
| `postgresql/stat_io.tsv` | `pg_stat_io` | I/O statistics by backend type (PG16+) |
| `postgresql/stat_progress_analyze.tsv` | `pg_stat_progress_analyze` | ANALYZE progress (PG13+) |
| `postgresql/stat_progress_basebackup.tsv` | `pg_stat_progress_basebackup` | Base backup progress (PG13+) |
| `postgresql/stat_progress_copy.tsv` | `pg_stat_progress_copy` | COPY progress (PG14+) |
| `postgresql/stat_progress_vacuum.tsv` | `pg_stat_progress_vacuum` | VACUUM progress (PG9.6+) |
| `postgresql/stat_slru.tsv` | `pg_stat_slru` | SLRU cache statistics |
| `postgresql/stat_wal.tsv` | `pg_stat_wal` | WAL generation statistics (PG14+) |
| `postgresql/subscriptions.tsv` | `pg_subscription` | Logical replication subscriptions |
| `postgresql/tablespaces.tsv` | `pg_tablespace` | Tablespace definitions |
| `postgresql/version.tsv` | `version()` | PostgreSQL version |
| `postgresql/waits_sample.tsv` | `pg_stat_activity` | Active wait events |

---

## Per-Database Collectors

Collected for each accessible database. Files stored in `databases/{dbname}/`.

| File | Source | Description |
|------|--------|-------------|
| `extensions.tsv` | `pg_extension` | Installed extensions |
| `funcs.tsv` | `pg_proc WHERE prokind='f'` | Functions |
| `indexes.tsv` | `pg_indexes` | Indexes |
| `languages.tsv` | `pg_language` | Procedural languages |
| `operators.tsv` | `pg_operator` | Operators |
| `partitioned_tables.tsv` | `pg_partitioned_table` | Partitioned tables (PG10+) |
| `partitions.tsv` | `pg_inherits` | Partition relationships |
| `procs.tsv` | `pg_proc WHERE prokind='p'` | Procedures (PG11+) |
| `publication_tables.tsv` | `pg_publication_tables` | Tables in publications |
| `publications.tsv` | `pg_publication` | Logical replication publications |
| `schemas.tsv` | `pg_namespace` | Schemas |
| `stat_database.tsv` | `pg_stat_database` | Database statistics (conflicts, deadlocks, temp files) |
| `statistics.tsv` | `pg_statistic_ext` | Extended statistics (PG10+) |
| `subscription_tables.tsv` | `pg_subscription_rel` | Subscription relation states |
| `tables.tsv` | `pg_tables` | Tables |
| `triggers.tsv` | `pg_trigger` | Triggers |
| `types.tsv` | `pg_type` | Data types |

---

## pg_statviz Collectors (Optional)

If the pg_statviz extension is installed in a database, these collectors are available. Files stored in `pg_statviz/{dbname}/`.

| File | Table | Description |
|------|-------|-------------|
| `buf.tsv` | `pgstatviz.buf` | Buffer and checkpoint statistics |
| `conf.tsv` | `pgstatviz.conf` | Configuration snapshots (JSONB) |
| `conn.tsv` | `pgstatviz.conn` | Connection statistics (JSONB) |
| `db.tsv` | `pgstatviz.db` | Database statistics |
| `io.tsv` | `pgstatviz.io` | I/O statistics (JSONB, PG16+) |
| `lock.tsv` | `pgstatviz.lock` | Lock statistics (JSONB) |
| `snapshots.tsv` | `pgstatviz.snapshots` | Snapshot timestamps |
| `wait.tsv` | `pgstatviz.wait` | Wait event statistics (JSONB) |
| `wal.tsv` | `pgstatviz.wal` | WAL statistics (PG14+) |
