# Collecting macOS System Data

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
