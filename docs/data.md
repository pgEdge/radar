# radar Data Collection Reference

This reference describes every file that radar collects. radar writes all
collected data to a single ZIP file named `radar-{hostname}-{timestamp}.zip`.
The archive root contains a `radar.out` entry that identifies the radar binary
that produced the archive. The `version` value in that entry comes from the
build-time `-X main.version` stamp. The `commit` value comes from the version
control information that Go embeds in the binary.

This reference describes the following collectors:

- Cross-platform system collectors work on both Linux and macOS.
- Linux-specific system collectors run only on Linux and use sources such as
  systemd, `/proc` and `/sys`.
- macOS-specific system collectors run only on macOS and use tools such as
  `sysctl`, `system_profiler` and `diskutil`.
- PostgreSQL instance collectors gather instance-level PostgreSQL data.
- Per-database collectors gather schema and object data for each database.
- The optional pg_statviz collectors gather time-series statistics for each
  database.
- The optional Spock collectors gather multi-master replication state for each
  database.
- The optional PgBouncer collectors gather the pooler configuration when
  PgBouncer is installed.

---

## Collecting Cross-Platform System Data

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

---

## Collecting Linux System Data

The Linux-specific system collectors run only on Linux systems. The following
table describes the files that these collectors produce:

| File | Source | Description |
|------|--------|-------------|
| `system/dmesg_t.out` | `dmesg -T` | Contains the kernel ring buffer with timestamps. |
| `system/free.out` | `free -h` | Reports a memory usage summary. |
| `system/fstab.out` | `/etc/fstab` | Contains the filesystem table. |
| `system/hostname.out` | `hostname` | Reports the kernel hostname (`gethostname`). |
| `system/hostname_fqdn.out` | `hostname -f` | Reports the fully qualified hostname (the resolver canonical name). |
| `system/hypervisor.out` | `systemd-detect-virt` | Reports the hypervisor detection result. |
| `system/ifconfig.out` | `ifconfig -a` | Lists the network interfaces (legacy). |
| `system/interfaces.out` | `ip -o address` | Lists the network interfaces in one-line format. |
| `system/io_queue_depth.out` | `/sys/block/*/queue/nr_requests` | Reports the I/O queue depth of each device. |
| `system/io_schedulers.out` | `/sys/block/*/queue/scheduler` | Reports the I/O scheduler settings. |
| `system/iostat.out` | `iostat -x 1 5` | Reports I/O statistics over 5 samples. |
| `system/ip_addr.out` | `ip address list` | Lists the IP addresses. |
| `system/ipcs.out` | `ipcs -a` | Lists the inter-process communication (IPC) resources. |
| `system/limits.out` | `/etc/security/limits.conf` | Contains the system resource limits. |
| `system/locale_conf.out` | `/etc/locale.conf` | Contains the locale configuration. |
| `system/localectl.out` | `localectl status` | Reports the locale and keymap settings. |
| `system/lsblk.out` | `lsblk` | Shows the block device layout. |
| `system/lscpu.out` | `lscpu` | Reports the CPU architecture and topology. |
| `system/lsdevmapper.out` | `ls -la /dev/mapper` | Lists the device mapper devices. |
| `system/lsmod.out` | `lsmod` | Lists the loaded kernel modules. |
| `system/lspci.out` | `lspci` | Lists the PCI devices. |
| `system/machine_id.out` | `/etc/machine-id` | Contains the machine identifier. |
| `system/mpstat.out` | `mpstat -P ALL 1 5` | Reports per-CPU statistics. |
| `system/netstat_stats.out` | `netstat -s` | Reports protocol statistics. |
| `system/nfsiostat.out` | `nfsiostat` | Reports NFS I/O statistics. |
| `system/numactl.out` | `numactl --hardware` | Reports the NUMA node layout and memory. |
| `system/numastat.out` | `numastat -m` | Reports per-node memory allocation statistics. |
| `system/openssl/crypto-policies-isapplied.out` | `update-crypto-policies --is-applied` | Reports the crypto policy status. |
| `system/openssl/crypto-policies-show.out` | `update-crypto-policies --show` | Reports the active crypto policy. |
| `system/openssl/fips-mode-setup.out` | `fips-mode-setup --check` | Reports the FIPS mode status. |
| `system/os_release.out` | `/etc/os-release` | Contains the OS distribution information. |
| `system/packages-apt-list-installed.out` | `apt list --installed *postgres*` | Lists the APT packages (Debian and Ubuntu). |
| `system/packages-dnf-list-installed.out` | `dnf list installed *postgres*` | Lists the DNF packages (Fedora and RHEL 8 or later). |
| `system/packages-dpkg.out` | `dpkg -l *postgres*` | Lists the Debian packages. |
| `system/packages-rpm.out` | `rpm -qa *postgres*` | Lists the RPM packages. |
| `system/packages-yum-list-installed.out` | `yum list installed *postgres*` | Lists the YUM packages (RHEL and CentOS). |
| `system/proc/cpuinfo.out` | `/proc/cpuinfo` | Contains the CPU information. |
| `system/proc/diskstats.out` | `/proc/diskstats` | Contains the raw kernel I/O counters. |
| `system/proc/loadavg.out` | `/proc/loadavg` | Reports the load average. |
| `system/proc/meminfo.out` | `/proc/meminfo` | Contains the memory information. |
| `system/proc/mounts.out` | `/proc/mounts` | Lists the mounted filesystems. |
| `system/proc/pressure_cpu.out` | `/proc/pressure/cpu` | Reports CPU pressure stall information. |
| `system/proc/pressure_io.out` | `/proc/pressure/io` | Reports I/O pressure stall information. |
| `system/proc/pressure_memory.out` | `/proc/pressure/memory` | Reports memory pressure stall information. |
| `system/proc/swaps.out` | `/proc/swaps` | Reports swap space usage. |
| `system/proc/uptime.out` | `/proc/uptime` | Reports the system uptime. |
| `system/proc/vmstat.out` | `/proc/vmstat` | Reports virtual memory statistics. |
| `system/read_ahead.out` | `blockdev --getra /dev/*` | Reports the block device read-ahead settings. |
| `system/sar.out` | `sar -A` | Contains the system activity report. |
| `system/sestatus.out` | `sestatus` | Reports the SELinux status. |
| `system/ss_listeners.out` | `ss -tunlp` | Lists the listening TCP and UDP sockets. |
| `system/ss_summary.out` | `ss -s` | Reports a socket statistics summary. |
| `system/sys/clocksource.out` | `/sys/devices/system/clocksource/clocksource0/current_clocksource` | Reports the current clocksource. |
| `system/sys/cpu_scaling_available_governors.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_available_governors` | Lists the available CPU governors. |
| `system/sys/cpu_scaling_driver.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_driver` | Reports the CPU frequency scaling driver. |
| `system/sys/cpu_scaling_governor.out` | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_governor` | Reports the active CPU governor. |
| `system/sys/energy_perf_bias.out` | `/sys/devices/system/cpu/cpu*/power/energy_perf_bias` | Reports the CPU energy performance bias. |
| `system/sys/intel_pstate.out` | `/sys/devices/system/cpu/intel_pstate/*` | Reports the Intel P-state settings. |
| `system/sys/kernel_mm_transparent_hugepage.out` | `/sys/kernel/mm/transparent_hugepage/*` | Reports the transparent hugepage settings. |
| `system/system_release.out` | `/etc/system-release` | Contains the system release information. |
| `system/systemd/list-units.out` | `systemctl list-units --all` | Lists the systemd units. |
| `system/systemd/postgresql-status.out` | `systemctl status 'postgresql*'` | Reports the PostgreSQL service status. |
| `system/timedatectl.out` | `timedatectl status` | Reports the NTP synchronization and timezone settings. |
| `system/top.out` | `top -b -c -w 512 -n 1` | Contains a process snapshot. |
| `system/tuned/tuned-active.out` | `tuned-adm active` | Reports the active tuned profile. |
| `system/tuned/tuned-list.out` | `tuned-adm list` | Lists the available tuned profiles. |
| `system/vmstat-command.out` | `vmstat 1 10` | Reports virtual memory statistics over 10 samples. |

### Reading Cgroup v2 Resource Limits

radar attempts the cgroup v2 collectors on Linux and skips any collector whose
source is unavailable. Inside a container, these files show the container's
resource limits; on a host, they show the values of the cgroup that radar runs
in. The following table describes the cgroup v2 files:

| File | Source | Description |
|------|--------|-------------|
| `system/cgroup/cpu_max.out` | `/sys/fs/cgroup/cpu.max` | Reports the CPU bandwidth limit (quota and period). |
| `system/cgroup/cpu_weight.out` | `/sys/fs/cgroup/cpu.weight` | Reports the CPU weight (relative share). |
| `system/cgroup/cpuset_cpus_effective.out` | `/sys/fs/cgroup/cpuset.cpus.effective` | Reports the effective CPU set. |
| `system/cgroup/io_max.out` | `/sys/fs/cgroup/io.max` | Reports the I/O bandwidth limits. |
| `system/cgroup/memory_current.out` | `/sys/fs/cgroup/memory.current` | Reports the current memory usage. |
| `system/cgroup/memory_max.out` | `/sys/fs/cgroup/memory.max` | Reports the memory limit. |
| `system/cgroup/memory_stat.out` | `/sys/fs/cgroup/memory.stat` | Reports detailed memory statistics. |
| `system/cgroup/memory_swap_max.out` | `/sys/fs/cgroup/memory.swap.max` | Reports the swap limit. |
| `system/cgroup/pids_current.out` | `/sys/fs/cgroup/pids.current` | Reports the current number of PIDs. |
| `system/cgroup/pids_max.out` | `/sys/fs/cgroup/pids.max` | Reports the PID limit. |

### Reading Cgroup v1 Resource Limits

radar attempts the cgroup v1 collectors on Linux and skips any collector whose
source is unavailable. These files are present on hosts that mount the cgroup
v1 hierarchy, in legacy or hybrid mode. The following table describes the
cgroup v1 files:

| File | Source | Description |
|------|--------|-------------|
| `system/cgroup-v1/cpu_cfs_period_us.out` | `/sys/fs/cgroup/cpu/cpu.cfs_period_us` | Reports the Completely Fair Scheduler (CFS) scheduling period. |
| `system/cgroup-v1/cpu_cfs_quota_us.out` | `/sys/fs/cgroup/cpu/cpu.cfs_quota_us` | Reports the CFS CPU quota; a value of -1 means unlimited. |
| `system/cgroup-v1/cpu_shares.out` | `/sys/fs/cgroup/cpu/cpu.shares` | Reports the CPU shares (relative weight). |
| `system/cgroup-v1/cpuset_cpus.out` | `/sys/fs/cgroup/cpuset/cpuset.cpus` | Lists the allowed CPUs. |
| `system/cgroup-v1/memory_limit_in_bytes.out` | `/sys/fs/cgroup/memory/memory.limit_in_bytes` | Reports the memory limit. |
| `system/cgroup-v1/memory_stat.out` | `/sys/fs/cgroup/memory/memory.stat` | Reports detailed memory statistics. |
| `system/cgroup-v1/memory_usage_in_bytes.out` | `/sys/fs/cgroup/memory/memory.usage_in_bytes` | Reports the current memory usage. |

### Identifying the Cloud Provider and Hardware

radar attempts the cloud and hardware identity collectors on Linux and skips
any collector whose source is unavailable. These files identify the cloud
provider and instance type from Desktop Management Interface (DMI) data. The
following table describes the cloud and hardware identity files:

| File | Source | Description |
|------|--------|-------------|
| `system/cloud/bios_vendor.out` | `/sys/class/dmi/id/bios_vendor` | Reports the BIOS vendor, for example Amazon EC2 or Google. |
| `system/cloud/chassis_asset_tag.out` | `/sys/class/dmi/id/chassis_asset_tag` | Reports the chassis asset tag, for example an AWS instance ID. |
| `system/cloud/product_name.out` | `/sys/class/dmi/id/product_name` | Reports the product name, for example the instance type. |
| `system/cloud/sys_vendor.out` | `/sys/class/dmi/id/sys_vendor` | Reports the system vendor. |

### Identifying the Container Runtime

radar attempts the container identity collectors on Linux and skips any
collector whose source is unavailable. These files help identify the container
runtime and environment. The following table describes the container identity
files:

| File | Source | Description |
|------|--------|-------------|
| `system/container/cgroup_membership.out` | `/proc/1/cgroup` | Reports the cgroup membership (container signatures). |
| `system/container/mountinfo.out` | `/proc/1/mountinfo` | Reports the PID 1 mount information (overlay detection). |

### Detecting a Container Environment

radar detects a container environment automatically on Linux. radar collects
these files only when it runs inside a container (Docker, Kubernetes, LXC or
containerd). The following table describes the container detection files:

| File | Source | Description |
|------|--------|-------------|
| `system/container/environment.out` | `env \| grep` | Records the allowlisted environment variables `HOSTNAME`, `CONTAINER_ID`, `DOCKER_HOST`, `ECS_CLUSTER`, `ECS_CONTAINER_METADATA_URI`, `KUBERNETES_SERVICE_HOST`, `KUBERNETES_SERVICE_PORT` and `KUBERNETES_PORT`. |
| `system/container/k8s_namespace.out` | `/run/secrets/.../namespace` | Reports the Kubernetes namespace. |

---

## Collecting macOS System Data

The macOS-specific system collectors run only on macOS systems. The following
table describes the files that these collectors produce:

| File | Source | Description |
|------|--------|-------------|
| `system/diskutil_info_all.out` | `diskutil info` (all disks) | Reports detailed disk information. |
| `system/diskutil_list.out` | `diskutil list` | Shows the disk layout. |
| `system/hostname.out` | `hostname` | Reports the kernel hostname (`gethostname`). |
| `system/hostname_fqdn.out` | `hostname -f` | Reports the fully qualified hostname (the resolver canonical name). |
| `system/hypervisor.out` | `sysctl kern.hv_vmm_present` | Reports the hypervisor detection result. |
| `system/ifconfig.out` | `ifconfig -a` | Lists the network interfaces. |
| `system/iostat.out` | `iostat -c 5 -w 1` | Reports I/O statistics over 5 samples. |
| `system/ipcs.out` | `ipcs -a` | Lists the IPC resources. |
| `system/kextstat.out` | `kextstat` | Lists the loaded kernel extensions. |
| `system/launchctl_list.out` | `launchctl list` | Lists the launch daemons and agents. |
| `system/memory_pressure.out` | `memory_pressure` | Reports the memory pressure level. |
| `system/netstat_interfaces.out` | `netstat -i` | Reports network interface statistics. |
| `system/netstat_routing.out` | `netstat -r` | Contains the routing table. |
| `system/netstat_stats.out` | `netstat -s` | Reports protocol statistics. |
| `system/packages_brew.out` | `brew list --versions` | Lists the Homebrew packages. |
| `system/packages_brew_postgres.out` | `brew list \| grep postgres` | Lists the PostgreSQL Homebrew packages. |
| `system/pmset_assertions.out` | `pmset -g assertions` | Lists the power management assertions. |
| `system/pmset_settings.out` | `pmset -g` | Reports the power management settings. |
| `system/sysctl.conf` | `/etc/sysctl.conf` | Contains the kernel parameter configuration. |
| `system/sysctl_cpu.out` | `sysctl -a machdep.cpu` | Reports CPU information. |
| `system/sysctl_hw.out` | `sysctl -a hw` | Reports hardware information. |
| `system/sysctl_kern.out` | `sysctl -a kern` | Reports kernel information. |
| `system/sysctl_vm.out` | `sysctl -a vm` | Reports the virtual memory settings. |
| `system/system_log_boot.out` | `log show --predicate 'processID == 0' --last boot --style syslog` | Contains the kernel log entries (process ID 0) since boot. |
| `system/system_profiler_hardware.out` | `system_profiler SPHardwareDataType` | Contains a hardware overview. |
| `system/system_profiler_network.out` | `system_profiler SPNetworkDataType` | Reports the network configuration. |
| `system/system_profiler_pci.out` | `system_profiler SPPCIDataType` | Lists the PCI devices. |
| `system/system_profiler_software.out` | `system_profiler SPSoftwareDataType` | Contains a software overview. |
| `system/system_profiler_storage.out` | `system_profiler SPStorageDataType` | Lists the storage devices. |
| `system/system_version.plist` | `/System/Library/CoreServices/SystemVersion.plist` | Reports the macOS version. |
| `system/top.out` | `top -l 1` | Contains a process snapshot. |
| `system/ulimit.out` | `ulimit -a` | Reports the resource limits. |
| `system/vm_stat.out` | `vm_stat` | Reports virtual memory statistics. |
| `system/vm_stat_interval.out` | `vm_stat -c 10 1` | Reports virtual memory statistics over 10 samples. |

---

## Collecting PostgreSQL Instance Data

The PostgreSQL instance collectors gather instance-level data and store their
files in the `postgresql/` directory. The following table describes the files
that these collectors produce:

| File | Source | Description |
|------|--------|-------------|
| `postgresql/archiver.tsv` | `pg_stat_archiver` | Reports WAL archiver statistics. |
| `postgresql/available_extensions.tsv` | `pg_available_extension_versions` | Lists the available extension versions with the installed flag and the trust and requires metadata (PostgreSQL 13 and later). |
| `postgresql/bgwriter.tsv` | `pg_stat_bgwriter` | Reports background writer statistics. |
| `postgresql/blocking_locks.tsv` | Complex query | Lists blocking and blocked lock pairs. |
| `postgresql/checkpointer.tsv` | `pg_stat_checkpointer` | Reports checkpointer statistics (PostgreSQL 17 and later). |
| `postgresql/configuration.tsv` | `pg_settings` | Lists the configuration parameters. |
| `postgresql/connection_summary.tsv` | `pg_stat_activity` | Counts the backends by `state` and `wait_event_type`. |
| `postgresql/control_checkpoint.tsv` | `pg_control_checkpoint()` | Reports the latest checkpoint, with its LSNs, timeline, and next and oldest XIDs and multixacts. |
| `postgresql/control_init.tsv` | `pg_control_init()` | Reports the cluster initialisation constants, including the block and WAL segment size and whether data checksums are enabled. |
| `postgresql/control_recovery.tsv` | `pg_control_recovery()` | Reports the recovery state, with the minimum recovery end point and the backup start and end LSNs. |
| `postgresql/control_system.tsv` | `pg_control_system()` | Reports the system identifier and the control and catalog version numbers. |
| `postgresql/database_conflicts.tsv` | `pg_stat_database_conflicts` | Reports recovery conflict statistics. |
| `postgresql/database_sizes.tsv` | `pg_database_size()` | Reports database disk usage. |
| `postgresql/databases.tsv` | `pg_database` | Lists the databases. |
| `postgresql/databases_blk.tsv` | `pg_stat_database` | Reports block read and write statistics. |
| `postgresql/databases_checksums.tsv` | `pg_stat_database` | Reports checksum failure counts. |
| `postgresql/databases_tup.tsv` | `pg_stat_database` | Reports tuple operation statistics. |
| `postgresql/databases_xact.tsv` | `pg_stat_database` | Reports transaction commit and rollback counts. |
| `postgresql/db_role_setting.tsv` | `pg_db_role_setting` | Lists the per-database and per-role settings. |
| `postgresql/file_settings.tsv` | `pg_file_settings` | Reports the configuration file parse results and errors. |
| `postgresql/log_directory.tsv` | `pg_ls_logdir()` | Lists the server log directory with the name, size and modification time of each file. The listing includes file names only and never log contents. The listing is readable by `pg_monitor` and needs no filesystem access, so the collector works when radar runs as an OS user other than `postgres`. The file contains only a header, with no rows, when `logging_collector` is off. |
| `postgresql/pg_hba.conf` | Data directory | Contains the host-based authentication configuration. |
| `postgresql/pg_hba_file_rules.tsv` | `pg_hba_file_rules` | Lists the parsed `pg_hba.conf` rules (PostgreSQL 10 and later). |
| `postgresql/pg_ident.conf` | Data directory | Contains the user name mapping configuration. |
| `postgresql/postgresql.auto.conf` | Data directory | Contains the auto-generated configuration. |
| `postgresql/postgresql.conf` | Data directory | Contains the main configuration file. |
| `postgresql/postmaster_start_time.tsv` | `pg_postmaster_start_time()` | Reports the time at which the server started. |
| `postgresql/prepared_xacts.tsv` | `pg_prepared_xacts` | Lists the prepared transactions. |
| `postgresql/recovery.conf` | Data directory | Contains the recovery configuration (PostgreSQL 11 and earlier). |
| `postgresql/recovery.done` | Data directory | Contains the recovery completion marker. |
| `postgresql/replication.tsv` | `pg_stat_replication` | Reports the replication status. |
| `postgresql/replication_origin.tsv` | `pg_replication_origin_status` | Reports the replication origin status. |
| `postgresql/replication_slots.tsv` | `pg_replication_slots` | Lists the replication slots. |
| `postgresql/roles.tsv` | `pg_roles` | Lists the database roles. |
| `postgresql/running_activity.tsv` | `pg_stat_activity` | Lists every backend in `pg_stat_activity` without a filter, including idle and background backends. |
| `postgresql/running_activity_maxage.tsv` | Complex query | Reports one row with the maximum query, transaction, backend and lock wait ages (`max_query_age`, `max_xact_age`, `max_backend_age` and `max_lock_wait_age`) across backends whose `state` is not `idle`. The file contains no query text. |
| `postgresql/running_locks.tsv` | `pg_locks WHERE granted` | Lists the held locks. |
| `postgresql/shmem_allocations.tsv` | `pg_shmem_allocations` | Reports the shared memory breakdown (PostgreSQL 13 and later). |
| `postgresql/stat_io.tsv` | `pg_stat_io` | Reports I/O statistics (PostgreSQL 16 and later). |
| `postgresql/stat_progress_analyze.tsv` | `pg_stat_progress_analyze` | Reports ANALYZE progress (PostgreSQL 13 and later). |
| `postgresql/stat_progress_basebackup.tsv` | `pg_stat_progress_basebackup` | Reports base backup progress (PostgreSQL 13 and later). |
| `postgresql/stat_progress_cluster.tsv` | `pg_stat_progress_cluster` | Reports CLUSTER and VACUUM FULL progress (PostgreSQL 12 and later). |
| `postgresql/stat_progress_copy.tsv` | `pg_stat_progress_copy` | Reports COPY progress (PostgreSQL 14 and later). |
| `postgresql/stat_progress_create_index.tsv` | `pg_stat_progress_create_index` | Reports CREATE INDEX progress (PostgreSQL 12 and later). |
| `postgresql/stat_progress_vacuum.tsv` | `pg_stat_progress_vacuum` | Reports VACUUM progress (PostgreSQL 9.6 and later). |
| `postgresql/stat_replication_slots.tsv` | `pg_stat_replication_slots` | Reports replication slot statistics, including spill counters (PostgreSQL 14 and later). |
| `postgresql/stat_slru.tsv` | `pg_stat_slru` | Reports simple least-recently-used (SLRU) cache statistics (PostgreSQL 13 and later). |
| `postgresql/stat_ssl.tsv` | `pg_stat_ssl` | Reports the SSL/TLS state of each backend. |
| `postgresql/stat_statements_calls.tsv` | `pg_stat_statements` | Lists the top 100 queries by call count (PostgreSQL 13 and later). |
| `postgresql/stat_statements_max_time.tsv` | `pg_stat_statements` | Lists the top 100 queries by maximum execution time (PostgreSQL 13 and later). |
| `postgresql/stat_statements_total_time.tsv` | `pg_stat_statements` | Lists the top 100 queries by total execution time (PostgreSQL 13 and later). |
| `postgresql/stat_subscription_stats.tsv` | `pg_stat_subscription_stats` | Reports the apply and sync error counts of each subscription (PostgreSQL 15 and later), and the per-kind conflict counters (PostgreSQL 18 and later). |
| `postgresql/stat_wal.tsv` | `pg_stat_wal` | Reports WAL statistics (PostgreSQL 14 and later). |
| `postgresql/subscriptions.tsv` | `pg_subscription` | Lists the logical replication subscriptions. |
| `postgresql/tablespace_sizes.tsv` | `pg_tablespace_size()` | Reports tablespace disk usage. |
| `postgresql/tablespaces.tsv` | `pg_tablespace` | Lists the tablespace definitions. |
| `postgresql/version.tsv` | `version()` | Reports the PostgreSQL version. |
| `postgresql/waits_sample.tsv` | `pg_stat_activity` | Lists the active wait events. |
| `postgresql/wal_position.tsv` | `pg_current_wal_lsn()` | Reports the current WAL write, insert and flush LSNs, and whether the server is in recovery. The last received and replayed LSNs and the last replay timestamp apply only during recovery, and are empty on a primary. |
| `postgresql/wal_receiver.tsv` | `pg_stat_wal_receiver` | Reports the WAL receiver status on a standby. |

---

## Collecting Per-Database Data

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

---

## Collecting pg_statviz Data

The pg_statviz collectors are optional; these collectors are available when a
database has the [pg_statviz](https://github.com/vyruss/pg_statviz) extension
installed. radar stores their files in `pg_statviz/{dbname}/`. The following
table describes the pg_statviz files and the table that each file comes from:

| File | Table | Description |
|------|-------|-------------|
| `blocking.tsv` | `pgstatviz.blocking` | Contains the blocking lock history with blocked and blocker counts, plus per-lock-type counts (JSONB, pg_statviz 1.2 and later). |
| `buf.tsv` | `pgstatviz.buf` | Contains buffer and checkpoint statistics. |
| `conf.tsv` | `pgstatviz.conf` | Contains configuration snapshots (JSONB). |
| `conn.tsv` | `pgstatviz.conn` | Contains connection statistics (JSONB). |
| `db.tsv` | `pgstatviz.db` | Contains database statistics. |
| `io.tsv` | `pgstatviz.io` | Contains I/O statistics (JSONB, PostgreSQL 16 and later). |
| `lock.tsv` | `pgstatviz.lock` | Contains lock statistics (JSONB). |
| `repl.tsv` | `pgstatviz.repl` | Contains replication statistics (JSONB). |
| `slru.tsv` | `pgstatviz.slru` | Contains SLRU cache statistics (JSONB). |
| `snapshots.tsv` | `pgstatviz.snapshots` | Contains the snapshot timestamps. |
| `wait.tsv` | `pgstatviz.wait` | Contains wait event statistics (JSONB). |
| `wal.tsv` | `pgstatviz.wal` | Contains WAL statistics (PostgreSQL 14 and later). |

---

## Collecting Spock Data

The Spock collectors are optional; these collectors are available when a
database has the [Spock](https://github.com/pgEdge/spock) extension installed.
radar stores their files in `spock/{dbname}/`.

Three Spock relations contain data that must never enter an archive, because an
archive routinely accompanies a support ticket. The `exception_log` relation
holds jsonb images of the conflicting rows, and the `resolutions` relation
holds the same row images as text. The `node_interface.if_dsn` column holds the
node connection string, including its password. The two tasks that read
`exception_log` and `resolutions` name their columns explicitly rather than
using `SELECT *`. radar does not collect `node_interface` at all.

The following table describes the Spock files and the relation that each file
comes from:

| File | Relation | Description |
|------|----------|-------------|
| `channel_summary_stats.tsv` | `spock.channel_summary_stats` | Reports per-subscription counters for rows applied, conflicts and delta-apply conflicts. The file contains aggregates only and no row data. |
| `exception_log.tsv` | `spock.exception_log` | Lists the last 1000 apply exceptions by retry time. The file excludes `local_tup`, `remote_old_tup` and `remote_new_tup`, which hold jsonb row images. The file keeps `error_message`, the primary error message with its SQLSTATE. Spock stores no `DETAIL`, so the file does not include the offending key of a constraint violation; however, an error whose primary message embeds a value (a failed type conversion, for example) would include that value. |
| `lag_tracker.tsv` | `spock.lag_tracker` | Reports the replication lag in bytes and time for each node pair. |
| `local_node.tsv` | `spock.local_node` | Identifies which node this database is. |
| `local_sync_status.tsv` | `spock.local_sync_status` | Reports the initial synchronization state of each relation. |
| `node.tsv` | `spock.node` | Lists the cluster nodes with their id, name, location and country. The file excludes `info`, which holds deployment-defined jsonb. |
| `pii.tsv` | `spock.pii` | Lists the columns registered as holding personally identifiable information. The file contains column names only and no values. |
| `progress.tsv` | `spock.progress` | Reports the apply progress of each origin, with the commit timestamp and LSNs. |
| `replication_set.tsv` | `spock.replication_set` | Lists the replication sets and the operations that each set replicates. |
| `replication_set_table.tsv` | `spock.replication_set_table` | Lists the tables in each replication set, with the column list and row filter. |
| `resolutions.tsv` | `spock.resolutions` | Lists the last 1000 conflict resolutions by log time. The file excludes `local_tuple` and `remote_tuple`, which hold text row images. |
| `subscription.tsv` | `spock.subscription` | Lists the subscriptions with their enabled state, slot name, replication sets and apply delay. |
| `tables.tsv` | `spock.tables` | Lists the user tables and the replication set, if any, that each table belongs to. |

---

## Collecting PgBouncer Configuration

radar collects the [PgBouncer](https://www.pgbouncer.org/) configuration from
the first of `/etc/pgbouncer/` and `/usr/local/etc/pgbouncer/` that exists.
radar does not fall back to the second directory when the first directory lacks
a file. When neither directory exists, radar treats PgBouncer as not installed
and skips these collectors. radar uses a directory named with the
`--pgbouncer-conf` option as given, and treats a wrong path as a failure rather
than a skip. These collectors need no database connection, so radar collects
the files even when the PostgreSQL instance is unreachable.

radar never collects the contents of `userlist.txt`. The file holds
credentials, so `files.tsv` records only that `userlist.txt` is present as a
regular file.

The following table describes the PgBouncer files:

| File | Source | Description |
|------|--------|-------------|
| `pgbouncer/pgbouncer.ini` | `pgbouncer.ini` | Contains the `[pgbouncer]` section without its comments, including the pool mode, connection limits, listen address, auth type and auth file. In every other section, radar keeps each key name and replaces its value with `<redacted by radar>`, because `[databases]` and `[peers]` connection strings routinely contain `password=`. radar drops comments entirely, wherever they appear, because a commented-out entry holds a password as readily as a live entry. radar keeps an `%include` line but does not follow the file that the line names. radar reads the file from the `--pgbouncer-conf` directory when given, and otherwise from the first of `/etc/pgbouncer/` and `/usr/local/etc/pgbouncer/` that exists, without falling back to the second. |
| `pgbouncer/files.tsv` | filesystem | Lists the regular files in the configuration directory with the name, size and modification time of each file. This listing is how radar records `userlist.txt`. radar leaves out symbolic links and subdirectories, so the listing does not record a `userlist.txt` that is a symbolic link. |
