# Risk Pattern Report: `BuildProgress`

- Source: `internal/verifylive/report.go`
- Range: `315-343`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/report.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/report.go:315-343` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/report.go:315-343` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
