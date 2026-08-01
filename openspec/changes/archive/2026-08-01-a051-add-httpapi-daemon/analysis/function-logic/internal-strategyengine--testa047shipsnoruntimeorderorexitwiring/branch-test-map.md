# Branch Test Map: `TestA047ShipsNoRuntimeOrderOrExitWiring`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | walk callback propagates filesystem error | wiring guard | existing behavior | GREEN |
| B2 | directory entries are handled separately | wiring guard | existing behavior | GREEN |
| B3 | `.git` and vendor directories are skipped | wiring guard | existing behavior | GREEN |
| B4 | non-Go and test files are skipped | wiring guard | existing behavior | GREEN |
| B5 | strategyengine package files do not inspect themselves | wiring guard | existing behavior | GREEN |
| B6 | parse errors fail the walk | wiring guard | existing behavior | GREEN |
| B7 | every import is inspected | wiring guard | existing behavior | GREEN |
| B8 | strategyengine imports enter the dormant allowlist check | wiring guard | RED exposed HTTP import | GREEN |
| B9 | unlisted HTTP API strategyengine import is detected | wiring guard | RED: `internal/httpapi/read.go` unlisted | GREEN: explicit dormant descriptor entry |
| B10 | console/HTTP adapter may call only `DormantRuntimeDescriptor` | wiring guard selector assertion | assertion added | GREEN |
| B11 | outer walk error is fatal | wiring guard | existing behavior | GREEN |
