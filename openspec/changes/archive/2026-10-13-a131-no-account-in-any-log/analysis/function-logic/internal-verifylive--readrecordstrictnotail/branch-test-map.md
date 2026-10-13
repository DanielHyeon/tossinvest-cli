# Branch Test Map: `readRecordStrictNoTail`

- Source: `internal/verifylive/reconcile_record.go` (30-44) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 32:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B2 | if at 36:2 | `if err != nil {` | a131 무편집 분기 — `go test ./internal/verifylive/` 전체 | 무편집 |
| B3 | if at 40:2 | `if err := openRecordEntries(entries, path); err != nil {` | `TestA131UnopenableVerifyLineIsAnError` · `TestA131ReconcileLineIsSealedAndStillBinds` | 편집 영향 |

## a131 시험

- `TestA131ReconcileLineIsSealedAndStillBinds`
