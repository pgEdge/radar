//go:build darwin

/*-------------------------------------------------------------------------
 *
 * radar
 *
 * Portions copyright (c) 2026, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package main

// macOS-specific command tasks (sorted alphabetically by name)
var systemCommandTasks = []SimpleCommandTask{
	{
		Name:        "brew-list",
		ArchivePath: "system/packages_brew.out",
		Command:     "brew",
		Args:        []string{"list", "--versions"},
	},
	{
		Name:        "brew-postgres",
		ArchivePath: "system/packages_brew_postgres.out",
		Command:     "sh",
		Args:        []string{"-c", "brew list --versions | grep -i postgres"},
	},
	{
		Name:        "diskutil-info-all",
		ArchivePath: "system/diskutil_info_all.out",
		Command:     "sh",
		Args:        []string{"-c", "diskutil list | grep -o '/dev/disk[0-9]*' | xargs -n1 diskutil info"},
	},
	{
		Name:        "diskutil-list",
		ArchivePath: "system/diskutil_list.out",
		Command:     "diskutil",
		Args:        []string{"list"},
	},
	{
		Name:        "hostname",
		ArchivePath: "system/hostname.out",
		Command:     "hostname",
		Args:        []string{},
	},
	{
		Name:        "hypervisor-check",
		ArchivePath: "system/hypervisor.out",
		Command:     "sh",
		Args:        []string{"-c", "sysctl kern.hv_vmm_present machdep.cpu.features | grep -i 'hypervisor\\|vmx\\|svm'"},
	},
	{
		Name:        "ifconfig",
		ArchivePath: "system/ifconfig.out",
		Command:     "ifconfig",
		Args:        []string{"-a"},
	},
	{
		Name:        "iostat",
		ArchivePath: "system/iostat.out",
		Command:     "iostat",
		Args:        []string{"-c", "5", "-w", "1"},
	},
	{
		Name:        "ipcs",
		ArchivePath: "system/ipcs.out",
		Command:     "ipcs",
		Args:        []string{"-a"},
	},
	{
		Name:        "kextstat",
		ArchivePath: "system/kextstat.out",
		Command:     "kextstat",
		Args:        []string{},
	},
	{
		Name:        "launchctl-list",
		ArchivePath: "system/launchctl_list.out",
		Command:     "launchctl",
		Args:        []string{"list"},
	},
	{
		Name:        "memory-pressure",
		ArchivePath: "system/memory_pressure.out",
		Command:     "memory_pressure",
		Args:        []string{},
	},
	{
		Name:        "netstat-interfaces",
		ArchivePath: "system/netstat_interfaces.out",
		Command:     "netstat",
		Args:        []string{"-i"},
	},
	{
		Name:        "netstat-routing",
		ArchivePath: "system/netstat_routing.out",
		Command:     "netstat",
		Args:        []string{"-r"},
	},
	{
		Name:        "netstat-stats",
		ArchivePath: "system/netstat_stats.out",
		Command:     "netstat",
		Args:        []string{"-s"},
	},
	{
		Name:        "pmset-assertions",
		ArchivePath: "system/pmset_assertions.out",
		Command:     "pmset",
		Args:        []string{"-g", "assertions"},
	},
	{
		Name:        "pmset-settings",
		ArchivePath: "system/pmset_settings.out",
		Command:     "pmset",
		Args:        []string{"-g"},
	},
	{
		Name:        "sysctl-cpu",
		ArchivePath: "system/sysctl_cpu.out",
		Command:     "sysctl",
		Args:        []string{"-a", "machdep.cpu"},
	},
	{
		Name:        "sysctl-hw",
		ArchivePath: "system/sysctl_hw.out",
		Command:     "sysctl",
		Args:        []string{"-a", "hw"},
	},
	{
		Name:        "sysctl-kern",
		ArchivePath: "system/sysctl_kern.out",
		Command:     "sysctl",
		Args:        []string{"-a", "kern"},
	},
	{
		Name:        "sysctl-vm",
		ArchivePath: "system/sysctl_vm.out",
		Command:     "sysctl",
		Args:        []string{"-a", "vm"},
	},
	{
		Name:        "system-log-boot",
		ArchivePath: "system/system_log_boot.out",
		Command:     "log",
		Args:        []string{"show", "--predicate", "processID == 0", "--last", "boot", "--style", "syslog"},
	},
	{
		Name:        "system-profiler-hardware",
		ArchivePath: "system/system_profiler_hardware.out",
		Command:     "system_profiler",
		Args:        []string{"SPHardwareDataType"},
	},
	{
		Name:        "system-profiler-network",
		ArchivePath: "system/system_profiler_network.out",
		Command:     "system_profiler",
		Args:        []string{"SPNetworkDataType"},
	},
	{
		Name:        "system-profiler-pci",
		ArchivePath: "system/system_profiler_pci.out",
		Command:     "system_profiler",
		Args:        []string{"SPPCIDataType"},
	},
	{
		Name:        "system-profiler-software",
		ArchivePath: "system/system_profiler_software.out",
		Command:     "system_profiler",
		Args:        []string{"SPSoftwareDataType"},
	},
	{
		Name:        "system-profiler-storage",
		ArchivePath: "system/system_profiler_storage.out",
		Command:     "system_profiler",
		Args:        []string{"SPStorageDataType"},
	},
	{
		Name:        "top",
		ArchivePath: "system/top.out",
		Command:     "top",
		Args:        []string{"-l", "1"},
	},
	{
		Name:        "ulimit",
		ArchivePath: "system/ulimit.out",
		Command:     "sh",
		Args:        []string{"-c", "ulimit -a"},
	},
	{
		Name:        "vm-stat",
		ArchivePath: "system/vm_stat.out",
		Command:     "vm_stat",
		Args:        []string{},
	},
	{
		Name:        "vm-stat-interval",
		ArchivePath: "system/vm_stat_interval.out",
		Command:     "vm_stat",
		Args:        []string{"-c", "10", "1"},
	},
}

// macOS-specific file tasks (sorted alphabetically by name)
var systemFileTasks = []SimpleFileTask{
	{
		Name:        "sysctl-conf",
		ArchivePath: "system/sysctl.conf",
		Path:        "/etc/sysctl.conf",
	},
	{
		Name:        "system-version",
		ArchivePath: "system/system_version.plist",
		Path:        "/System/Library/CoreServices/SystemVersion.plist",
	},
}
