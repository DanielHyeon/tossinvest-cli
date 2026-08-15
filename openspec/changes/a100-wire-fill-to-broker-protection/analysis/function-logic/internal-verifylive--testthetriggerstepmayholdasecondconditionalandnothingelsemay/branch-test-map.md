# Branch Test Map: `TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay`

- Source: `internal/verifylive/steps_trigger_test.go:427-449`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:435` — `if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err != nil {` | `TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:438` — `if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalRegister)}); err == nil {` | `TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:446` — `if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err == nil {` | `TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
