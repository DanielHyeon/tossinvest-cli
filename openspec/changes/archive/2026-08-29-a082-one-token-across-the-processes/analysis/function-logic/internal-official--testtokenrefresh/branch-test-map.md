# Branch Test Map: `TestTokenRefresh`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 갱신이 실패하거나 기대와 다른 토큰을 준다 | 자기 자신 | no (기존 통과) | yes |
| B2 | 채택할 것이 없는데 채택했다고 보고한다 | 자기 자신 | **yes** (M4: 채택 조건을 넓히면 여기서 걸린다) | yes |
| B3 | 교환이 2회가 아니다 | 자기 자신 | **yes** (M4) | yes |
