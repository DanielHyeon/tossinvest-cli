# Branch Test Map: `gateMembersOf`

- Source: `internal/config/limits_io.go` (196-218)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 198 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B2 | if 209 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B3 | if 213 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
