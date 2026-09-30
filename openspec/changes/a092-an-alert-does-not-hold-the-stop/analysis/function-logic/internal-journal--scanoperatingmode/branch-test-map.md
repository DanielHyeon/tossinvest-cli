# Branch Test Map: `scanOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 교체 B1 → B1; B2~B3 → B2~B3(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 718:2 | `if err := row.Scan(&record.Seq, &record.ID, &record.AccountR` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M04(간접) | 측정 표본의 시험 0개 |
| B2 | if at 720:3 | `if errors.Is(err, sql.ErrNoRows)` | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M04(간접) | 블록 720.36-722.4을 시험 16개가 실행, PASS |
| B3 | if at 726:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M03 · M04(간접) | 측정 표본의 시험 0개 |
