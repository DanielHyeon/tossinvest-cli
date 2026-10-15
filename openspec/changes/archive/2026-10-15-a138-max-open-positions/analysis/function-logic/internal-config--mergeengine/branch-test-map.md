# Branch Test Map: `mergeEngine`

- Source: `internal/config/engine.go` (324-358)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 325 | 분기 동작 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B2 | if 328 | 분기 동작 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B3 | if 332 | 분기 동작 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B4 | if 342 | 분기 동작 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B5 | if 346 | 분기 동작 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
