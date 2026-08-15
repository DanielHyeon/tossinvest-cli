# Risk Pattern Report: `TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay`

- Source: `internal/verifylive/steps_trigger_test.go`
- Range: `427-449`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml internal/verifylive/steps_trigger_test.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `internal/verifylive/steps_trigger_test.go:427-449` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `internal/verifylive/steps_trigger_test.go:427-449` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
