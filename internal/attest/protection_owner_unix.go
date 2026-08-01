//go:build unix

package attest

import "os"

func currentProtectionOwnerUID() (uint32, bool) {
	uid := os.Geteuid()
	if uid < 0 {
		return 0, false
	}
	return uint32(uid), true
}
