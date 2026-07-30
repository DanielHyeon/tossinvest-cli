//go:build unix

package main

import (
	"os"
	"syscall"
)

func consoleTerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
