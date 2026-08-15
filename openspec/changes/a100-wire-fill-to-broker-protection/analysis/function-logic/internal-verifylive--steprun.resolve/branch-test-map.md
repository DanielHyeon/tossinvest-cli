# Branch Test Map: `stepRun.resolve`

- Source: `internal/verifylive/runner.go:954-976`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch` at `internal/verifylive/runner.go:955` — `switch {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `case` at `internal/verifylive/runner.go:956` — `case err == nil:` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/runner.go:957` — `if sr.verdict == "" {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `case` at `internal/verifylive/runner.go:960` — `case errors.Is(err, ErrOutsidePlan):` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `case` at `internal/verifylive/runner.go:962` — `case errors.Is(err, ErrM0TerminalHold):` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `case` at `internal/verifylive/runner.go:965` — `case errors.Is(err, ErrRefused), errors.Is(err, ErrConfirmationExpired):` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `case` at `internal/verifylive/runner.go:967` — `case errors.Is(err, ErrNotATerminal):` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `case` at `internal/verifylive/runner.go:970` — `case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `case` at `internal/verifylive/runner.go:973` — `default:` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
