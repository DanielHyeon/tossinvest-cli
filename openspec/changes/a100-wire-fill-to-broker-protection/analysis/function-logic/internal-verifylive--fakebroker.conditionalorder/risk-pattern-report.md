# Risk Pattern Report: `fakeBroker.ConditionalOrder`

- Source: `internal/verifylive/fake_broker_test.go`
- Range: `738-767`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/fake_broker_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/fake_broker_test.go:738-767` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/fake_broker_test.go:738-767` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
