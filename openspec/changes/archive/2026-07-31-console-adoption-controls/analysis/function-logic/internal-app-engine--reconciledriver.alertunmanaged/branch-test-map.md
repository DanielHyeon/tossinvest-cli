# Branch Test Map: `ReconcileDriver.alertUnmanaged`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST if 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |
| B2 | AST switch 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |
| B3 | AST case 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |
| B4 | AST case 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |
| B5 | AST case 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |
| B6 | AST case 분기 — 위 목적·경계 서술의 일부 | TestAFailedIncludeCycleSaysTriedNotOff(RED: 'adoption is off' 오표기)·TestIncludeDoesNotBypassExclusion·TestTheUnmanagedAlertSurvivesAdoptionBeingOff green | 해당(아래 기록) | yes |

RED: config/engine 신규 테스트 4건 실패 관측(2026-07-27, 구현 전 — adopted=0·사유 미표기·'off' 오표기·audit 항목 부재), CSRF 가드 목록 불일치 실패 관측.
GREEN: `go test ./internal/config/ ./internal/app/engine/ ./internal/console/ ./cmd/tossctl/ -race -count=1` — 629 passed (2026-07-27).
