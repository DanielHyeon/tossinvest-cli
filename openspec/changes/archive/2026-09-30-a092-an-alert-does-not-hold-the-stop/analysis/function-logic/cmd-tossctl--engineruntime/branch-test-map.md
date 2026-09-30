# Branch Test Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit5-cmd.json`(`./cmd/tossctl` 시험 99개), 연결 워크트리 `e55102f0`.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 분기 좌표 불변.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 639:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 측정 표본의 시험 0개 |
| B2 | if at 648:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 블록 648.16-650.3을 시험 1개가 실행, PASS |
| B3 | if at 659:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 블록 659.16-661.3을 시험 1개가 실행, PASS |
| B4 | if at 664:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 블록 664.16-666.3을 시험 1개가 실행, PASS |
| B5 | if at 668:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 측정 표본의 시험 0개 |
| B6 | if at 680:2 | `if err != nil` | (미실행) | RED: `analysis/mutation-unit5/red-b7-pre-edit.log`(B#7 · A#2) · 변이 N08 | 측정 표본의 시험 0개 |
