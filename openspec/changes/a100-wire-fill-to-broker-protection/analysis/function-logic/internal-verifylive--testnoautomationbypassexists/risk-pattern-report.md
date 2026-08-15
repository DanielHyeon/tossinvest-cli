# Risk Pattern Report: `TestNoAutomationBypassExists`

- Source: `internal/verifylive/static_test.go`
- Range: `192-216`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/static_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/static_test.go:192-216` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/static_test.go:192-216` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
