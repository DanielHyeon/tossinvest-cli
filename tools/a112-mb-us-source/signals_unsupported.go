//go:build !unix

package main

import "os"

func measurementSignals() []os.Signal { return []os.Signal{os.Interrupt} }
