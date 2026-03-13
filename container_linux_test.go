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
	"testing"
)

func TestIsContainer(t *testing.T) {
	_, err := os.Stat("/.dockerenv")
	dockerExists := err == nil
	got := isContainer()
	if dockerExists && !got {
		t.Fatalf("isContainer() = false, want true when /.dockerenv exists")
	}
	if !dockerExists && got {
		t.Fatalf("isContainer() = true, want false on non-container host")
	}
}

func TestContainerCommandTasksStructure(t *testing.T) {
	for i, task := range containerCommandTasks {
		if task.Name == "" {
			t.Errorf("containerCommandTasks[%d] missing Name", i)
		}
		if task.ArchivePath == "" {
			t.Errorf("containerCommandTasks[%d] (%s) missing ArchivePath", i, task.Name)
		}
		if task.Command == "" {
			t.Errorf("containerCommandTasks[%d] (%s) missing Command", i, task.Name)
		}
	}
}

func TestContainerCommandTasksAlphabeticalOrder(t *testing.T) {
	for i := 1; i < len(containerCommandTasks); i++ {
		if containerCommandTasks[i].Name < containerCommandTasks[i-1].Name {
			t.Errorf("containerCommandTasks not alphabetically ordered: %q comes after %q",
				containerCommandTasks[i].Name, containerCommandTasks[i-1].Name)
		}
	}
}
