# Branch Test Map: `Journal.RestoreOperatingModeProjection`

- Source: `internal/journal/operating_mode.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-journal.json`(`./internal/journal` 시험 28개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 576:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 579:2 | `if p := j.modeProjectorRef(); p != nil` | `TestTheModeIsRestoredAfterARestart` | 편집 전(기준선) | 블록 579.41-587.3을 시험 1개가 실행, PASS |
