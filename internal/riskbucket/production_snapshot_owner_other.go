//go:build !unix

package riskbucket

import (
	"os"
)

func productionRiskOwnerUID() (uint32, bool) { return 0, false }
func readProductionRiskFile(string, uint32, os.FileMode, int64) ([]byte, error) {
	return nil, ErrProductionRiskSnapshotUnavailable
}
func validateProductionRiskJournalFile(string, uint32) error {
	return ErrProductionRiskSnapshotUnavailable
}
