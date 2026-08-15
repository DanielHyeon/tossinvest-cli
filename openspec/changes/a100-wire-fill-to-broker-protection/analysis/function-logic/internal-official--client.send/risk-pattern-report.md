# Risk Pattern Report: `Client.send`

- Source: `internal/official/client.go`
- Range: `320-366`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/official/client.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/official/client.go:320-366` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/official/client.go:320-366` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
