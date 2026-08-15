# Branch Test Map: `fakeBroker.OrderRawByID`

- Source: `internal/verifylive/fake_broker_test.go:441-481`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/fake_broker_test.go:442` — `if f.beforeOrderRawByID != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/fake_broker_test.go:449` — `if !ok {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/fake_broker_test.go:453` — `if orderID != f.firedID {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `else` at `internal/verifylive/fake_broker_test.go:455` — `} else {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/fake_broker_test.go:460` — `if filled {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/fake_broker_test.go:464` — `if f.childIdentitySymbol != "" {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/fake_broker_test.go:466` — `if err := json.Unmarshal(result, &decoded); err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `if` at `internal/verifylive/fake_broker_test.go:472` — `if f.childIdentityQty != "" {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/fake_broker_test.go:474` — `if err := json.Unmarshal(result, &decoded); err != nil {` | `TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
