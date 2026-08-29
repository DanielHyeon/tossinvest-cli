# Branch Test Map: `signalsBandTalliesFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | code 순서대로 | `internal/console/signals_test.go`의 밴드 블록 | — (동작 무변경) | yes |
| B2 | 밴드가 없는 code는 블록이 없다 | 동상 | — (동작 무변경) | yes |
| B3 | 밴드 칸 | 동상 | — (동작 무변경) | yes |
| B4 | 밴드별 미측정 census | 동상 | — (동작 무변경) | yes |
