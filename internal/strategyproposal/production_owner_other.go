//go:build !unix

package strategyproposal

import (
	"os"
)

func productionOwnerUID() (uint32, bool)           { return 0, false }
func productionFileUID(os.FileInfo) (uint32, bool) { return 0, false }
func readProductionFile(string, uint32, os.FileMode, int64) ([]byte, error) {
	return nil, ErrProductionProposalUnavailable
}
