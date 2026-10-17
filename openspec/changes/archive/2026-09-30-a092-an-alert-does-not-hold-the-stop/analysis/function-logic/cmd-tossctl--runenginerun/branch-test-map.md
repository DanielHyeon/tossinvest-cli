# Branch Test Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-cmd.json`(`./cmd/tossctl` 시험 99개), 연결 워크트리 `e55102f0`.
- 재번호: 편집 전 번들 없음(이 단위에서 새로 만든 번들).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 190:2 | `if ctx == nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B2 | if at 196:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B3 | if at 202:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B4 | if at 212:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B5 | if at 213:3 | `if clauses := engine.UnmetInterlockClauses(err); clauses != ` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B6 | range at 215:4 | `for _, clause := range clauses` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B7 | if at 224:2 | `if !ectx.Automation.Verified` | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 224.31-226.3을 시험 1개가 실행, PASS |
| B8 | if at 234:2 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 234.63-235.81을 시험 4개가 실행, PASS |
| B9 | if at 235:3 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.S` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B10 | if at 251:2 | `if token, terr := engineProcInstance(os.Getpid()); terr == n` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 251.65-253.3을 시험 4개가 실행, PASS |
| B11 | if at 254:2 | `if merr != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B12 | else at 258:9 | `} else` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B13 | if at 270:2 | `if err != nil` | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 270.16-272.3을 시험 1개가 실행, PASS |
| B14 | if at 274:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B15 | if at 297:2 | `if policyControl != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 297.26-299.3을 시험 3개가 실행, PASS |
| B16 | if at 300:2 | `if policyControlErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 300.29-303.3을 시험 1개가 실행, PASS |
| B17 | if at 305:2 | `if policyRuntime != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 305.26-307.3을 시험 3개가 실행, PASS |
| B18 | if at 308:2 | `if policyRuntimeErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 308.29-311.3을 시험 1개가 실행, PASS |
| B19 | if at 323:2 | `if strategyRuntime != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 323.28-332.3을 시험 1개가 실행, PASS |
| B20 | if at 333:2 | `if projErr != nil` | `TestAFailedStrategyProjectionDoesNotStopTheEngine`, `TestTheDegradedBootLeavesReadyToTheRuntimeSeam` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 333.20-335.3을 시험 2개가 실행, PASS |
| B21 | if at 340:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B22 | if at 344:2 | `if alertControl != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 344.25-346.3을 시험 3개가 실행, PASS |
| B23 | if at 347:2 | `if alertControlErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 347.28-350.3을 시험 1개가 실행, PASS |
| B24 | if at 356:2 | `if modeControlErr == nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B25 | if at 359:3 | `if modeControl != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 측정 표본의 시험 0개 |
| B26 | if at 363:2 | `if modeControlErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 (불변) | 블록 363.27-365.3을 시험 3개가 실행, PASS |
