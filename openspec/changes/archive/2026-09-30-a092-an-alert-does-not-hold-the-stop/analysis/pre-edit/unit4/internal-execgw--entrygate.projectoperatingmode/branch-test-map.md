# Branch Test Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-execgw.json`(`./internal/execgw` 시험 11개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 40:2 | `if !rec.BlocksEntry()` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 44:2 | `if rec.Actor != ""` | `TestARejectedCredentialTightensTheOperatingMode`, `TestTheModeLatchIsReplacedRatherThanAccumulated` | 편집 전(기준선) | 블록 44.21-46.3을 시험 4개가 실행, PASS |
| B3 | if at 47:2 | `if rec.Cause != ""` | `TestARejectedCredentialTightensTheOperatingMode`, `TestTheModeLatchIsReplacedRatherThanAccumulated` | 편집 전(기준선) | 블록 47.21-49.3을 시험 4개가 실행, PASS |
