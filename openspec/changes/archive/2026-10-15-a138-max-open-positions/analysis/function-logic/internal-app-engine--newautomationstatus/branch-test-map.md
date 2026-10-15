# Branch Test Map: `newAutomationStatus`

- Source: `internal/app/engine/interlock.go` (328-341)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/app/engine -run 'Interlock|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
