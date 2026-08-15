# Branch Test Map: `TestTheOrderCapIsUntouchedByTheTriggerStep`

- Source: `internal/verifylive/steps_trigger_test.go:455-468`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:456` — `if MaxLiveOrders != 1 {` | `TestTheOrderCapIsUntouchedByTheTriggerStep` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:465` — `if err := r.checkOrderCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err == nil {` | `TestTheOrderCapIsUntouchedByTheTriggerStep` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
