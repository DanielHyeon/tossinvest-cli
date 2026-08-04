//go:build !unix

package strategyevidence

func evidenceOwnerUID() (uint32, bool)                  { return 0, false }
func validateEvidenceReadOnlyFile(string, uint32) error { return ErrSnapshotUnavailable }
