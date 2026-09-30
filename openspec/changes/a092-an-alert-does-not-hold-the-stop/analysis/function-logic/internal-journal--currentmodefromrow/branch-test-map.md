# Branch Test Map: `currentModeFromRow`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B3 → B1~B3(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 676:2 | `switch` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M01 · M04(간접) | 블록 674.79-676.9을 시험 16개가 실행, PASS |
| B2 | case at 677:2 | `case errors.Is(err, errNoOperatingMode):` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M01 · M04(간접) | 블록 677.42-679.66을 시험 16개가 실행, PASS |
| B3 | case at 680:2 | `case err != nil:` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M01 · M04(간접) | 측정 표본의 시험 0개 |
