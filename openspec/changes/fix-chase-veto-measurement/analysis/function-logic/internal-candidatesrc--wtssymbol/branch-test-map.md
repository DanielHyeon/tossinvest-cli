# Branch Test Map: `wtsSymbol`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Symbol이 있으면 그것, 없으면 ProductCode — 그리고 **집합도 같은 문자열로** | `TestThePopularityRankingFallsBackToTheProductCode`(선택) · `TestAWTSRowIdentifiedByItsProductCodeIsNotANewEntrantEveryTime`(두 호출부의 일치) | yes (후자는 집합을 `s.Symbol`로 만들면 실패) | yes |
