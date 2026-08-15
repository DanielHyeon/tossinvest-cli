# Risk Pattern Report: `Client.ConditionalOrdersRaw`

- Source: `internal/official/conditional_reads.go`
- Range: `156-211`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/official/conditional_reads.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/official/conditional_reads.go:156-211` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/official/conditional_reads.go:156-211` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
