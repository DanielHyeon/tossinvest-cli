//go:build unix && !linux

package strategyrouter

import "os"

func productionRouteFileUID(os.FileInfo) (uint32, bool) { return 0, false }
