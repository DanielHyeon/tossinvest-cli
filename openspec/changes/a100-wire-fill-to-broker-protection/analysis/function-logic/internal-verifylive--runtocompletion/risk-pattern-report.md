# Risk Pattern Report: `runToCompletion`

- Source: `internal/verifylive/runner_test.go`
- Range: `28-59`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/runner_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/runner_test.go:28-59` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/runner_test.go:28-59` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
