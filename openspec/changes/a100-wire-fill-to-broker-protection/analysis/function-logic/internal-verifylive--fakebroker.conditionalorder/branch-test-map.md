# Branch Test Map: `fakeBroker.ConditionalOrder`

- Source: `internal/verifylive/fake_broker_test.go:738-767`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:743` — `if !ok {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:746` — `if id != f.triggerCondID {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:751` — `if f.conditional404AfterReads > 0 && n >= f.conditional404AfterReads {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:754` — `if f.fireAfterReads > 0 && n >= f.fireAfterReads {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:758` — `if f.linkAfterReads > 0 && n >= f.linkAfterReads {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/fake_broker_test.go:759` — `if f.firedID == "" {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
