# Branch Test Map: `TestTheTriggerBasisIsDecidedByTheOrderOfTwoObservations`

- Source: `internal/verifylive/steps_trigger_test.go:201-224`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:209` — `if got, _ := h.observation(StepConditionalTrigger, "conditional.trigger.basis"); got != "bid" {` | `TestTheTriggerBasisIsDecidedByTheOrderOfTwoObservations` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:220` — `if got, _ := h.observation(StepConditionalTrigger, "conditional.trigger.basis"); got != "last-trade" {` | `TestTheTriggerBasisIsDecidedByTheOrderOfTwoObservations` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
