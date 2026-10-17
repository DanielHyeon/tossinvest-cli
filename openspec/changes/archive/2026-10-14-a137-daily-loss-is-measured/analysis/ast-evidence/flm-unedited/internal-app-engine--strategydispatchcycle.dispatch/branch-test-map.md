# Branch Test Map: `strategyDispatchCycle.dispatch`

편집 전 좌표. 편집 뒤 재번호(옛/새 ast.json 정렬).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B10 | trigger 일일 손실 모드 투영 거절 → 계좌 진입 차단 결과 · 다른 trigger · 기동 복원 실패 · RECONCILE latch 거절은 지금처럼 오류(effective 잠금 포함) | `TestA137ADailyLossModeBlockEndsTheCycleQuietly`(새) · `TestA137OtherModeBlocksBehaveAsBefore`(새, 기동 복원 실패 포함) · `TestA137OtherEntryGateRefusalsStillFail`(새) | no | no |
| 220 | 강화와 경합한 펜스: trigger 일일 손실 모드 변경 → 계좌 진입 차단 결과 · 다른 펜스 원인은 지금처럼 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(새) · `TestA137OtherFencesBehaveAsBefore`(새) | no | no |
| (전달 루프) | 앞서 건너뛴 범위 거절 + 한도 도달 → 강화 · 나머지 전달 멈춤 · `errors.Join(skipped...)` 보고 유지 | `TestA137AMixedCycleStillReportsItsSkippedScopes`(새) | no | no |
| B12 | `DAILY_LOSS_LIMIT` → 계좌 진입 차단 결과, 임대 · 제출 0 | `TestA137ADailyLossStopEndsTheCycleWithoutAnError`(새) | no | no |
| B13 | precheck 거절의 원인이 `%w` 로 보존 | `TestA137APrecheckRefusalKeepsItsCause`(새) | no | no |
| 그 밖 | 무변경 | 기존 dispatch 시험 | — | — |
