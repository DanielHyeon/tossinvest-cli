# Branch Test Map: `wtsPopular.rememberRead`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 행마다 집합에 넣는다 | `TestTheWTSPopularityListReportsTheSameTwoFacts` | yes | yes |
| B2 | 식별자가 빈 행은 집합에 넣지 않고, 그래서 그 읽기는 온전하지 않다 | `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps` | yes (2026-07-28) | yes |
| B3 | 첫 호출에서 맵을 만든다 | `TestTheWTSPopularityListReportsTheSameTwoFacts` | yes | yes |
| B4 | 1행 읽기(2 요청)는 기억을 교체하지 않는다 | `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes (F2) | yes |
| B5 | TTL 경과 후 답이 미상으로 돌아간다 | `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes (F1) | yes |

**2026-07-28 해소**: `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps`가 `wtsSymbol`이 빈
문자열을 돌려주는 행을 만든다. `wtsPopular.Read` B4와 짝인 공백이었고 함께 닫혔다.
