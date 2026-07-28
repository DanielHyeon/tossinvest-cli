# Branch Test Map: `Console.handleDashboard`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 등록되지 않은 경로가 `/` 핸들러에 도달하면 404 | `TestTheConsoleServesNothingButItsThreePages` | — | yes |
| B2 | 검증 진행 중이면 run 뷰와 2초 재로드, 아니면 없음 | `TestTheApprovedFlowRunsExactlyTheApprovedBatch`, `TestTheDashboardReportsAnUnstartedMachineWithoutFailing`, `TestTheVerificationScreensKeepTheirTwoSecondReload` | — | yes |
