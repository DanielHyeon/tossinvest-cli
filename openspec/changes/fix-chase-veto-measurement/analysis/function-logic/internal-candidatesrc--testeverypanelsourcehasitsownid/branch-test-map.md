# Branch Test Map: `TestEveryPanelSourceHasItsOwnID`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 시장 | 자체 실행 | yes (컴파일) | yes |
| B2 | 패널 순회 | 자체 실행 | — (기존 동작) | yes |
| B3 | id가 서로 다르다 | 자체 실행 | — (기존 동작) | yes |
| B4 | 패널이 비지 않는다 — `Panel` B3의 도달 불가 arm을 지킨다 | 자체 실행 | — (기존 동작) | yes |
