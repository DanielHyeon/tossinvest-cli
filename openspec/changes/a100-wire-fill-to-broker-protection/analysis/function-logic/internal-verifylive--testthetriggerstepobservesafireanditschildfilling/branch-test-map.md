# Branch Test Map: `TestTheTriggerStepObservesAFireAndItsChildFilling`

- Source: `internal/verifylive/steps_trigger_test.go:131-193`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:140` — `if got := h.verdict(StepConditionalTrigger); got != VerdictPass {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `range` at `internal/verifylive/steps_trigger_test.go:146` — `for _, key := range []string{` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:153` — `if !ok \|\| value == "unobserved" {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/steps_trigger_test.go:157` — `if _, err := time.Parse(time.RFC3339Nano, value); err != nil {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/steps_trigger_test.go:162` — `if detail := observationDetail(t, entries, StepConditionalTrigger, key); !strings.Contains(detail, "±") {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/steps_trigger_test.go:167` — `if !observationEquals(t, entries, StepConditionalTrigger, "conditional.trigger_observed", "true") {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/steps_trigger_test.go:170` — `if !observationEquals(t, entries, StepConditionalTrigger, "conditional.triggered_order_id_exposed", "true") {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/steps_trigger_test.go:173` — `if v, _ := h.observation(StepConditionalTrigger, "conditional.triggered_order_latency"); v == "unverified" {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/steps_trigger_test.go:176` — `if v, _ := h.observation(StepConditionalTrigger, "conditional.trigger.book_at_trigger"); !strings.Contains(v, "bid") {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/steps_trigger_test.go:182` — `if out := Outstanding(entries); len(out) != 0 {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `internal/verifylive/steps_trigger_test.go:186` — `if !child.Filled \|\| child.Cancelled {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `if` at `internal/verifylive/steps_trigger_test.go:190` — `if child.ChainID == "" {` | `TestTheTriggerStepObservesAFireAndItsChildFilling` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
