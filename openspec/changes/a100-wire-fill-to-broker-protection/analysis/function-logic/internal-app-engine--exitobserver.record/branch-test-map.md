# Branch Test Map: `ExitObserver.record`

Source: `internal/app/engine/exitloop.go` (1232-1362).

| Branch | Scenario | Test |
|---|---|---|
| B1 | invalid quote no-op | `TestA111InvalidQuoteNeverRefreshesAnEvaluatedSnapshot` |
| B2 | missing source fallback | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` |
| B3 | missing fetched time cycle fallback | `TestStableObservationIDReusesOneFallbackWithinCycle` |
| B4 | fetched time retained | `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` |
| B5 | clear before arm | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B6 | rejudge take-profit withheld | `TestA111SupersededRejudgementStillReleasesWithoutNonprotectiveOrderSideEffects` |
| B7 | normal clear path | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B8 | clearing error | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` |
| B9 | uncleared delay | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` |
| B10 | delay clears | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` |
| B11 | proposal constructed | `TestABaselineBreachProposesTheWholePosition` |
| B12 | fresh intent ID | `TestTheWholeExitPathEndToEnd` |
| B13 | record error | `TestAnUnresolvedProposalSuppressesTheNextOne` |
| B14 | pending no-op | `TestAnUnresolvedProposalSuppressesTheNextOne` |
| B15 | quarantine announcement | `TestUnknownLegacyPolicyIdentityIsDurablyGenerationQuarantined` |
| B16 | unarmed no submit | `TestARefusedProposalReleasesTheLevelAndAlerts` |

> **a100 R0 재동결(2026-10-04) — 좌표 이동.** 옛 · 새 AST 의 분기 (id · 종류) 목록이 같아 ast.json 을 현재 소스 추출로 바꾸고 `(start-end)` 범위 · 파일 SHA-256 만 옮겼다(`analysis/harness/r0_refresh_bundle.py shift`). 산문 · 시험 인용은 손대지 않았다. 옛 ast 는 returns/calls 를 싣지 않던 추출기 판본이라 그 둘은 대조 대상이 아니다.
