# Branch Test Map: `Panel`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 공식 클라이언트가 없으면 WTS만 남는다 | `TestThePanelWithoutAnOfficialClientIsNotSilentlyWTSOnly` | — (기존 동작) | yes |
| B2 | 세 ranking type이 각자 소스가 된다 | `TestEveryPanelSourceHasItsOwnID` | — (기존 동작) | yes |
| B3 | 세 type 모두 id가 있어 하나도 버려지지 않는다 | `TestEveryPanelSourceHasItsOwnID`(빈 패널·중복 id 모두 실패) | — (기존 동작) | yes |
| B4 | US 패널에는 인기 순위가 없다 | `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking` | — (기존 동작) | yes |
| (꼬리) | `clk`가 네 소스 전부에 전달된다 | `internal/candidatesrc/reading_validity_test.go` 전체가 주입 clock으로 시간을 몬다 | yes | yes |

`Panel`을 **주입 clock으로** 부르는 테스트는 없다 — 유효성 테스트들은 두 생성자를 직접
부른다. `Panel`이 `clk`를 떨어뜨려도(둘 다 nil로 넘겨도) 컴파일되고 전 테스트가 통과한다.
정직하게 기록한다. 생산 호출부(`cmd/tossctl/candidatepanel.go`)는 `clock.System()`을 넘기므로
동작상 같지만, **전달 자체를 잡는 테스트는 없다.**
