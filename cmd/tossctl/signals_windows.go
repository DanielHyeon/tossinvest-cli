//go:build windows

package main

import "os"

func consoleTerminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
