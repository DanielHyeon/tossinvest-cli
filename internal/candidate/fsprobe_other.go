//go:build !linux

package candidate

import "fmt"

// SystemFSProber returns a prober that refuses to guess.
//
// Same reasoning as the ledger's: statfs(2) is what tells a local journaling
// filesystem from a network or FUSE mount, and without that distinction the guard
// fails closed rather than assuming a durability contract it cannot check. The
// rest of the CLI keeps working on such a platform — it does not open this store.
func SystemFSProber() FSProber { return unsupportedProber{} }

type unsupportedProber struct{}

func (unsupportedProber) Probe(dir string) (FSInfo, error) {
	return FSInfo{}, fmt.Errorf("%w (inspecting %s)", ErrProbeUnsupported, dir)
}
