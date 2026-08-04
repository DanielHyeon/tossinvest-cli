//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package scheduler

import (
	"errors"
	"os"
)

func productionActivationOwnerUID() (uint64, bool) { return 0, false }

func readProductionActivationFile(string, uint64, os.FileMode, int64) ([]byte, error) {
	return nil, errors.New("scheduler: production activation ownership is unsupported on this platform")
}
