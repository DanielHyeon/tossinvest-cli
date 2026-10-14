# Branch Test Map: `RiskGuardian.PrecheckQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go` (89-177, 편집 뒤). 편집 전 분기수 17 → 편집 뒤 17. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 90 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B2 | if 93 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B3 | if 99 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B4 | if 103 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B5 | if 106 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B6 | if 120 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B7 | if 124 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B8 | if 128 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B9 | if 134 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B10 | if 137 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B11 | if 142 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B12 | else 144 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B13 | if 147 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B14 | if 151 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B15 | if 156 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
| B16 | if 168 | a137 편집 | `TestA137ThePrecheckDoorRefusesAMeasuredLossAtTheLimit` · `TestA137ThePrecheckDoorRefusesAnUnmeasurableDayAsTyped` · `TestA137ThePrecheckDoorConvertsAUSLossOnlyWithThePairedAuthority` · `TestA137ThePrecheckDoorReportsNoCapitalAsTheDailyLossLimit` · `TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead` · `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` · `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` | 편집 분기 |
| B17 | if 171 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` | 무편집 분기 |
