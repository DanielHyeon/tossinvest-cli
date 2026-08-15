# Risk Pattern Report: `engineRuntime`

Source: `cmd/tossctl/engine.go`; target `engineRuntime`.

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| detached loop/goroutine | no current direct goroutine in target | reviewed-safe | `function-logic-map.md` |
| constructor failure after partial loop creation | B1--B6 return before `engine.NewRuntime` | reviewed-safe | `branch-test-map.md` |
| gate/order sequencing | A100 task 3.9 has no current worker seam | pre-edit hazard | `function-logic-map.md` |

The AST was regenerated from current main after the frozen base was recaptured at
`882a0b490b0b6d2eb7abe5c5040c514776f49f3e`. This bundle is pre-edit evidence,
not a claim that task 3.9 is implemented; recapture is required again only if HEAD moves.
