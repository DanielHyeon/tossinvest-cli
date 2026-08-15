# Risk Pattern Report: `harness.build`

- Source: `internal/verifylive/fake_broker_test.go`
- Range: `1207-1271`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/fake_broker_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/fake_broker_test.go:1207-1271` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/fake_broker_test.go:1207-1271` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
