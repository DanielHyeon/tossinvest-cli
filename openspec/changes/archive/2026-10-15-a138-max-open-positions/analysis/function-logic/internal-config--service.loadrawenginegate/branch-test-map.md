# Branch Test Map: `LoadRawEngineGate`

- Source: `internal/config/limits_io.go` (42-78)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 44 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B2 | if 47 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B3 | if 56 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B4 | if 74 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
