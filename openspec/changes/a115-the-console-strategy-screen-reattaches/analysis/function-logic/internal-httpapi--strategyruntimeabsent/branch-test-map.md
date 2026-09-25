# Branch Test Map: `StrategyRuntimeAbsent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil → 부재 | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` | no | no |
| B2 | presence 를 말하는 reader → 그 답의 부정 | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` | no | no |
