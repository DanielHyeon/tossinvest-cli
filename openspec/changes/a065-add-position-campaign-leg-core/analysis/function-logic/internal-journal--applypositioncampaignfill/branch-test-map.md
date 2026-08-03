# Branch Test Map: `ApplyPositionCampaignFill`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | duplicate/lower/restart | `TestCampaignFillRollbackAndRestartAreDeterministic` | yes | yes |
| B2 | predecessor late positive | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
| B3 | cap excess | `TestCampaignAggregateFillExcessLatchesReconcileWithoutTruncation` | yes | yes |
| B4 | legacy multiple matches preserves fill | `TestAmbiguousCampaignFillPreservesAuthoritativeFillTransaction` | yes | yes |
| B5 | CLOSED late fill | `TestClosedCampaignLateFillStaysClosedAndLatchesReconcile` | yes | yes |
| B6 | zero/partial terminal | `TestCampaignZeroAndPartialUnchangedTerminalObservationsCancelResidualIdempotently` | yes | yes |
| B7 | zero-fill all terminal closes/releases | `TestZeroFillTerminalClosesCampaignAndReleasesClaim` | yes | yes |
| B8 | live tx check | fill tests | yes | yes |
| B9 | authoritative query | scope tests | yes | yes |
| B10 | row scan | fill tests | yes | yes |
| B11 | row iteration error | query failure tests | yes | yes |
| B12 | no match | unrelated fill tests | yes | yes |
| B13 | multiple match | ambiguous-fill test | yes | yes |
| B14 | cumulative validation | fill validation tests | yes | yes |
| B15 | watermark compare | duplicate/lower tests | yes | yes |
| B16 | lower no-op | duplicate/lower tests | yes | yes |
| B17 | terminal change | terminal tests | yes | yes |
| B18 | duplicate no-op | restart tests | yes | yes |
| B19 | positive delta | partial tests | yes | yes |
| B20 | CLOSED/terminal reconcile | late-fill tests | yes | yes |
| B21 | AppliedFill delta mismatch | ambiguity tests | yes | yes |
| B22 | cap compare | cap test | yes | yes |
| B23 | first generation bind | first-fill tests | yes | yes |
| B24 | generation query | first-fill tests | yes | yes |
| B25 | expected successor | mismatch tests | yes | yes |
| B26 | claim bind | first-fill tests | yes | yes |
| B27 | generation mismatch | mismatch tests | yes | yes |
| B28 | existing generation check | restart tests | yes | yes |
| B29 | bound generation mismatch | reconcile tests | yes | yes |
| B30 | watermark update | fill tests | yes | yes |
| B31 | leg aggregate query | fill tests | yes | yes |
| B32 | residual calculation | fill tests | yes | yes |
| B33 | full leg | full-fill tests | yes | yes |
| B34 | terminal residual | terminal tests | yes | yes |
| B35 | successor query | replacement tests | yes | yes |
| B36 | zero-fill cancel | terminal tests | yes | yes |
| B37 | cap excess latch | cap test | yes | yes |
| B38 | terminal state preservation | late-fill test | yes | yes |
| B39 | leg update | fill tests | yes | yes |
| B40 | current/successor remaining update from each order cap | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` + restart test | yes | yes |
| B41 | CLOSED latch | CLOSED late-fill test | yes | yes |
| B42 | reconcile latch | ambiguity tests | yes | yes |
| B43 | PLANNED activation | first-fill tests | yes | yes |
| B44 | campaign CAS update | race tests | yes | yes |
| B45 | all-terminal query | zero-fill close test | yes | yes |
| B46 | campaign close | zero-fill close test | yes | yes |
| B47 | claim release | zero-fill close test | yes | yes |
| B48 | event append | replay tests | yes | yes |
| B49 | digest append | drift tests | yes | yes |
| B50 | ambiguous retry | ambiguous-fill retry test | yes | yes |
| B51 | durable account reconcile | CLOSED/ambiguous tests | yes | yes |
| B52 | authoritative fill preserved | ambiguity test | yes | yes |
| B53 | final nil | all fill tests | yes | yes |
| B54 | current per-order remaining calculation failure | fail-closed tx rollback contract + watermark/restart suite | yes | yes |
