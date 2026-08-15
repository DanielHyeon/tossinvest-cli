# Branch Test Map: `triggerOptions`

- Source: `internal/verifylive/steps_trigger_test.go:57-70`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/steps_trigger_test.go:60` — `if err := os.Chmod(dir, 0o700); err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/steps_trigger_test.go:64` — `if err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
