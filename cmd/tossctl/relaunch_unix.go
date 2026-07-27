//go:build unix

package main

// relaunch_unix.go is the one place in this binary that replaces this process with
// itself (openspec change verify-execution-capability, task 1.8 ①).
//
// syscall.Exec, not a fork: the console has already closed its listener by the time
// this runs, and an exec keeps the terminal, the standard streams and the job
// control the operator started the console with. A forked child would either
// inherit them and outlive its parent's shell in a way nobody asked for, or lose
// them and take the printed session URL with it.
//
// The PID survives an exec, which is the whole reason internal/verifylive judges the
// conditional-persistence boundary on process.instance_id — a fresh random token per
// startup — instead. See internal/console/restart.go.

import (
	"fmt"
	"os"
	"syscall"
)

// reexecSelf replaces this process with argv, run from path.
//
// It does not return on success.
func reexecSelf(path string, argv []string) error {
	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		return fmt.Errorf("재실행 실패 (%s): %w", path, err)
	}
	// Unreachable: a successful Exec never comes back.
	return nil
}
