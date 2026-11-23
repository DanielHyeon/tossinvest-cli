# Branch Test Map: `ExitObserver.endExitOrderCycle` (pre-edit 번호)

시험 이름은 grep 근거(internal/app/engine/a129_exit_order_log_test.go). RED/GREEN 은 4d 원문에서 채움.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전용 로거 없음 | TestA129FWithoutTheDedicatedLoggerNothingChanges | n/a(기존) | 4d |
| B2 | a129 대기열 쓰기 | TestA129AConfirmedStopAndItsFillsEachLeaveALine | n/a(기존) | 4d |
| B3 | 래치 순회 | TestA129MTheRefusalLatchDoesNotOutliveItsEpisode | n/a(기존) | 4d |
| B4 | 래치 삭제 | TestA129MTheRefusalLatchDoesNotOutliveItsEpisode | n/a(기존) | 4d |
