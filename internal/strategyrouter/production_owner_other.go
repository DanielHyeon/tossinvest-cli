//go:build !unix

package strategyrouter

import "os"

func productionRouteOwnerUID() (uint32, bool) { return 0, false }
func readProductionRouteFile(string, uint32, os.FileMode, int64) ([]byte, error) {
	return nil, ErrProductionRouteUnavailable
}
func validateProductionRouteJournalFile(string, uint32) error { return ErrProductionRouteUnavailable }
