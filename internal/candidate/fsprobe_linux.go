//go:build linux

package candidate

import (
	"fmt"
	"syscall"
)

// SystemFSProber returns the production prober, backed by statfs(2).
func SystemFSProber() FSProber { return statfsProber{} }

type statfsProber struct{}

func (statfsProber) Probe(dir string) (FSInfo, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(dir, &st); err != nil {
		return FSInfo{}, fmt.Errorf("statfs %s: %w", dir, err)
	}
	// Statfs_t.Type is int64 on 64-bit and int32 on 32-bit Linux; masking to 32
	// bits keeps a magic such as btrfs' 0x9123683E from sign-extending into a
	// negative int64 on the latter.
	magic := int64(st.Type) & 0xFFFFFFFF
	return FSInfo{Name: FilesystemName(magic), Magic: magic}, nil
}
