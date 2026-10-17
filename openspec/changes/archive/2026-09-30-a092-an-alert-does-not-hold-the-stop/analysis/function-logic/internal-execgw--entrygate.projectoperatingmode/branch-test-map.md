# Branch Test Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-execgw.json`(`./internal/execgw` 시험 16개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 삭제 B1; B2~B3 → B1~B2(같은 분기, 줄 이동); 새 분기 B3(:49), B4(:54), B5(:55), B6(:64).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 37:2 | `if rec.Actor != ""` | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 37.21-39.3을 시험 8개가 실행, PASS |
| B2 | if at 40:2 | `if rec.Cause != ""` | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 40.21-42.3을 시험 8개가 실행, PASS |
| B3 | if at 49:2 | `if rec.Seq <= g.modeSeq` | `TestA092AStaleProjectionIsNotApplied` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 49.26-51.3을 시험 1개가 실행, PASS |
| B4 | if at 54:2 | `if !rec.BlocksEntry()` | `TestA092AStaleProjectionIsNotApplied`, `TestA092TheModeRevisionMovesOnlyWhenPresenceChanges` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 54.24-55.10을 시험 2개가 실행, PASS |
| B5 | if at 55:3 | `if had` | `TestA092AStaleProjectionIsNotApplied`, `TestA092TheModeRevisionMovesOnlyWhenPresenceChanges` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 55.10-58.4을 시험 2개가 실행, PASS |
| B6 | if at 64:2 | `if !had` | `TestA092AProjectionReplacesTheModeLatch`, `TestA092AStaleProjectionIsNotApplied` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M05 · M06 · M07 · M08 | 블록 64.10-66.3을 시험 8개가 실행, PASS |
