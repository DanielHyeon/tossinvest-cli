# Risk Pattern Report: `adaptProtectionConditional`

- Source: `internal/official/protection_reads.go`
- Range: `49-58`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/official/protection_reads.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/official/protection_reads.go:49-58` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/official/protection_reads.go:49-58` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
