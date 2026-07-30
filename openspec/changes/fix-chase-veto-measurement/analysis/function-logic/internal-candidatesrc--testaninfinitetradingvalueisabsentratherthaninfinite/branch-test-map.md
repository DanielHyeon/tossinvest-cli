# Branch Test Map: `TestAnInfiniteTradingValueIsAbsentRatherThanInfinite`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 생성 실패 | 자체 실행 | yes (컴파일) | yes |
| B2 | 읽기 실패 | 자체 실행 | yes | yes |
| B3 | 무한값은 부재 | 자체 실행 | — (기존 동작) | yes |
| B4 | NaN은 부재 | 자체 실행 | — (기존 동작) | yes |
| B5 | 유한값은 보존 | 자체 실행 | — (기존 동작) | yes |
