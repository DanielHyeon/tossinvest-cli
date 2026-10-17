# Branch Test Map: `strategyRuntimeAttachment.StrategyRuntimeConfigured`

인용 전용 — 기존 a109 시험 · a115 시험이 덮는다(편집 없음).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 자리 상태를 읽고 **무조건 wake** 한 뒤 `reader != nil` 을 답한다(a109 G2) | `TestTheAbsenceJudgementIsOneForBothPackages` · `TestTheAbsenceSignalIsOneJudgementForEveryConsumer` · `TestAnUnconfiguredWrapperStillRendersDormant`(콘솔 쪽 판정) | no — 무편집 | yes |
