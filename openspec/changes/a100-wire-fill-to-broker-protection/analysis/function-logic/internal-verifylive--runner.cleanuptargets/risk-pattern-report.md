# Risk Pattern Report: `Runner.cleanupTargets`

- Source: `internal/verifylive/cleanup.go`
- Range: `103-108`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/cleanup.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/cleanup.go:103-108` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/cleanup.go:103-108` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
