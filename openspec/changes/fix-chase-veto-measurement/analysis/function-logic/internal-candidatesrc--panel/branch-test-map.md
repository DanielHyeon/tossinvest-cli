# Branch Test Map: `Panel`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 공식 클라이언트가 없으면 WTS만 남는다 | `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly` | — (기존 동작) | yes |
| B2 | 세 ranking type이 각자 소스가 된다 | `TestEveryPanelSourceHasItsOwnID` | — (기존 동작) | yes |
| B3 | 세 type 모두 id가 있어 하나도 버려지지 않는다 | `TestEveryPanelSourceHasItsOwnID`(빈 패널·중복 id 모두 실패) | — (기존 동작) | yes |
| B4 | US 패널에는 인기 순위가 없다 | `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking` | — (기존 동작) | yes |
| (꼬리) | `clk`가 네 소스 전부에 전달된다 | `TestThePanelHandsItsClockToEverySourceItBuilds` — 네 소스를 각각 온전한 읽기 → 1 tick → `previousReadingTTL` 경과로 몬다 | yes (2026-07-28: 두 생성자 중 어느 쪽에 `nil`을 넣어도 해당 subtest가 실패) | yes |

**2026-07-28 해소**. 이전 판은 "`Panel`이 `clk`를 떨어뜨려도 컴파일되고 전 테스트가
통과한다"고 정직하게 적어 두었고, pre-gate 리뷰가 그것을 P2로 올렸다.
`clock_wiring_test.go`가 그 자리를 막는다. 짧은 읽기는 신선도 이전에 온전 판정에서 걸리므로
fixture는 **정확히 100행·30행**을 서브한다 — 그렇지 않으면 시계를 전혀 갖지 않은 소스에
대해서도 통과한다.
