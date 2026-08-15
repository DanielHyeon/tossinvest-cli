# Branch Test Map: `fakeBroker.CreateConditionalOrder`

- Source: `internal/verifylive/fake_broker_test.go:910-941`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:913` — `if body.ClientOrderID != "" {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:917` — `if seen && prior.body == canonical && f.honourIdempotency {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:930` — `if body.ClientOrderID != "" {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:933` — `if strings.HasPrefix(body.ClientOrderID, "TRIGGER-") {` | `TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
