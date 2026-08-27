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

// getSystemTasks returns system-level collection tasks
func getSystemTasks() []CollectionTask {
	// Build tasks from registries
	tasks := []CollectionTask{}

	// Add platform-specific tasks (linux or darwin build tag)
	tasks = append(tasks, buildCommandTasks("system", systemCommandTasks)...)
	tasks = append(tasks, buildFileTasks("system", systemFileTasks)...)

	// Add shared cross-platform tasks
	tasks = append(tasks, buildCommandTasks("system", sharedCommandTasks)...)
	tasks = append(tasks, buildFileTasks("system", sharedFileTasks)...)

	// Add container-specific tasks when running inside a container
	tasks = append(tasks, getContainerTasks()...)

	// Add PgBouncer tasks, which skip cleanly when it is not installed
	tasks = append(tasks, getPgBouncerTasks()...)

	return tasks
}
