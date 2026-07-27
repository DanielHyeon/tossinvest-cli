# Branch Test Map: `RiskGuardian.IssueEntry`

Source: `internal/execgw/riskguardian.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `req.Collect == nil` | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether` (negative case in riskguardian_test.go) | pre-existing | yes |
| B2 | `scopedIntent` error | riskguardian_test.go scoping cases | pre-existing | yes |
| B3 | `intent.Side != SideBuy` | `TestAReductionIsNotObserved` | pre-existing | yes |
| B4 | the chain refused (`!verdict.Allowed`) | `TestARefusalIsRecordedForTheFirstTime` | pre-existing | yes |
| B5 | `risk.EntryExposureValue` refused | `TestARefusalIsRecordedForTheFirstTime` (same arm, exposure variant) | pre-existing | yes |
| B6 | the recollection closure's `req.Collect` error | riskguardian_test.go recollection cases | pre-existing | yes |
| B7 | `exposureUsage` error inside the closure | riskguardian_test.go | pre-existing | yes |
| B8 | `RecordDecisionAndReserveWithRecollection` failed | `TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome` | pre-existing | yes |
| B9 | the refusal is an `*IssueRefusal` at `StageIssuance` | `TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome` | pre-existing | yes |
