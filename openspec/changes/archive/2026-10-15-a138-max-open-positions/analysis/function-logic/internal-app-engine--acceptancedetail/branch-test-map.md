# Branch Test Map: `acceptanceDetail`

- Source: `internal/app/engine/interlock.go` (425-442)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | 직선(분기 없음) | 정상 경로 | `go test ./internal/app/engine -run 'Interlock|Acceptance|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
