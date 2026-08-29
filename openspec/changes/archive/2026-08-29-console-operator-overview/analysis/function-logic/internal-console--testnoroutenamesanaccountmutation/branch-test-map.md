# Branch Test Map: `TestNoRouteNamesAnAccountMutation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 실제 19건 표에 발견이 없다 | 이 테스트 자신 + `TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist`(판정부의 positive control) | yes — 판정부가 항상 빈 목록을 돌려주면 이 테스트만으로는 통과하고, positive control이 그것을 잡는다 | yes |
