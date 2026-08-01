# Branch Test Map: `RiskGuardian.IssueEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | `req.Collect == nil` refuses before risk evaluation | `TestRiskGuardianIssueEntryRejectsMissingCollectorBeforeIssuingAuthority` | missing direct test | pass |
| B2 | `g.scopedIntent` rejects invalid/cross-account intent | `TestTheGuardianRefusesAnIntentForAnotherAccount` | baseline | pass |
| B3 | a non-BUY intent is rejected by the entry-only issuer | entry/reduction separation tests | baseline | pass |
| B4 | pure chain refusal is observed and optional loss escalation is joined | `TestAChainRefusalIssuesNothing`, escalation tests | baseline | pass |
| B5 | `risk.EntryExposureValue` refuses invalid exposure arithmetic | risk/Guardian invalid-input tests | baseline | pass |
| B6 | snapshot collector returns an error inside `collectIssue` | recollection collector-error tests | baseline | pass |
| B7 | `exposureUsage` rejects mixed/invalid snapshot currency | mixed-currency reservation tests | baseline | pass |
| B8 | no private strategy plan: use ordinary atomic decision+reservation recollection | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether`, recollection/concurrency tests | baseline | pass |
| B9 | private strategy plan exists: use strategy atomic recollection | no authentic direct strategy success while source manifest is unavailable | split draft | activation-gated unverified |
| A047-1 | ordinary caller cannot attach a strategy plan | compile/API surface inspection | public draft existed | pass |
| A047-2 | strategy issuance commits decision/reservation/lineage/start together | journal strategy atomic/rollback tests | split draft | pass |
| A047-3 | strategy receipt returns canonical quantity/client id | strategy adapter integration | post-commit lookup draft | pass |
| B10 | strategy recollection callback returns the `collectIssue` error | no authentic direct strategy call while source manifest is unavailable | missing | activation-gated unverified |
| B11 | strategy recollection succeeds, so ordinary issue result and strategy receipt are copied | no authentic direct strategy call while source manifest is unavailable | missing | activation-gated unverified |
| B12 | either ordinary or strategy recollection returns error; map it to stable issuance refusal | reservation/refusal tests cover ordinary path; strategy half activation-gated | baseline | partial |
| B13 | typed issuance-stage refusal copies its stable reason into the observation | `TestTheIssuanceReasonsAreDistinguishable`, observation tests | baseline | pass |
