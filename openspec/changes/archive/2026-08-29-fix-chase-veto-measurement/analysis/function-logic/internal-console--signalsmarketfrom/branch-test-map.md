# Branch Test Map: `signalsMarketFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽지 못한 시장은 사유만 렌더 | `TestEveryUnmeasuredStateOnTheSignalsScreenCarriesACodeAndASentence` | — (기존 동작) | yes |
| B2 | 후보 행 | `signals_test.go` 전반 | — (기존 동작) | yes |
| (신규 1줄) | census 블록이 조립된다 | `TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem` · `TestThePerSourceBlockIsAbsentRatherThanEmptyWhenNothingHasASighting` | yes | yes |
