# Branch Test Map: `Runner.finishTrigger`

- Source: `internal/verifylive/steps_trigger.go:566-659`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger.go:586` — `if !obs.triggeredAt.IsZero() && !obs.childSeenAt.IsZero() {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `else` at `internal/verifylive/steps_trigger.go:591` — `} else {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps_trigger.go:596` — `if !obs.triggeredAt.IsZero() {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `switch` at `internal/verifylive/steps_trigger.go:602` — `switch {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `case` at `internal/verifylive/steps_trigger.go:603` — `case !obs.childFilledAt.IsZero():` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/steps_trigger.go:604` — `if !r.m0PassReady() {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `case` at `internal/verifylive/steps_trigger.go:617` — `case !obs.triggeredAt.IsZero():` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `case` at `internal/verifylive/steps_trigger.go:632` — `case obs.crossedWithoutFiring:` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `case` at `internal/verifylive/steps_trigger.go:642` — `case obs.cancelled && obs.raceUnknown:` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `case` at `internal/verifylive/steps_trigger.go:650` — `case obs.cancelled:` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `case` at `internal/verifylive/steps_trigger.go:655` — `default:` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
