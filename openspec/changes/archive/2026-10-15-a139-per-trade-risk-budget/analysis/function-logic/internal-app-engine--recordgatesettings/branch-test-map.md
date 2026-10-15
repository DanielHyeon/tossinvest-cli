# Branch Test Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (739-811)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 742 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
| B2 | if 752 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
| B3 | else 755 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
| B4 | if 755 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
| B5 | range 805 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
| B6 | if 806 | 분기 동작 | `go test ./internal/app/engine -run 'Audit|A074|A139'` · `TestA139TheInterlockAuditsTheDerivedRiskBudget` · `TestA139AGateOffStartWithoutADailyLossStillStarts` |
