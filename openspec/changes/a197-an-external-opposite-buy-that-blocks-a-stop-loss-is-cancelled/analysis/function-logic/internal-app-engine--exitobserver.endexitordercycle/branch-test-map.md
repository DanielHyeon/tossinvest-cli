# Branch Test Map: `ExitObserver.endExitOrderCycle`

**편집 전 좌표**(AST 4 분기, base `b93ed133a`, `exit_order_log.go` `235`–`255`). 실행 0 — 시험 이름은 rg 로 골랐고 덮음 여부는 미측정(실행하지 않았다).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전용 로거 없음 → 전부 건너뜀 | `TestA129FWithoutTheDedicatedLoggerNothingChanges` · `TestA198WithoutTheDedicatedLoggerTheCancelQueueStaysEmpty` | no | no |
| B2 | 대기열 줄 쓰기 | `TestA129AConfirmedStopAndItsFillsEachLeaveALine` · `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` | no | no |
| B3 | 래치 순회 | `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode` | no | no |
| B4 | 이번 주기 거절 없음 → 래치 삭제 | `TestA129MTheRefusalLatchDoesNotOutliveItsEpisode` | no | no |
