# Branch Test Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-engine.json`(`./internal/app/engine` 시험 72개), 연결 워크트리 `e55102f0`.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 교체 B5, B6 → B5; B7~B7 → B6~B6(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 320:2 | `if c == nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 측정 표본의 시험 0개 |
| B2 | if at 323:2 | `if !c.Automation.Verified` | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 블록 323.28-325.3을 시험 1개가 실행, PASS |
| B3 | if at 327:2 | `if !ok` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 측정 표본의 시험 0개 |
| B4 | if at 343:2 | `if opts.Names == nil` | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 블록 343.23-345.3을 시험 3개가 실행, PASS |
| B5 | if at 348:2 | `if c.Notifier != nil` | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 블록 348.23-351.3을 시험 3개가 실행, PASS |
| B6 | if at 352:2 | `if opts.Floor == nil` | `TestA092ExitObserverGetsRecordOnlyAlertPaths`, `TestA092TheExitObserverOverridesACallersSyncAlertPath` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N07 · N11 | 블록 352.23-354.3을 시험 3개가 실행, PASS |
