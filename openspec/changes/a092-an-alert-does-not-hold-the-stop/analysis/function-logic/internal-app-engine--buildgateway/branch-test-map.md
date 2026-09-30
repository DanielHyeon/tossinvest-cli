# Branch Test Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-engine.json`(`./internal/app/engine` 시험 45개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 새 분기 B4(:273); B4~B5 → B5~B6(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 239:2 | `if err := checkProjectionWired(in.journal); err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
| B2 | if at 261:2 | `if err := tracker.Restore(ctx); err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
| B3 | if at 269:2 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); er` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
| B4 | if at 273:2 | `if err := bindOperatingModeProjection(ctx, in.journal, entry` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
| B5 | if at 297:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
| B6 | if at 321:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M09 | 측정 표본의 시험 0개 |
