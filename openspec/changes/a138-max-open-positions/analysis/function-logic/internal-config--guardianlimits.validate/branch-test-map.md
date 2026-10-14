# Branch Test Map: `Validate`

- Source: `internal/config/limits.go` (329-359)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | range 330 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B2 | switch 332 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B3 | case 333 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B4 | case 335 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B5 | if 340 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B6 | if 344 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B7 | if 348 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B8 | if 352 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B9 | if 353 | 분기 동작 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
