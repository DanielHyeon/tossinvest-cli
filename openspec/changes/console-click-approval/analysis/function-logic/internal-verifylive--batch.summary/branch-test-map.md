# Branch Test Map: `Batch.Summary`

- Source: `internal/verifylive/confirm.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if b.Resumed {` (internal/verifylive/confirm.go:217) | TestSummaryCarriesTheListWithoutTheTypedInstruction, TestPromptIsTheSummaryPlusTheTypedTail | yes | yes |
