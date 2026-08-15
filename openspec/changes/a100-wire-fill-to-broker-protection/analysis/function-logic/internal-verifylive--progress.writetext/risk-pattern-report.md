# Risk Pattern Report: `Progress.WriteText`

- Source: `internal/verifylive/report.go`
- Range: `346-379`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/report.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/report.go:346-379` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/report.go:346-379` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
