# Branch Test Map: `TestANewlyListedSymbolDoesNotClimbFromLastPlace`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 상위 진입과 경계 진입 | 자체 실행 | yes (100/98로 바뀌기 전 fixture는 150/148이었다) | yes |
| B2 | gain이 계산되지 않는다 | 자체 실행 | yes | yes |
| B3 | 사유가 `NO_PRIOR_RANK` | 자체 실행 | yes | yes |
| B4 | gain은 빈 문자열 | 자체 실행 | yes | yes |
| B5 | `NewlyListed`가 **측정된 yes** | 자체 실행 | yes (`.Yes()`가 아니라 `!bool`이면 미상도 통과했다) | yes |
| B6 | 위치가 보존된다 | 자체 실행 | yes | yes |

`panelsize_drift_test.go:TestNoCommentClaimsAPanelSizeTheSourcesDoNotDeclare`가 이
테스트의 **주석**까지 검사한다 — 150행 주장을 다시 적으면 실패한다.
