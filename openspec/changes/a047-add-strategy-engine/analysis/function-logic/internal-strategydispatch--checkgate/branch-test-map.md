# Branch Test Map: `checkGate`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | manifest-bound decision provenance mismatch precedes lane blocker | `TestPostValidationDispatchCoreRefusesEveryInitialGateBeforeIssuerWithStablePrecedence`, activation row | missing direct precedence | pass |
| B2 | any DecisionRecord field or order settings/type/currency mismatch activates refusal | `TestDecisionBindingRejectsMismatchInEveryDecisionRecordField` plus three direct order rows | existing partial | pass, 60/60 plus 3 direct rows |
| B3 | exact activation reaches ordered operational switch | initial gate table and positive spy test | missing full order | pass |
| B4 | lane desired/effective OFF | initial gate table | missing | pass |
| B5 | kill switch ON | initial gate table | missing | pass |
| B6 | protection unwired | initial gate table | missing | pass |
| B7 | reconciliation unhealthy | initial gate table | missing | pass |
| B8 | scheduler invalid | initial gate table | missing | pass |
| B9 | autostart disabled | initial gate table | missing | pass |
| B10 | gate closed | initial gate table | missing | pass |
| B11 | LIVE unapproved | initial gate table | missing | pass |
| Order | activation → lane → kill → protection → reconcile → scheduler → autostart → gate → LIVE | every adjacent simultaneous-failure pair in initial gate table | partial | pass |
| Invariant | every refusal happens before issuer and official gateway | initial gate table asserts all counters zero | partial | pass |
| Success | exact snapshot reaches post-validation planning core | `TestPostValidationDispatchCorePlansOnceAndPersistsExactOfficialOutcomeWithSpies` | existing | pass; package-private seam only |
