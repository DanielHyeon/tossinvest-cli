# Branch Test Map: `TestThePopularityRankingReportsNoTradingFigures`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기 실패 | 자체 실행 | yes (컴파일) | yes |
| B2 | 한 행 | 자체 실행 | — (기존 동작) | yes |
| B3 | 싣지 않는 수치는 비어 있다 | 자체 실행 | — (기존 동작) | yes |
| B4 | 위치는 실린다 | 자체 실행 | — (기존 동작) | yes |

이 fixture는 30을 요청해 1행이 도착하므로 **짧은 읽기**다. 그래서 이 change 뒤로는 기억을
교체하지 않는다 — 이 테스트가 재는 것과 무관하지만, `whole` 비교가 여기서도 돌고 있다는
사실은 기록해 둔다.
