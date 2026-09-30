# Branch Test Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-engine.json`(`./internal/app/engine` 시험 12개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 239:2 | `if err := checkProjectionWired(in.journal); err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 261:2 | `if err := tracker.Restore(ctx); err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B3 | if at 269:2 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); er` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 293:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B5 | if at 317:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
