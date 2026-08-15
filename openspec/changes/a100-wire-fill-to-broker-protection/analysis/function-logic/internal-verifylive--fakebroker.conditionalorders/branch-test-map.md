# Branch Test Map: `fakeBroker.ConditionalOrders`

- Source: `internal/verifylive/fake_broker_test.go:437-463`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:441` — `if f.rejectBadConditionalStatus && status != "" && status != "OPEN" && status != "CLOSED" {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `range` at `internal/verifylive/fake_broker_test.go:448` — `for _, c := range f.conds {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:451` — `if status == "OPEN" && c.Status != "WATCHING" {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/fake_broker_test.go:454` — `if status != "" && status != "OPEN" && status != "CLOSED" && c.Status != status {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:457` — `if symbol != "" && c.Symbol != symbol {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
