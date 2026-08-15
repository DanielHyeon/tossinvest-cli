# Risk Pattern Report: `Runner.readConditional`

- Source: `internal/verifylive/steps.go`
- Range: `975-1055`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/steps.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/steps.go:975-1055` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/steps.go:975-1055` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
