# Branch Test Map: `runToCompletion`

- Source: `internal/verifylive/runner_test.go:28-59`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/runner_test.go:30` — `if opts.IncludeTrigger && opts.Receipt != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/runner_test.go:35` — `if err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/runner_test.go:38` — `if only.Halted {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/runner_test.go:44` — `if err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/runner_test.go:47` — `if !first.Halted {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/runner_test.go:55` — `if err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
