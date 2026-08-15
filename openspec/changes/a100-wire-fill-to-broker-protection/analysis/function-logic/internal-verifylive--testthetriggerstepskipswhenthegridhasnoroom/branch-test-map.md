# Branch Test Map: `TestTheTriggerStepSkipsWhenTheGridHasNoRoom`

- Source: `internal/verifylive/steps_trigger_test.go:407-417`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:411` — `if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {` | `TestTheTriggerStepSkipsWhenTheGridHasNoRoom` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:414` — `if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {` | `TestTheTriggerStepSkipsWhenTheGridHasNoRoom` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
