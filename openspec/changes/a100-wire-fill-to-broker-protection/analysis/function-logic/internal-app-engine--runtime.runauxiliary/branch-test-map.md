# Branch Test Map: `Runtime.runAuxiliary`

- Source: `internal/app/engine/auxiliary.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-engine.json`(`./internal/app/engine` 시험 72개), 연결 워크트리 `e55102f0`.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): B1~B1 → B1~B1(같은 분기, 줄 이동); 새 분기 B2(:110); B2~B2 → B3~B3(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 92:2 | `if r.gracefulStop(ctx, err)` | `TestEveryAuxiliaryExecutorIsStarted`, `TestTheEngineStopsPromptlyWithTheDelivererAttached` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N09 · N10 | 블록 92.30-101.3을 시험 2개가 실행, PASS |
| B2 | if at 110:2 | `if event == ""` | `TestA092TheDelivererStopKeepsItsEvent`, `TestADeadAuxiliaryExecutorIsNotRestarted` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N09 · N10 | 블록 110.17-112.3을 시험 6개가 실행, PASS |
| B3 | if at 117:2 | `if aux.OnStop == nil` | `TestA092TheDelivererStopKeepsItsEvent`, `TestA092TheRelayStopIsNotAnUndeliveredCriticalAlert` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N09 · N10 | 블록 117.23-119.3을 시험 4개가 실행, PASS |
