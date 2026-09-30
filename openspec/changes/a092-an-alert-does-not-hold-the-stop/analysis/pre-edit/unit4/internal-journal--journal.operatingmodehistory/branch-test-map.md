# Branch Test Map: `Journal.OperatingModeHistory`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 597:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | for at 603:2 | `for rows.Next()` | `TestConcurrentEscalationsConvergeOnTheStrictestMode`, `TestConservativePrecedenceKeepsTheStricterMode` | 편집 전(기준선) | 블록 603.18-605.17을 시험 3개가 실행, PASS |
| B3 | if at 605:3 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 610:2 | `if err := rows.Err(); err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
