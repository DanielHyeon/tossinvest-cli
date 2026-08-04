//go:build unix && !linux

package strategyevidence

import "os"

func evidenceFileUID(os.FileInfo) (uint32, bool) { return 0, false }
