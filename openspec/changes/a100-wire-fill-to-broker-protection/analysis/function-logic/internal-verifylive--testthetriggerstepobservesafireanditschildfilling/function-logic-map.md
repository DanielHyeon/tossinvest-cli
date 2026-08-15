# Function Logic Map: `TestTheTriggerStepObservesAFireAndItsChildFilling`

- Source: `internal/verifylive/steps_trigger_test.go:131-193`
- Qualified function: `TestTheTriggerStepObservesAFireAndItsChildFilling`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/steps_trigger_test.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:140` — `if got := h.verdict(StepConditionalTrigger); got != VerdictPass {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B2 | `range` at `internal/verifylive/steps_trigger_test.go:146` — `for _, key := range []string{` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B3 | `if` at `internal/verifylive/steps_trigger_test.go:153` — `if !ok \|\| value == "unobserved" {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B4 | `if` at `internal/verifylive/steps_trigger_test.go:157` — `if _, err := time.Parse(time.RFC3339Nano, value); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B5 | `if` at `internal/verifylive/steps_trigger_test.go:162` — `if detail := observationDetail(t, entries, StepConditionalTrigger, key); !strings.Contains(detail, "±") {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B6 | `if` at `internal/verifylive/steps_trigger_test.go:167` — `if !observationEquals(t, entries, StepConditionalTrigger, "conditional.trigger_observed", "true") {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B7 | `if` at `internal/verifylive/steps_trigger_test.go:170` — `if !observationEquals(t, entries, StepConditionalTrigger, "conditional.triggered_order_id_exposed", "true") {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B8 | `if` at `internal/verifylive/steps_trigger_test.go:173` — `if v, _ := h.observation(StepConditionalTrigger, "conditional.triggered_order_latency"); v == "unverified" {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B9 | `if` at `internal/verifylive/steps_trigger_test.go:176` — `if v, _ := h.observation(StepConditionalTrigger, "conditional.trigger.book_at_trigger"); !strings.Contains(v, "bid") {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B10 | `if` at `internal/verifylive/steps_trigger_test.go:182` — `if out := Outstanding(entries); len(out) != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B11 | `if` at `internal/verifylive/steps_trigger_test.go:186` — `if !child.Filled \|\| child.Cancelled {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |
| B12 | `if` at `internal/verifylive/steps_trigger_test.go:190` — `if child.ChainID == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestTheTriggerStepObservesAFireAndItsChildFilling` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `triggerHarness` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `f.firesOnRead` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `runToCompletion` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `triggerOptions` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.verdict` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `LastEntry` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.entries` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Fatalf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `h.observation` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `time.Parse` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `observationDetail` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.Contains` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `observationEquals` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `t.Error` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Outstanding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `lastArtifactOfKind` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
