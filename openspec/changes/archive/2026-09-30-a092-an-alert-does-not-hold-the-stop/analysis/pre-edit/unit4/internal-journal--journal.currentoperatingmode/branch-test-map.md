# Branch Test Map: `Journal.CurrentOperatingMode`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 555:2 | `if account == ""` | `TestCurrentOperatingModeNeedsAnAccount` | 편집 전(기준선) | 블록 555.19-558.3을 시험 1개가 실행, PASS |
