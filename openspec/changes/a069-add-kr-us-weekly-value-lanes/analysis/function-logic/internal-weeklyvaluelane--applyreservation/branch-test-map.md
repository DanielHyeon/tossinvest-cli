# Branch Test Map: `applyReservationTransition`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | KR/US and campaign versions are independent | TestReservationScopesVersionCountAndOrdinals | yes | yes |
| B2 | ordinals consume distinctly and sequentially | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B3 | same-week correction does not add slot | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B4 | zero release does not advance ordinal | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B5 | trusted evaluatedAt rejects stale calendar | TestReservationRejectsStaleCalendarAtCommandEvaluation | yes | yes |
| B6 | idempotency receipt match/conflict | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B7 | scope version CAS mismatch | TestReservationScopesVersionCountAndOrdinals | yes | yes |
| B8 | reserve calendar refusal | TestReservationRejectsStaleCalendarAtCommandEvaluation | yes | yes |
| B9 | seven-leg exhaustion | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B10 | non-next or concurrent active reserve | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B11 | duplicate key or reservation ID | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B12 | positive fill authority required | TestPositiveFillCannotBypassAtomicRiskTransition | yes | yes |
| B13 | missing positive-fill reservation | TestApplyPositiveFillAtomicCommitsReservationAndRiskTogether | yes | yes |
| B14 | terminal positive-fill reservation | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B15 | fill identity/quantity/ordinal conflict | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B16 | fill consumes next distinct ordinal | TestPositiveFillsRequireDistinctSequentialOrdinals | yes | yes |
| B17 | missing zero-release reservation | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B18 | terminal zero-release reservation | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B19 | non-authoritative/nonzero release refusal | TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease | existing | existing |
| B20 | unknown action refusal | TestPositiveFillCannotBypassAtomicRiskTransition | yes | yes |
