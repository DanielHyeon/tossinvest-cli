//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package candidate

import "os"

func productionThresholdOwnerUID() (uint64, bool) { return 0, false }

func readProductionThresholdFile(string, uint64, os.FileMode, int64) ([]byte, error) {
	return nil, ErrProductionThresholdAuthorityUnavailable
}
