# Branch Test Map: `Runner.readOrder`

- Source: `internal/verifylive/steps.go:924-973`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps.go:925` — `if r.m0ReadSource != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps.go:931` — `if r.m0ReceiptErr != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/steps.go:934` — `if err != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/steps.go:935` — `if r.m0CriticalWindow {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/steps.go:940` — `if r.m0ReceiptErr != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/steps.go:944` — `if err := json.Unmarshal(raw, &view); err != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/steps.go:953` — `if err != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/steps.go:954` — `if r.m0CriticalWindow {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/steps.go:959` — `if r.m0ReceiptErr != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/steps.go:960` — `if r.m0CriticalWindow {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `internal/verifylive/steps.go:966` — `if err := json.Unmarshal(raw, &view); err != nil {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `if` at `internal/verifylive/steps.go:967` — `if r.m0CriticalWindow {` | `TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
