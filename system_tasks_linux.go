//go:build linux

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

// System command tasks (sorted alphabetically by name)
var systemCommandTasks = []SimpleCommandTask{
	{
		Name:        "df",
		ArchivePath: "system/diskspace.out",
		Command:     "df",
		Args:        []string{"-h"},
	},
	{
		Name:        "dmesg",
		ArchivePath: "system/dmesg.out",
		Command:     "dmesg",
		Args:        []string{},
	},
	{
		Name:        "dmesg-t",
		ArchivePath: "system/dmesg_t.out",
		Command:     "dmesg",
		Args:        []string{"-T"},
	},
	{
		Name:        "hostname",
		ArchivePath: "system/hostname.out",
		Command:     "hostname",
		Args:        []string{"-f"},
	},
	{
		Name:        "hypervisor",
		ArchivePath: "system/hypervisor.out",
		Command:     "systemd-detect-virt",
		Args:        []string{},
	},
	{
		Name:        "ifconfig",
		ArchivePath: "system/ifconfig.out",
		Command:     "ifconfig",
		Args:        []string{"-a"},
	},
	{
		Name:        "interfaces",
		ArchivePath: "system/interfaces.out",
		Command:     "ip",
		Args:        []string{"-o", "address"},
	},
	{
		Name:        "iostat",
		ArchivePath: "system/iostat.out",
		Command:     "iostat",
		Args:        []string{"-x", "1", "5"},
	},
	{
		Name:        "ip-addr",
		ArchivePath: "system/ip_addr.out",
		Command:     "ip",
		Args:        []string{"address", "list"},
	},
	{
		Name:        "ipcs",
		ArchivePath: "system/ipcs.out",
		Command:     "ipcs",
		Args:        []string{"-a"},
	},
	{
		Name:        "locale",
		ArchivePath: "system/locale.out",
		Command:     "locale",
		Args:        []string{},
	},
	{
		Name:        "locale-all",
		ArchivePath: "system/locale_all.out",
		Command:     "locale",
		Args:        []string{"-a"},
	},
	{
		Name:        "localectl",
		ArchivePath: "system/localectl.out",
		Command:     "localectl",
		Args:        []string{"status"},
	},
	{
		Name:        "lsblk",
		ArchivePath: "system/lsblk.out",
		Command:     "lsblk",
		Args:        []string{},
	},
	{
		Name:        "lsdevmapper",
		ArchivePath: "system/lsdevmapper.out",
		Command:     "ls",
		Args:        []string{"-la", "/dev/mapper"},
	},
	{
		Name:        "lsmod",
		ArchivePath: "system/lsmod.out",
		Command:     "lsmod",
		Args:        []string{},
	},
	{
		Name:        "lspci",
		ArchivePath: "system/lspci.out",
		Command:     "lspci",
		Args:        []string{},
	},
	{
		Name:        "mount",
		ArchivePath: "system/mount.out",
		Command:     "mount",
		Args:        []string{},
	},
	{
		Name:        "mpstat",
		ArchivePath: "system/mpstat.out",
		Command:     "mpstat",
		Args:        []string{"-P", "ALL", "1", "5"},
	},
	{
		Name:        "nfsiostat",
		ArchivePath: "system/nfsiostat.out",
		Command:     "nfsiostat",
		Args:        []string{},
	},
	{
		Name:        "openssl-ciphers",
		ArchivePath: "system/openssl/ciphers.out",
		Command:     "openssl",
		Args:        []string{"ciphers"},
	},
	{
		Name:        "openssl-crypto-policies-isapplied",
		ArchivePath: "system/openssl/crypto-policies-isapplied.out",
		Command:     "update-crypto-policies",
		Args:        []string{"--is-applied"},
	},
	{
		Name:        "openssl-crypto-policies-show",
		ArchivePath: "system/openssl/crypto-policies-show.out",
		Command:     "update-crypto-policies",
		Args:        []string{"--show"},
	},
	{
		Name:        "openssl-engines",
		ArchivePath: "system/openssl/engines.out",
		Command:     "openssl",
		Args:        []string{"engine"},
	},
	{
		Name:        "openssl-fips-mode-setup",
		ArchivePath: "system/openssl/fips-mode-setup.out",
		Command:     "fips-mode-setup",
		Args:        []string{"--check"},
	},
	{
		Name:        "openssl-version",
		ArchivePath: "system/openssl/version.out",
		Command:     "openssl",
		Args:        []string{"version", "-a"},
	},
	{
		Name:        "packages-apt-list-installed",
		ArchivePath: "system/packages-apt-list-installed.out",
		Command:     "apt",
		Args:        []string{"list", "--installed", "*postgres*"},
	},
	{
		Name:        "packages-dnf-list-installed",
		ArchivePath: "system/packages-dnf-list-installed.out",
		Command:     "dnf",
		Args:        []string{"list", "installed", "*postgres*"},
	},
	{
		Name:        "packages-dpkg",
		ArchivePath: "system/packages-dpkg.out",
		Command:     "dpkg",
		Args:        []string{"-l", "*postgres*"},
	},
	{
		Name:        "packages-rpm",
		ArchivePath: "system/packages-rpm.out",
		Command:     "rpm",
		Args:        []string{"-qa", "*postgres*"},
	},
	{
		Name:        "packages-yum-list-installed",
		ArchivePath: "system/packages-yum-list-installed.out",
		Command:     "yum",
		Args:        []string{"list", "installed", "*postgres*"},
	},
	{
		Name:        "ps",
		ArchivePath: "system/ps.out",
		Command:     "ps",
		Args:        []string{"auxww"},
	},
	{
		Name:        "sar",
		ArchivePath: "system/sar.out",
		Command:     "sar",
		Args:        []string{"-A"},
	},
	{
		Name:        "sestatus",
		ArchivePath: "system/sestatus.out",
		Command:     "sestatus",
		Args:        []string{},
	},
	{
		Name:        "systemctl-list-units",
		ArchivePath: "system/systemd/list-units.out",
		Command:     "systemctl",
		Args:        []string{"list-units", "--all"},
	},
	{
		Name:        "top",
		ArchivePath: "system/top.out",
		Command:     "top",
		Args:        []string{"-b", "-c", "-w", "512", "-n", "1"},
	},
	{
		Name:        "tuned-active",
		ArchivePath: "system/tuned/tuned-active.out",
		Command:     "tuned-adm",
		Args:        []string{"active"},
	},
	{
		Name:        "tuned-list",
		ArchivePath: "system/tuned/tuned-list.out",
		Command:     "tuned-adm",
		Args:        []string{"list"},
	},
	{
		Name:        "uname",
		ArchivePath: "system/uname.out",
		Command:     "uname",
		Args:        []string{"-a"},
	},
	{
		Name:        "vmstat-command",
		ArchivePath: "system/vmstat-command.out",
		Command:     "vmstat",
		Args:        []string{"1", "10"},
	},
	{
		Name:        "cpu_scaling_available_governors",
		ArchivePath: "system/sys/cpu_scaling_available_governors.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_available_governors 2>/dev/null | sort -u"},
	},
	{
		Name:        "cpu_scaling_driver",
		ArchivePath: "system/sys/cpu_scaling_driver.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_driver 2>/dev/null | sort -u"},
	},
	{
		Name:        "cpu_scaling_governor",
		ArchivePath: "system/sys/cpu_scaling_governor.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor 2>/dev/null | sort -u"},
	},
	{
		Name:        "energy_perf_bias",
		ArchivePath: "system/sys/energy_perf_bias.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/cpu/cpu*/power/energy_perf_bias 2>/dev/null | sort -u"},
	},
	{
		Name:        "intel_pstate",
		ArchivePath: "system/sys/intel_pstate.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/cpu/intel_pstate/* 2>/dev/null"},
	},
	{
		Name:        "io-schedulers",
		ArchivePath: "system/io_schedulers.out",
		Command:     "sh",
		Args:        []string{"-c", "for f in /sys/block/*/queue/scheduler; do [ -f \"$f\" ] && echo \"$(basename $(dirname $(dirname $f))): $(cat $f)\"; done"},
	},
	{
		Name:        "read_ahead",
		ArchivePath: "system/read_ahead.out",
		Command:     "sh",
		Args:        []string{"-c", "blockdev --getra /dev/sd* /dev/nvme* 2>/dev/null"},
	},
	{
		Name:        "sysctl",
		ArchivePath: "system/sysctl.out",
		Command:     "sysctl",
		Args:        []string{"-a"},
	},
	{
		Name:        "transparent_hugepage",
		ArchivePath: "system/sys/kernel_mm_transparent_hugepage.out",
		Command:     "sh",
		Args:        []string{"-c", "grep -r . /sys/kernel/mm/transparent_hugepage/ 2>/dev/null"},
	},
}

// System file read tasks (sorted alphabetically by name)
var systemFileTasks = []SimpleFileTask{
	{
		Name:        "cpuinfo",
		ArchivePath: "system/proc/cpuinfo.out",
		Path:        "/proc/cpuinfo",
	},
	{
		Name:        "fstab",
		ArchivePath: "system/fstab.out",
		Path:        "/etc/fstab",
	},
	{
		Name:        "hosts",
		ArchivePath: "system/hosts.out",
		Path:        "/etc/hosts",
	},
	{
		Name:        "limits",
		ArchivePath: "system/limits.out",
		Path:        "/etc/security/limits.conf",
	},
	{
		Name:        "locale-conf",
		ArchivePath: "system/locale_conf.out",
		Path:        "/etc/locale.conf",
	},
	{
		Name:        "machine-id",
		ArchivePath: "system/machine_id.out",
		Path:        "/etc/machine-id",
	},
	{
		Name:        "meminfo",
		ArchivePath: "system/proc/meminfo.out",
		Path:        "/proc/meminfo",
	},
	{
		Name:        "os-release",
		ArchivePath: "system/os_release.out",
		Path:        "/etc/os-release",
	},
	{
		Name:        "pressure-cpu",
		ArchivePath: "system/proc/pressure_cpu.out",
		Path:        "/proc/pressure/cpu",
	},
	{
		Name:        "pressure-io",
		ArchivePath: "system/proc/pressure_io.out",
		Path:        "/proc/pressure/io",
	},
	{
		Name:        "pressure-memory",
		ArchivePath: "system/proc/pressure_memory.out",
		Path:        "/proc/pressure/memory",
	},
	{
		Name:        "proc-loadavg",
		ArchivePath: "system/proc/loadavg.out",
		Path:        "/proc/loadavg",
	},
	{
		Name:        "proc-mounts",
		ArchivePath: "system/proc/mounts.out",
		Path:        "/proc/mounts",
	},
	{
		Name:        "proc-uptime",
		ArchivePath: "system/proc/uptime.out",
		Path:        "/proc/uptime",
	},
	{
		Name:        "proc-vmstat",
		ArchivePath: "system/proc/vmstat.out",
		Path:        "/proc/vmstat",
	},
	{
		Name:        "swaps",
		ArchivePath: "system/proc/swaps.out",
		Path:        "/proc/swaps",
	},
	{
		Name:        "system-release",
		ArchivePath: "system/system_release.out",
		Path:        "/etc/system-release",
	},
}
