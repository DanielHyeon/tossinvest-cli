# Risk Pattern Report: `AbortTargets`

- Source: `internal/verifylive/abort.go`
- Range: `65-67`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/abort.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/abort.go:65-67` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/abort.go:65-67` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
