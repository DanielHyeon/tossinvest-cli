//go:build unix && !linux

package riskbucket

import "os"

func productionRiskFileUID(os.FileInfo) (uint32, bool) { return 0, false }
