# Branch Test Map: `signalsVetoTallyFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 통과 수가 0이 아니면 note가 바뀐다 | `internal/console/signals_test.go`의 `signalsPassedUnexpected` 케이스 | — (기존 동작) | yes |
| B2 | 모순이 있으면 경보가 차고 없으면 빈다 | `TestTheSignalsScreenSaysSoWhenTheTallyContradictsItself` · `TestTheOrdinarySignalsScreenRaisesNoAlarm` | yes | yes |
| B3 | code별 칸 | `signals_test.go` | — (기존 동작) | yes |
| B4 | 사유 census | 동상 | — (기존 동작) | yes |
