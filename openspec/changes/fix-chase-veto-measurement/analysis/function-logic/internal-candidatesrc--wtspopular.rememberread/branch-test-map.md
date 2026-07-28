# Branch Test Map: `wtsPopular.rememberRead`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행마다 집합에 넣는다 | `TestTheWTSPopularityListReportsTheSameTwoFacts` | yes | yes |
| B2 | 식별자가 빈 행은 집합에 넣지 않는다 | **커버 없음** | no | no |
| B3 | 첫 호출에서 맵을 만든다 | `TestTheWTSPopularityListReportsTheSameTwoFacts` | yes | yes |
| B4 | 1행 읽기(2 요청)는 기억을 교체하지 않는다 | `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes (F2) | yes |
| B5 | TTL 경과 후 답이 미상으로 돌아간다 | `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes (F1) | yes |

**정직한 커버리지 기록**: B2는 커버되지 않는다 — `wtsSymbol`이 빈 문자열을 돌려주는
fixture가 없다. `wtsPopular.Read` B4와 짝인 공백이다.
