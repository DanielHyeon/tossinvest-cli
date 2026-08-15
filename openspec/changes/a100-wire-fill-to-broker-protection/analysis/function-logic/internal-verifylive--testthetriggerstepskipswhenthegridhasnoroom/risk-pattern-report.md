# Risk Pattern Report: `TestTheTriggerStepSkipsWhenTheGridHasNoRoom`

- Source: `internal/verifylive/steps_trigger_test.go`
- Range: `407-417`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/steps_trigger_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/steps_trigger_test.go:407-417` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/steps_trigger_test.go:407-417` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
