//go:build !unix

package protectionreadiness

import "os"

func currentOwnerUID() (uint32, bool)      { return 0, false }
func fileOwner(os.FileInfo) (uint32, bool) { return 0, false }
func readOwnedFile(string, uint32, os.FileMode) (os.FileInfo, []byte, error) {
	return nil, nil, os.ErrPermission
}
func acquireStateLock(string, uint32) (func(), bool, func() error, error) {
	return nil, false, nil, os.ErrPermission
}
