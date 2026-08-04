//go:build !unix

package officialfx

import (
	"os"
)

func productionOwnerUID() (uint32, bool) { return 0, false }
func readProductionFile(string, uint32, os.FileMode) (os.FileInfo, []byte, error) {
	return nil, nil, os.ErrPermission
}
func acquireProductionStateLock(string, uint32) (func(), bool, func() error, error) {
	return nil, false, nil, os.ErrPermission
}
