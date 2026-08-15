# Branch Test Map: `TestTheMarketNeverComingIsNotAFailure`

- Source: `internal/verifylive/steps_trigger_test.go:236-254`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:240` — `if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {` | `TestTheMarketNeverComingIsNotAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:244` — `if !strings.Contains(e.Reason, "INCONCLUSIVE") {` | `TestTheMarketNeverComingIsNotAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:247` — `if !observationEquals(t, h.entries(), StepConditionalTrigger,` | `TestTheMarketNeverComingIsNotAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/steps_trigger_test.go:251` — `if out := Outstanding(h.entries()); len(out) != 0 {` | `TestTheMarketNeverComingIsNotAFailure` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
