# Branch Test Map: `Observation.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 검증이 정규화보다 앞 | `TestARankWithoutItsListLengthIsRefused` | — (기존 동작) | yes |
| B2 | market 없는 관측 | **커버 없음** | no | no |
| B3 | symbol 없는 관측 | **커버 없음** | no | no |
| B4 | source 없는 관측 | **커버 없음** | no | no |
| B5 | instant 없는 관측 | **커버 없음** | no | no |
| B6 | 음수 rank | **커버 없음** | no | no |
| B7 | 음수 요청 행 수 | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` | yes (case를 `false &&`로 무력화하면 실패 확인) | yes |
| B8 | 목록 길이 없는 rank | `TestARankWithoutItsListLengthIsRefused` | — (기존 동작) | yes |
| B9 | 목록보다 큰 rank | **커버 없음** | no | no |

B7은 이 change가 만든 유일한 분기이고, evidence 생산 시점에는 **아무 테스트도 그것을
넘기지 않았다**. `truncation_test.go`의 표에 "a negative request, which validate refuses at
the boundary"라는 이름만 있고 그 주장을 확인하는 것이 없었다 — 그래서 이번에
`negative_request_test.go`를 추가했다.

**그 옆의 기존 공백도 기록한다**: 이 switch의 여덟 case 중 테스트가 넘기는 것은
`Rank > 0 && RankTotal == 0`(B8) 하나뿐이다. market·symbol·source·instant·음수 rank·
목록 초과 rank 여섯 case는 이 change 이전부터 커버되지 않았고, 이 change의 범위 밖이라
여기서는 기록만 한다.
