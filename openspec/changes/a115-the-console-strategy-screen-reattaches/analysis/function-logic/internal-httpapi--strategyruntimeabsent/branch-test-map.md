# Branch Test Map: `StrategyRuntimeAbsent`

구현 후 AST 기준(분기 0 — 위임 한 줄). 편집 전 B1·B2 는 `strategyprojection.StrategyRuntimeAbsent` 로 옮겨 갔다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 위임: nil·부재 wrapper → true, sentinel·live wrapper·신호 없는 reader → false (httpapi·strategyprojection 같은 답) | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` · `TestTheAbsenceIsJudgedInOnePlace` | yes — 변이 K5(신호 무시)·K13(사본 판정)·K14(인터페이스 두 벌) CAUGHT | yes |
