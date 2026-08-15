# Branch Test Map: `TestACancelThatRacesATriggerDoesNotReportThatNothingHappened`

- Source: `internal/verifylive/steps_trigger_test.go:262-281`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:270` — `if !observationEquals(t, entries, StepConditionalTrigger,` | `TestACancelThatRacesATriggerDoesNotReportThatNothingHappened` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:275` — `if got := h.verdict(StepConditionalTrigger); got == VerdictSkipped {` | `TestACancelThatRacesATriggerDoesNotReportThatNothingHappened` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:278` — `if !observationEquals(t, entries, StepConditionalTrigger, "conditional.trigger_observed", "true") {` | `TestACancelThatRacesATriggerDoesNotReportThatNothingHappened` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
