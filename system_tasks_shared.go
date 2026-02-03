//go:build linux || darwin

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

// SimpleCommandTask defines a shell command-based collection
type SimpleCommandTask struct {
	Name        string
	ArchivePath string
	Command     string
	Args        []string
}

// SimpleFileTask defines a file read-based collection
type SimpleFileTask struct {
	Name        string
	ArchivePath string
	Path        string
}

// buildCommandTasks converts SimpleCommandTask registry to CollectionTask slice
func buildCommandTasks(category string, tasks []SimpleCommandTask) []CollectionTask {
	result := make([]CollectionTask, len(tasks))
	for i, t := range tasks {
		result[i] = CollectionTask{
			Category:    category,
			Name:        t.Name,
			ArchivePath: t.ArchivePath,
			Collector:   execCommandCollector(t.Command, t.Args...),
		}
	}
	return result
}

// buildFileTasks converts SimpleFileTask registry to CollectionTask slice
func buildFileTasks(category string, tasks []SimpleFileTask) []CollectionTask {
	result := make([]CollectionTask, len(tasks))
	for i, t := range tasks {
		result[i] = CollectionTask{
			Category:    category,
			Name:        t.Name,
			ArchivePath: t.ArchivePath,
			Collector:   readFileCollector(t.Path),
		}
	}
	return result
}

// Cross-platform command tasks (work on both Linux and macOS)
var sharedCommandTasks = []SimpleCommandTask{
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
		Name:        "mount",
		ArchivePath: "system/mount.out",
		Command:     "mount",
		Args:        []string{},
	},
	{
		Name:        "openssl-ciphers",
		ArchivePath: "system/openssl/ciphers.out",
		Command:     "openssl",
		Args:        []string{"ciphers"},
	},
	{
		Name:        "openssl-engines",
		ArchivePath: "system/openssl/engines.out",
		Command:     "openssl",
		Args:        []string{"engine"},
	},
	{
		Name:        "openssl-version",
		ArchivePath: "system/openssl/version.out",
		Command:     "openssl",
		Args:        []string{"version", "-a"},
	},
	{
		Name:        "ps",
		ArchivePath: "system/ps.out",
		Command:     "ps",
		Args:        []string{"auxww"},
	},
	{
		Name:        "uname",
		ArchivePath: "system/uname.out",
		Command:     "uname",
		Args:        []string{"-a"},
	},
	{
		Name:        "sysctl",
		ArchivePath: "system/sysctl.out",
		Command:     "sysctl",
		Args:        []string{"-a"},
	},
}

// Cross-platform file read tasks
var sharedFileTasks = []SimpleFileTask{
	{
		Name:        "hosts",
		ArchivePath: "system/hosts.out",
		Path:        "/etc/hosts",
	},
}
