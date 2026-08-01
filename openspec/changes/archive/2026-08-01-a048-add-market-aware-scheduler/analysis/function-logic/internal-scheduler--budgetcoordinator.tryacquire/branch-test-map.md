# Branch Test Map: `BudgetCoordinator.TryAcquire`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unknown future class refuses | `TestUnknownPollClassFailsClosed` | unknown class granted | yes |
| B2 | emergency/reconcile/fill/protection continue without provenance | `TestSafetyClassesContinueWithoutBudgetProvenance` | existing | yes |
| B3 | nil coordinator refuses low-priority poll | missing provenance test | existing | yes |
| B4 | conflicting equal-time provenance refuses low-priority poll | `TestEqualTimestampBudgetCorrectionWithConflictingProvenanceFailsClosed` | conflict masked | yes |
| B5 | endpoint/report/reset/observation missing refuses | missing/reset provenance tests | existing | yes |
| B6 | unknown future reset encoding refuses | `TestUnknownResetKindFailsClosed` | unknown reset granted | yes |
| B7 | clock before observation refuses | clock-skew test | existing | yes |
| B8 | elapsed reset refuses | stale provenance test | existing | yes |
| B9 | aged observed-at refuses | old observation test | existing | yes |
| B10 | invalid/overflow-edge numeric bounds fail closed while MaxInt remains safe | `TestBudgetArithmeticFailsClosedForInvalidBounds`, `TestSafetyReserveIsHalfRoundedUpWithFiveCallFloor` | ceil-half overflowed MaxInt | yes |
| B11 | outstanding and completed/unreconciled commitments cannot cross safety reserve | `TestCandidateEntryAndAnalyticsNeverSpendReservedBudget`, `TestCompleteNeverReopensCapacityWithoutAuthoritativeObservation`, outcome/stale tests | completion reopened stale capacity | yes |
| B12 | analytics enters reserved candidate/entry share check | analytics-share tests | existing | yes |
| B13 | analytics at half discretionary share refuses across refresh | `TestAnalyticsCannotConsumeCandidateAndEntryShare`, `TestAnalyticsCommitmentsSurviveNewObservationInSameResetWindow` | newer observation reset analytics counter | yes |
| B14 | entropy unavailable, generation exhausted, or absolute per-generation issue cap fails closed while safety remains allowed | entropy/generation tests and `TestCommitmentIssueCapIsAbsoluteAndResetScoped` | reported MaxInt permitted unbounded issued memory | yes |
| B15 | random capability read error permanently fails closed | `TestCommitmentEntropyFailureFailsClosedWithoutBlockingSafety` | partial/failed entropy could issue authority | yes |
| B16 | absent endpoint commitment map initializes defensively | focused budget tests | existing | yes |
| B17 | absent generation-issued set initializes defensively | focused budget tests | existing | yes |
| B18 | repeated capability is never reissued within one generation | `TestCommitmentCapabilityIsNeverReissuedWithinGeneration` | prior capability could regain authority | yes |
| success | capability has no exported fields and is coordinator/key/class/generation scoped | `TestCommitmentCapabilityIsOpaqueAndBoundToCoordinatorKeyClassAndGeneration` | forgeable sequential ID | yes |
