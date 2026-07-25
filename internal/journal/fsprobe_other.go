//go:build !linux

package journal

import "fmt"

// SystemFSProber returns a prober that refuses to guess.
//
// The engine's journal is Linux-targeted: statfs(2) is what lets us tell a local
// journaling filesystem from a network or FUSE mount. On any other platform we
// cannot make that distinction, so the guard fails closed instead of assuming the
// durability contract holds. The CLI itself keeps working on those platforms — it
// does not use the journal.
func SystemFSProber() FSProber { return unsupportedProber{} }

type unsupportedProber struct{}

func (unsupportedProber) Probe(dir string) (FSInfo, error) {
	return FSInfo{}, fmt.Errorf("%w (inspecting %s)", ErrProbeUnsupported, dir)
}
