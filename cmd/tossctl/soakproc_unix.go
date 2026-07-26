//go:build unix

package main

// detachProcess puts the spawned survey in its own session, which is what
// soak-autostart.sh gets from `setsid`.
//
// Without it the child keeps the console's controlling terminal, and the Ctrl-C
// that stops the console — or the hangup when its terminal closes — takes the
// multi-day survey with it.

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
