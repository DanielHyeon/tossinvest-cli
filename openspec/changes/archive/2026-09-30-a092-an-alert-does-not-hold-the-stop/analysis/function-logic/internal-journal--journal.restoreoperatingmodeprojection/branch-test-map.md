# Branch Test Map: `Journal.RestoreOperatingModeProjection`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B2 → B1~B2(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 591:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M04 | 측정 표본의 시험 0개 |
| B2 | if at 594:2 | `if p := j.modeProjectorRef(); p != nil` | `TestA092TransitionsCarryTheirCommitSequence`, `TestTheModeIsRestoredAfterARestart` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M04 | 블록 594.41-603.3을 시험 2개가 실행, PASS |
