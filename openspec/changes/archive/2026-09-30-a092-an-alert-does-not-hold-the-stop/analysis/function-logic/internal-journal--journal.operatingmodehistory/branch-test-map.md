# Branch Test Map: `Journal.OperatingModeHistory`

- Source: `internal/journal/operating_mode.go`; **편집 뒤** 측정 — `analysis/harness/``coverage-post-unit4-journal.json`(`./internal/journal` 시험 30개), 연결 워크트리 `2714e393`.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B4 → B1~B4(같은 분기, 줄 이동).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 613:2 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M02 | 측정 표본의 시험 0개 |
| B2 | for at 619:2 | `for rows.Next()` | `TestConcurrentEscalationsConvergeOnTheStrictestMode`, `TestConservativePrecedenceKeepsTheStricterMode` | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M02 | 블록 619.18-621.17을 시험 3개가 실행, PASS |
| B3 | if at 621:3 | `if err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M02 | 측정 표본의 시험 0개 |
| B4 | if at 626:2 | `if err := rows.Err(); err != nil` | (미실행) | 편집 전 RED: `analysis/mutation-unit4/red-*.log` · 변이 M02 | 측정 표본의 시험 0개 |
