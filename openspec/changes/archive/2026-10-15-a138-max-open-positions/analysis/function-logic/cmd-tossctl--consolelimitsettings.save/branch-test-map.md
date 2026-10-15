# Branch Test Map: `Save`

- Source: `cmd/tossctl/limitsettings.go` (49-78)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 51 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B2 | if 57 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B3 | else 59 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B4 | if 64 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B5 | if 66 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B6 | if 72 | 분기 동작 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
