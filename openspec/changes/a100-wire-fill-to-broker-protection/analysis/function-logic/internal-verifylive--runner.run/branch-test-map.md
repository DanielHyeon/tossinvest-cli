# Branch Test Map: `Runner.Run`

- Source: `internal/verifylive/runner.go:462-595`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/runner.go:464` — `if r.includeTrigger {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/runner.go:466` — `if err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/runner.go:474` — `if r.includeTrigger && !r.m0ReceiptUsable() {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/runner.go:479` — `if pending, ok, pendingErr := r.m0PendingCheckpoint(); pendingErr != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `else` at `internal/verifylive/runner.go:484` — `} else if ok {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/runner.go:484` — `} else if ok {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/runner.go:492` — `if halt, err := r.approveBatch(ctx); err != nil \|\| halt != "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/runner.go:502` — `if outcome, err, stop := r.cleanup(ctx); outcome.Step != "" \|\| err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/runner.go:503` — `if outcome.Step != "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/runner.go:506` — `if stop {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B11 | `if` at `internal/verifylive/runner.go:509` — `if outcome.Reason == "" {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B12 | `range` at `internal/verifylive/runner.go:517` — `for _, step := range Steps() {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B13 | `if` at `internal/verifylive/runner.go:518` — `if err := ctx.Err(); err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B14 | `if` at `internal/verifylive/runner.go:525` — `if settled, verdict := r.settled(step.ID); settled {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B15 | `if` at `internal/verifylive/runner.go:536` — `if reason, skip := r.preflight(step); skip {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B16 | `else` at `internal/verifylive/runner.go:538` — `} else {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B17 | `if` at `internal/verifylive/runner.go:544` — `if err := r.recorder.Append(entry); err != nil {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B18 | `if` at `internal/verifylive/runner.go:553` — `if sr.verdict == VerdictAwaitingRestart {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B19 | `if` at `internal/verifylive/runner.go:559` — `if errors.Is(sr.abort, ErrNotATerminal) {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B20 | `if` at `internal/verifylive/runner.go:565` — `if errors.Is(sr.abort, ErrOutsidePlan) {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B21 | `if` at `internal/verifylive/runner.go:574` — `if errors.Is(sr.abort, ErrM0TerminalHold) {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B22 | `if` at `internal/verifylive/runner.go:580` — `if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B23 | `if` at `internal/verifylive/runner.go:589` — `if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {` | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
