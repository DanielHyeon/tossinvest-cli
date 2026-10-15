# Branch Test Map: `DailyBuyNotionalLimit`

- Source: `internal/app/engine/interlock_test.go` (56-61)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 57 | 분기 동작 | `go test ./internal/app/engine -run 'Interlock|A136|A138'` · `TestA136TheInterlockRefusesAGuardianThatDisagreesOnTheDailyBuyCeiling` |
