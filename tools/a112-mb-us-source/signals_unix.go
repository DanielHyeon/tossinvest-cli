//go:build unix

package main

import (
	"os"
	"syscall"
)

func measurementSignals() []os.Signal { return []os.Signal{os.Interrupt, syscall.SIGTERM} }
