# Branch Test Map: `TestAChurningSymbolIsMeasuredAgainstWhatWeActuallySaw`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 저장된 직전 읽기가 있으면 측정된다 | 자체 실행 | yes | yes |
| B2 | `NewlyListed`가 측정된 yes | 자체 실행 | yes | yes |
| B3 | gain이 실제 2칸 | 자체 실행 | yes (기대값이 1.333333333333에서 바뀌었다) | yes |
