# Branch Test Map: `RiskGuardian.IssueEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | missing collector refuses with zero authority | Guardian refusal tests | baseline | baseline |
| B2 | scoped intent validation fails | scoped intent tests | baseline | baseline |
| B3 | non-BUY intent cannot use entry issuance | reduction/entry separation tests | baseline | baseline |
| B4 | chain refusal is observed and only enumerated losses escalate | `TestReachingTheDailyLossLimitTightensTheOperatingMode`, `TestOtherChainRefusalsAreNotTriggers` | baseline | baseline |
| B5 | exposure arithmetic refusal is observed | risk/Guardian arithmetic tests | baseline | baseline |
| B6 | snapshot collection failure aborts transaction | issuance/recollection tests | baseline | baseline |
| B7 | mixed/invalid exposure usage aborts transaction | mixed-currency reservation tests | baseline | baseline |
| B8 | stale/concurrent/recollection failure leaves no spendable decision | reservation, issuance, and `TestConcurrentIssuancesCannotExceedTheAggregateLimit` | baseline | baseline |
| B9 | typed issuance-stage refusal receives the ledger reason code | entry observation/issuance refusal tests | baseline | baseline |
| Success | decision and reservation commit together | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether`, `TestTheGuardiansOwnIssuanceSubmits` | baseline | baseline |
| A047-1 | ordinary caller cannot attach a strategy plan | compile/API surface inspection | public draft existed | pass |
| A047-2 | strategy issuance commits decision/reservation/lineage/start together | journal strategy atomic/rollback tests | split draft | pass |
| A047-3 | strategy receipt returns canonical quantity/client id | strategy adapter integration | post-commit lookup draft | pass |
| B10 | collection error propagates without partial issuance | issuance/recollection tests | baseline | pass |
| B11 | private strategy plan selects atomic v14 path | production strategy issuance tests | split draft | pass |
| B12 | strategy recollection success returns exact receipt | production atomic success | missing | pass |
| B13 | post-commit issued observation cannot roll back authority | observation/Guardian suite | baseline | pass |
