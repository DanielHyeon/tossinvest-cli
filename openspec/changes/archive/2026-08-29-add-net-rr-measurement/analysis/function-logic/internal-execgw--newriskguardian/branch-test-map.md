# Branch Test Map: `NewRiskGuardian`

Source: `internal/execgw/riskguardian.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `opts.Journal == nil` | riskguardian_test.go construction cases | pre-existing | yes |
| B2 | `account == ""` | riskguardian_test.go | pre-existing | yes |
| B3 | `opts.PolicyVersion` blank | riskguardian_test.go | pre-existing | yes |
| B4 | `opts.Policy.Validate()` error | riskguardian_test.go | pre-existing | yes |
| B5 | `!opts.Costs.Configured()` | riskguardian_test.go | pre-existing | yes |
| B6 | `ExposureLimitsFor` error | riskguardian_test.go | pre-existing | yes |
| B7 | `EncodeLimits` error | riskguardian_test.go | pre-existing | yes |
| B8 | `opts.Clock == nil` | riskguardian_test.go | pre-existing | yes |
| B9 | `opts.TTL <= 0` | riskguardian_test.go | pre-existing | yes |
| B10 | `opts.NewID == nil` | riskguardian_test.go | pre-existing | yes |
| B11 | the constructed struct literal's field set | `TestAGuardianWithNoObserverIssuesAsBefore`, `TestTheCostBasisIsRecordedWithEveryRow` | pre-existing | yes |
