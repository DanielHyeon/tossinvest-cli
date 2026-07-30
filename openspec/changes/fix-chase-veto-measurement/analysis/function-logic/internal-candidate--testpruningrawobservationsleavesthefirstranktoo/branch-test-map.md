# Branch Test Map: `TestPruningRawObservationsLeavesTheFirstRankToo`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 관측 기록 | 자체 실행 | yes (컴파일) | yes |
| B2 | 승격 | 자체 실행 | yes | yes |
| B3 | 최초 순위 기록 | 자체 실행 | yes | yes |
| B4 | 정리 | 자체 실행 | yes | yes |
| B5 | 정리 후 읽기 | 자체 실행 | yes | yes |
| B6 | 원시 행이 사라졌다 | 자체 실행 | yes | yes |
| B7 | 최초 순위 읽기 | 자체 실행 | yes | yes |
| B8 | 최초 순위는 남는다 | 자체 실행 | yes | yes |
| B9 | 남은 위치로 여전히 측정된다 | 자체 실행 | yes (자격이 zero면 미측정이 되어 실패) | yes |
