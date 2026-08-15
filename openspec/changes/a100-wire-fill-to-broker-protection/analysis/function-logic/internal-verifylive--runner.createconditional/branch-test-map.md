# Branch Test Map: `Runner.createConditional`

- Source: `internal/verifylive/mutate.go:503-556`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/mutate.go:504` — `if err := r.checkConditionalCap(sr); err != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/mutate.go:508` — `if err := r.gate(sr, request{` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/mutate.go:515` — `if err := r.appendM0Checkpoint(M0Checkpoint{` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/mutate.go:522` — `if r.m0ReceiptUsable() {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/mutate.go:525` — `if r.m0BeforeConditionalCreate != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/mutate.go:526` — `if err := r.m0BeforeConditionalCreate(); err != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/mutate.go:534` — `if err != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/mutate.go:537` — `if strings.TrimSpace(ref.ID) == "" {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/mutate.go:540` — `if r.m0AfterConditionalCreate != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/mutate.go:541` — `if err := r.m0AfterConditionalCreate(); err != nil {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `internal/verifylive/mutate.go:545` — `if err := r.appendM0Checkpoint(M0Checkpoint{` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
