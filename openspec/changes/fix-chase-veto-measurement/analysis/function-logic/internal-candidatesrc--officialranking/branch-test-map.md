# Branch Test Map: `OfficialRanking`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록되지 않은 ranking type | `TestAnUnknownRankingTypeIsRefused` | — (기존 동작) | yes |
| B2 | 500 요청 → 엔드포인트에 100, `RankRequested`도 100 | `TestTheRequestedCountIsCappedAtTheDocumentedMaximum` · `TestTheRequestedCountIsTheCappedOneRatherThanTheOneTheCallerAsked` | yes (후자는 이 change의 RED — 상한 전 값을 실으면 실패) | yes |
| B3 | nil clock이 시스템 시계가 된다 | `TestTheOfficialRankingAsksForTheRealtimeList` 외 12건 | — (신규 인자) | yes |

`count <= 0` 반쪽은 어떤 테스트도 넘기지 않는다 — 생산 호출부는 `Panel`의 리터럴 100뿐이다.
`count > 100` 반쪽만 측정되어 있고, 두 반쪽은 같은 대입으로 합류하므로 관측된 결과는 같다.
정직하게 기록한다: **0 이하를 넘기는 테스트는 없다.**
