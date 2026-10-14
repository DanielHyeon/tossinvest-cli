# Branch Test Map: `strategyDispatchCycle.dispatch`

편집 전 좌표. 편집 뒤 재번호(옛/새 ast.json 정렬).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B10 | 운영 모드 투영 거절 → 계좌 진입 차단 결과(오류 아님) · RECONCILE latch 거절은 지금처럼 오류 | `TestA137AModeProjectionRefusalEndsTheCycleQuietly`(새) · `TestA137OtherEntryGateRefusalsStillFail`(새) | no | no |
| B12 | `DAILY_LOSS_LIMIT` → 계좌 진입 차단 결과, 임대 · 제출 0 | `TestA137ADailyLossStopEndsTheCycleWithoutAnError`(새) | no | no |
| B13 | precheck 거절의 원인이 `%w` 로 보존 | `TestA137APrecheckRefusalKeepsItsCause`(새) | no | no |
| 그 밖 | 무변경 | 기존 dispatch 시험 | — | — |
