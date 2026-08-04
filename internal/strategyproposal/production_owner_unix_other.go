//go:build unix && !linux

package strategyproposal

import "os"

func productionFileUID(os.FileInfo) (uint32, bool) { return 0, false }
