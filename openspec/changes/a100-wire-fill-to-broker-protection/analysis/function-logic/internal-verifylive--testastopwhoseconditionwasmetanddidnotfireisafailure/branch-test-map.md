# Branch Test Map: `TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure`

- Source: `internal/verifylive/steps_trigger_test.go:359-380`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:366` — `if got := h.verdict(StepConditionalTrigger); got != VerdictFail {` | `TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:371` — `if !observationEquals(t, h.entries(), StepConditionalTrigger,` | `TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:377` — `if out := Outstanding(h.entries()); len(out) != 0 {` | `TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
