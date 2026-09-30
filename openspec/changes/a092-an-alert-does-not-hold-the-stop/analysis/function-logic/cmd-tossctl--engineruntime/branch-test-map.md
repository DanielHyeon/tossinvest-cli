# Branch Test Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-cmd.json`(`./cmd/tossctl` 시험 90개), 연결 워크트리 `2714e393`.
- 재번호: 편집 전 번들 없음(이 단위에서 새로 만든 번들).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 639:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 측정 표본의 시험 0개 |
| B2 | if at 648:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 블록 648.16-650.3을 시험 1개가 실행, PASS |
| B3 | if at 659:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 블록 659.16-661.3을 시험 1개가 실행, PASS |
| B4 | if at 664:2 | `if err != nil` | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 블록 664.16-666.3을 시험 1개가 실행, PASS |
| B5 | if at 668:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 측정 표본의 시험 0개 |
| B6 | if at 680:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 (불변) | 측정 표본의 시험 0개 |
