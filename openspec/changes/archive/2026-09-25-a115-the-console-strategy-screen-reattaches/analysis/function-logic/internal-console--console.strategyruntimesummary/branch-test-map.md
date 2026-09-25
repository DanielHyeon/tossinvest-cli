# Branch Test Map: `Console.strategyRuntimeSummary`

구현 후 AST 기준(분기 2, 번호 불변 — B1 조건만 교체).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 부재 신호 false → dormant 요약(Read 0) | `TestTheSummaryAsksThePresenceSignalToo` · `TestStrategyRuntimeSummaryUsesPairedDormantTruth` | yes — base 에서 요약이 「KR ON→ON …」(판정 갈림) · 변이 K4·K5 CAUGHT | yes |
| B2 | 읽기 실패 요약 | `TestStrategyRuntimeSummaryReportsReadFailure` · `TestTheSummaryAsksThePresenceSignalToo` | no — 무변경 분기(회귀 핀) | yes |
