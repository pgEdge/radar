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

import (
	"os"
	"strings"
)

// getContainerTasks returns container-specific collection tasks if running inside a container
func getContainerTasks() []CollectionTask {
	if !isContainer() {
		return nil
	}
	return buildCommandTasks("system", containerCommandTasks)
}

// isContainer returns true if radar is running inside a container
func isContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := strings.ToLower(string(data))
		for _, sig := range []string{"docker", "kubepods", "containerd", "lxc"} {
			if strings.Contains(content, sig) {
				return true
			}
		}
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// System command tasks (sorted alphabetically by name)
var systemCommandTasks = []SimpleCommandTask{
	{
		Name:        "dmesg-t",
		ArchivePath: "system/dmesg_t.out",
		Command:     "dmesg",
		Args:        []string{"-T"},
	},
	{
		Name:        "free",
		ArchivePath: "system/free.out",
		Command:     "free",
		Args:        []string{"-h"},
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
		Name:        "lscpu",
		ArchivePath: "system/lscpu.out",
		Command:     "lscpu",
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
		Name:        "mpstat",
		ArchivePath: "system/mpstat.out",
		Command:     "mpstat",
		Args:        []string{"-P", "ALL", "1", "5"},
	},
	{
		Name:        "netstat-stats",
		ArchivePath: "system/netstat_stats.out",
		Command:     "netstat",
		Args:        []string{"-s"},
	},
	{
		Name:        "nfsiostat",
		ArchivePath: "system/nfsiostat.out",
		Command:     "nfsiostat",
		Args:        []string{},
	},
	{
		Name:        "numactl",
		ArchivePath: "system/numactl.out",
		Command:     "numactl",
		Args:        []string{"--hardware"},
	},
	{
		Name:        "numastat",
		ArchivePath: "system/numastat.out",
		Command:     "numastat",
		Args:        []string{"-m"},
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
		Name:        "openssl-fips-mode-setup",
		ArchivePath: "system/openssl/fips-mode-setup.out",
		Command:     "fips-mode-setup",
		Args:        []string{"--check"},
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
		Name:        "pg-service-status",
		ArchivePath: "system/systemd/postgresql-status.out",
		Command:     "sh",
		Args:        []string{"-c", "systemctl status 'postgresql*' 2>/dev/null || systemctl status 'postgres*' 2>/dev/null"},
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
		Name:        "ss-listeners",
		ArchivePath: "system/ss_listeners.out",
		Command:     "ss",
		Args:        []string{"-tunlp"},
	},
	{
		Name:        "ss-summary",
		ArchivePath: "system/ss_summary.out",
		Command:     "ss",
		Args:        []string{"-s"},
	},
	{
		Name:        "systemctl-list-units",
		ArchivePath: "system/systemd/list-units.out",
		Command:     "systemctl",
		Args:        []string{"list-units", "--all"},
	},
	{
		Name:        "timedatectl",
		ArchivePath: "system/timedatectl.out",
		Command:     "timedatectl",
		Args:        []string{"status"},
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
		Name:        "vmstat-command",
		ArchivePath: "system/vmstat-command.out",
		Command:     "vmstat",
		Args:        []string{"1", "10"},
	},
	{
		Name:        "clocksource",
		ArchivePath: "system/sys/clocksource.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /sys/devices/system/clocksource/clocksource0/current_clocksource 2>/dev/null"},
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
		Name:        "io-queue-depth",
		ArchivePath: "system/io_queue_depth.out",
		Command:     "sh",
		Args:        []string{"-c", "for f in /sys/block/*/queue/nr_requests; do [ -f \"$f\" ] && echo \"$(basename $(dirname $(dirname $f))): $(cat $f)\"; done"},
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
		Name:        "transparent_hugepage",
		ArchivePath: "system/sys/kernel_mm_transparent_hugepage.out",
		Command:     "sh",
		Args:        []string{"-c", "grep -r . /sys/kernel/mm/transparent_hugepage/ 2>/dev/null"},
	},
}

// System file read tasks (sorted alphabetically by name)
var systemFileTasks = []SimpleFileTask{
	{
		Name:        "cgroup-cpu-max",
		ArchivePath: "system/cgroup/cpu_max.out",
		Path:        "/sys/fs/cgroup/cpu.max",
	},
	{
		Name:        "cgroup-cpu-weight",
		ArchivePath: "system/cgroup/cpu_weight.out",
		Path:        "/sys/fs/cgroup/cpu.weight",
	},
	{
		Name:        "cgroup-cpuset-cpus",
		ArchivePath: "system/cgroup/cpuset_cpus_effective.out",
		Path:        "/sys/fs/cgroup/cpuset.cpus.effective",
	},
	{
		Name:        "cgroup-io-max",
		ArchivePath: "system/cgroup/io_max.out",
		Path:        "/sys/fs/cgroup/io.max",
	},
	{
		Name:        "cgroup-memory-current",
		ArchivePath: "system/cgroup/memory_current.out",
		Path:        "/sys/fs/cgroup/memory.current",
	},
	{
		Name:        "cgroup-memory-max",
		ArchivePath: "system/cgroup/memory_max.out",
		Path:        "/sys/fs/cgroup/memory.max",
	},
	{
		Name:        "cgroup-memory-stat",
		ArchivePath: "system/cgroup/memory_stat.out",
		Path:        "/sys/fs/cgroup/memory.stat",
	},
	{
		Name:        "cgroup-memory-swap-max",
		ArchivePath: "system/cgroup/memory_swap_max.out",
		Path:        "/sys/fs/cgroup/memory.swap.max",
	},
	{
		Name:        "cgroup-pids-current",
		ArchivePath: "system/cgroup/pids_current.out",
		Path:        "/sys/fs/cgroup/pids.current",
	},
	{
		Name:        "cgroup-pids-max",
		ArchivePath: "system/cgroup/pids_max.out",
		Path:        "/sys/fs/cgroup/pids.max",
	},
	{
		Name:        "cgroup-v1-cpu-cfs-period",
		ArchivePath: "system/cgroup-v1/cpu_cfs_period_us.out",
		Path:        "/sys/fs/cgroup/cpu/cpu.cfs_period_us",
	},
	{
		Name:        "cgroup-v1-cpu-cfs-quota",
		ArchivePath: "system/cgroup-v1/cpu_cfs_quota_us.out",
		Path:        "/sys/fs/cgroup/cpu/cpu.cfs_quota_us",
	},
	{
		Name:        "cgroup-v1-cpu-shares",
		ArchivePath: "system/cgroup-v1/cpu_shares.out",
		Path:        "/sys/fs/cgroup/cpu/cpu.shares",
	},
	{
		Name:        "cgroup-v1-cpuset-cpus",
		ArchivePath: "system/cgroup-v1/cpuset_cpus.out",
		Path:        "/sys/fs/cgroup/cpuset/cpuset.cpus",
	},
	{
		Name:        "cgroup-v1-memory-limit",
		ArchivePath: "system/cgroup-v1/memory_limit_in_bytes.out",
		Path:        "/sys/fs/cgroup/memory/memory.limit_in_bytes",
	},
	{
		Name:        "cgroup-v1-memory-stat",
		ArchivePath: "system/cgroup-v1/memory_stat.out",
		Path:        "/sys/fs/cgroup/memory/memory.stat",
	},
	{
		Name:        "cgroup-v1-memory-usage",
		ArchivePath: "system/cgroup-v1/memory_usage_in_bytes.out",
		Path:        "/sys/fs/cgroup/memory/memory.usage_in_bytes",
	},
	{
		Name:        "cloud-bios-vendor",
		ArchivePath: "system/cloud/bios_vendor.out",
		Path:        "/sys/class/dmi/id/bios_vendor",
	},
	{
		Name:        "cloud-chassis-asset-tag",
		ArchivePath: "system/cloud/chassis_asset_tag.out",
		Path:        "/sys/class/dmi/id/chassis_asset_tag",
	},
	{
		Name:        "cloud-product-name",
		ArchivePath: "system/cloud/product_name.out",
		Path:        "/sys/class/dmi/id/product_name",
	},
	{
		Name:        "cloud-sys-vendor",
		ArchivePath: "system/cloud/sys_vendor.out",
		Path:        "/sys/class/dmi/id/sys_vendor",
	},
	{
		Name:        "container-cgroup-membership",
		ArchivePath: "system/container/cgroup_membership.out",
		Path:        "/proc/1/cgroup",
	},
	{
		Name:        "container-mountinfo",
		ArchivePath: "system/container/mountinfo.out",
		Path:        "/proc/1/mountinfo",
	},
	{
		Name:        "cpuinfo",
		ArchivePath: "system/proc/cpuinfo.out",
		Path:        "/proc/cpuinfo",
	},
	{
		Name:        "diskstats",
		ArchivePath: "system/proc/diskstats.out",
		Path:        "/proc/diskstats",
	},
	{
		Name:        "fstab",
		ArchivePath: "system/fstab.out",
		Path:        "/etc/fstab",
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

// Container-only command tasks (only included when isContainer() returns true)
var containerCommandTasks = []SimpleCommandTask{
	{
		Name:        "container-env",
		ArchivePath: "system/container/environment.out",
		Command:     "sh",
		Args:        []string{"-c", "env | grep -E '^(HOSTNAME|CONTAINER_ID|DOCKER_HOST|ECS_CLUSTER|ECS_CONTAINER_METADATA_URI|KUBERNETES_SERVICE_HOST|KUBERNETES_SERVICE_PORT|KUBERNETES_PORT)=' | sort || true"},
	},
	{
		Name:        "container-k8s-namespace",
		ArchivePath: "system/container/k8s_namespace.out",
		Command:     "sh",
		Args:        []string{"-c", "cat /run/secrets/kubernetes.io/serviceaccount/namespace 2>/dev/null || true"},
	},
}
