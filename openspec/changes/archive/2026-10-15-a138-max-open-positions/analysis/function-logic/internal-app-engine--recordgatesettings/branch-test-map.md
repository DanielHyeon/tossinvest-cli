# Branch Test Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (734-803)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 737 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B2 | if 747 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B3 | else 750 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B4 | if 750 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B5 | range 797 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B6 | if 798 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
