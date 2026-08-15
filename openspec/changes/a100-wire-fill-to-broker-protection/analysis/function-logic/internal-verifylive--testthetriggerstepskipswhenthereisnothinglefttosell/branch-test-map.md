# Branch Test Map: `TestTheTriggerStepSkipsWhenThereIsNothingLeftToSell`

- Source: `internal/verifylive/steps_trigger_test.go:387-401`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:395` — `if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {` | `TestTheTriggerStepSkipsWhenThereIsNothingLeftToSell` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:398` — `if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {` | `TestTheTriggerStepSkipsWhenThereIsNothingLeftToSell` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
