# Risk Pattern Report: `stepRun.resolve`

- Source: `internal/verifylive/runner.go`
- Range: `954-976`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/runner.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/runner.go:954-976` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/runner.go:954-976` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
