//go:build !unix

package main

// relaunch_other.go is the honest refusal on platforms with no exec.
//
// The console's restart button is drawn from the presence of the seam, and the seam
// is present here — so the button exists and says why it cannot work, rather than
// disappearing and leaving an operator wondering which build they are on. Windows
// would need a spawn-and-exit, and that is a different design (the successor has to
// wait for the port, and the terminal ownership changes); it is not written because
// nothing this change measures runs there.

import (
	"errors"
	"fmt"
)

func reexecSelf(path string, _ []string) error {
	return fmt.Errorf("이 플랫폼에서는 자기 재실행을 지원하지 않는다 (%s): %w", path, errors.ErrUnsupported)
}
