# Collecting Linux System Data

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

## Reading Cgroup v2 Resource Limits

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

## Reading Cgroup v1 Resource Limits

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

## Identifying the Cloud Provider and Hardware

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

## Identifying the Container Runtime

radar attempts the container identity collectors on Linux and skips any
collector whose source is unavailable. These files help identify the container
runtime and environment. The following table describes the container identity
files:

| File | Source | Description |
|------|--------|-------------|
| `system/container/cgroup_membership.out` | `/proc/1/cgroup` | Reports the cgroup membership (container signatures). |
| `system/container/mountinfo.out` | `/proc/1/mountinfo` | Reports the PID 1 mount information (overlay detection). |

## Detecting a Container Environment

radar detects a container environment automatically on Linux. radar collects
these files only when it runs inside a container (Docker, Kubernetes, LXC or
containerd). The following table describes the container detection files:

| File | Source | Description |
|------|--------|-------------|
| `system/container/environment.out` | `env \| grep` | Records the allowlisted environment variables `HOSTNAME`, `CONTAINER_ID`, `DOCKER_HOST`, `ECS_CLUSTER`, `ECS_CONTAINER_METADATA_URI`, `KUBERNETES_SERVICE_HOST`, `KUBERNETES_SERVICE_PORT` and `KUBERNETES_PORT`. |
| `system/container/k8s_namespace.out` | `/run/secrets/.../namespace` | Reports the Kubernetes namespace. |
