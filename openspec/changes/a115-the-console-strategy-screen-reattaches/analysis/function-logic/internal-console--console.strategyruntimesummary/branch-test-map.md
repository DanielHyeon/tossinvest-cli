# Branch Test Map: `Console.strategyRuntimeSummary`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 부재 신호 false → dormant 요약 | `TestTheSummaryAsksThePresenceSignalToo` · `TestStrategyRuntimeSummaryUsesPairedDormantTruth` | no | no |
| B2 | 읽기 실패 요약 | `TestStrategyRuntimeSummaryReportsReadFailure` | no | no |
