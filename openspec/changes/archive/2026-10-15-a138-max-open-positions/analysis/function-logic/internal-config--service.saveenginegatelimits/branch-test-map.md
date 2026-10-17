# Branch Test Map: `SaveEngineGateLimits`

- Source: `internal/config/limits_io.go` (99-168)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 103 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B2 | if 107 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B3 | if 113 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B4 | switch 119 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B5 | case 120 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B6 | if 132 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B7 | if 136 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B8 | if 140 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B9 | case 144 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B10 | if 148 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B11 | if 157 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B12 | else 159 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B13 | if 159 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B14 | if 164 | 분기 동작 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
