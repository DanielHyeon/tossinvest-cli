//go:build !unix

package main

// detachProcess has no portable equivalent outside Unix. The spawned survey keeps
// the parent's process group, which means it can be taken down with the console —
// worse than the Unix behaviour, and still better than not offering the restart.

import "os/exec"

func detachProcess(*exec.Cmd) {}
