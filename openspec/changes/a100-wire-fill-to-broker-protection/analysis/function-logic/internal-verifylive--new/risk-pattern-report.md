# Risk Pattern Report: `New`

- Source: `internal/verifylive/runner.go`
- Range: `320-425`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/runner.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/runner.go:320-425` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/runner.go:320-425` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
