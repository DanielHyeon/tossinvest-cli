# Branch Test Map: `TestAnUnmanagedHoldingIsLabelledExactlyOnce`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 라벨 3회 미만이면 실패 | 자체 실행 (전체 스위트 green) | yes (구현 전 신규 문자열 부재로 실패) | yes |
| B2 | 필수 문자열 순회 | 자체 실행 | yes | yes |
| B3 | 필수 문자열 부재 검출 | 자체 실행 | yes | yes |
| B4 | 금지 문자열 순회 | 자체 실행 | — | yes |
| B5 | 금지 문자열 검출 | 자체 실행 | — | yes |
