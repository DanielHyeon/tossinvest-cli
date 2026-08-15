# Risk Pattern Report: `Runner.createConditional`

- Source: `internal/verifylive/mutate.go`
- Range: `503-556`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/mutate.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/mutate.go:503-556` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/mutate.go:503-556` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
