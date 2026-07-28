# Branch Test Map: `TestTheRankTotalIsTheListWeActuallyReceived`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 생성 실패 | 자체 실행 | yes (컴파일) | yes |
| B2 | 읽기 실패 | 자체 실행 | yes | yes |
| B3 | 행 순회 | 자체 실행 | — (기존 동작) | yes |
| B4 | `RankTotal`은 도착한 3 | 자체 실행 | — (기존 동작) | yes |

같은 fixture(100 요청 3 도착)에 대해 `RankRequested == 100`을 확인하는 것은
`TestTheRequestedRowCountTravelsBesideTheOneThatArrived`다 — 이 change가 더한 절반이 그쪽에
있고, 이 테스트는 기존 절반을 지킨다.
